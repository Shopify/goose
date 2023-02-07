package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleSupervisor_Wait() {
	ctx := context.Background()
	cmd := New(ctx, "cat")
	stdin, _ := cmd.Cmd().StdinPipe()
	stdout, _ := cmd.Cmd().StdoutPipe()

	cmd.Start()
	stdin.Write([]byte("foo"))
	stdin.Close()

	output, _ := io.ReadAll(stdout)

	// Must call Wait _after_ interacting with pipes
	cmd.Wait()

	fmt.Println(string(output))
	// Output:
	// foo
}

func ExampleSupervisor_RunAndGetOutput() {
	ctx := context.Background()
	cmd := New(ctx, "echo", "-n", "foo")
	stdout, stderr, err := cmd.RunAndGetOutput()

	fmt.Println(string(stdout))
	fmt.Println(string(stderr))
	fmt.Println(err)
	// Output:
	// foo
	//
	// <nil>
}

func ExampleBuilder_WithEnv() {
	ctx := context.Background()
	stdout, _, _ := NewBuilder(ctx, "bash", "-c", "echo -n $foo").
		WithEnv(Env{"foo": "bar"}).
		Prepare().
		RunAndGetOutput()

	fmt.Println(string(stdout))
	// Output:
	// bar
}

func TestCommandRunWait(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "bash", "-c", "echo -n foo; echo -n bar >&2")
	c := cmd.Cmd()
	stdout, stderr, err := cmd.RunAndGetOutput()

	assert.NoError(t, err)
	assert.True(t, c.ProcessState.Success())
	assert.True(t, c.ProcessState.Exited())
	assert.Equal(t, []byte("foo"), stdout)
	assert.Equal(t, []byte("bar"), stderr)
}

type syncBuffer struct {
	L sync.Mutex
	b bytes.Buffer
}

func (b *syncBuffer) Len() int {
	b.L.Lock()
	defer b.L.Unlock()
	return b.b.Len()
}

func (b *syncBuffer) String() string {
	b.L.Lock()
	defer b.L.Unlock()
	return b.b.String()
}

func (b *syncBuffer) Write(p []byte) (n int, err error) {
	b.L.Lock()
	defer b.L.Unlock()
	return b.b.Write(p)
}

func TestCommandRunNoWait(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "bash", "-c", "sleep 0.05; echo -n foo")
	stdout := &syncBuffer{}
	c := cmd.Cmd()
	c.Stdout = stdout

	err := cmd.Start()
	assert.NoError(t, err)

	assert.Nil(t, c.ProcessState)        // No state yet
	assert.NotEqual(t, 0, c.Process.Pid) // But the process exists

	// No output yet.
	// Invoking with Run(false) and a Stdout buffer is not really useful.
	// Consider using the StdoutPipe instead.
	assert.Equal(t, 0, stdout.Len())

	// One could manually call Wait(), but might as well use Run(true)
	err = cmd.Wait()
	assert.NoError(t, err)
	assert.NotNil(t, c.ProcessState)
	assert.True(t, c.ProcessState.Success())
	assert.True(t, c.ProcessState.Exited())

	assert.NoError(t, err)
	assert.Equal(t, "foo", stdout.String())
}

func TestCommandRunPipe(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "bash", "-c", "sleep 0.05; echo -n foo")
	c := cmd.Cmd()

	pipe, err := c.StdoutPipe()
	assert.NoError(t, err)

	err = cmd.Start()
	assert.NoError(t, err)

	// Still no state, even though we have an output, because the command was a fire-and-forget.
	assert.Nil(t, c.ProcessState)        // No state yet
	assert.NotEqual(t, 0, c.Process.Pid) // But the process exists

	// Calling ReadAll will wait for the pipe to close, so all the output is there.
	output, err := io.ReadAll(pipe)
	assert.NoError(t, err)
	assert.Equal(t, []byte("foo"), output)

	assert.Nil(t, c.ProcessState) // Still no state yet

	err = cmd.Wait()
	assert.NoError(t, err)

	// The command completed, and we waited it, so we have a processstate.
	assert.NotNil(t, c.ProcessState)
	assert.True(t, c.ProcessState.Exited())
}

func TestCommandRunWaitPipeFails(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "bash", "-c", "sleep 0.05; echo -n foo")
	c := cmd.Cmd()

	pipe, err := c.StdoutPipe()
	assert.NoError(t, err)

	err = cmd.Run()
	assert.NoError(t, err)

	// The command completed, and we waited it, so we have a processstate.
	assert.NotNil(t, c.ProcessState)
	assert.True(t, c.ProcessState.Exited())

	// Calling ReadAll will wait for the pipe to close, so all the output is there.
	_, err = io.ReadAll(pipe)
	assert.Error(t, err, "read |0: file already closed")
}

func TestCommandWithWorkingDir(t *testing.T) {
	ctx := context.Background()

	tmpdir, err := filepath.EvalSymlinks(os.TempDir())
	assert.NoError(t, err)

	stdout, _, err := NewBuilder(ctx, "pwd").
		WithWorkingDir(tmpdir).
		Prepare().
		RunAndGetOutput()
	assert.NoError(t, err)

	expected, err := filepath.EvalSymlinks(tmpdir)
	assert.NoError(t, err)

	assert.Equal(t, expected, strings.TrimSuffix(string(stdout), "\n"))
}

func TestCommandEnvDoesNotEval(t *testing.T) {
	ctx := context.Background()
	stdout, _, err := NewBuilder(ctx, "echo", "-n", "$foo").
		WithEnv(Env{"foo": "bar"}).
		Prepare().
		RunAndGetOutput()

	assert.NoError(t, err)
	// Not expanded
	assert.Equal(t, []byte("$foo"), stdout)
}

func TestCommandContextCancels(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := NewBuilder(ctx, "bash", "-c", "read; echo -n foo").
		WithContextCancellation(100 * time.Millisecond).
		Prepare()

	c := cmd.Cmd()
	stdout := &bytes.Buffer{}
	c.Stdout = stdout

	err := cmd.Start()
	assert.NoError(t, err)
	cancel()

	err = cmd.Wait()
	assert.Equal(t, context.Canceled, err)
	ws := cmd.Cmd().ProcessState.Sys().(syscall.WaitStatus)
	sig := ws.Signal()
	assert.Equal(t, "terminated", sig.String())

	assert.Equal(t, 0, stdout.Len())
}

func TestCommandForceKills(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := NewBuilder(ctx, "bash", "-c", "trap '' TERM; read; echo -n foo").
		WithContextCancellation(100 * time.Millisecond).
		Prepare()

	c := cmd.Cmd()
	_, err := c.StdinPipe()
	assert.NoError(t, err)

	stdout := &bytes.Buffer{}
	c.Stdout = stdout

	err = cmd.Start()
	assert.NoError(t, err)

	// give bash time to boot and register the SIGTERM handler
	time.Sleep(100 * time.Millisecond)

	cancel()

	err = cmd.Wait()
	assert.Equal(t, context.Canceled, err)
	ws := cmd.Cmd().ProcessState.Sys().(syscall.WaitStatus)
	sig := ws.Signal()
	assert.Equal(t, "killed", sig.String())

	assert.Equal(t, 0, stdout.Len())
}

func TestCommandStdinBeforeStart(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "bash", "-c", "cat; echo -n foo")

	c := cmd.Cmd()
	stdin, err := c.StdinPipe()
	assert.NoError(t, err)

	_, err = stdin.Write([]byte("bar"))
	assert.NoError(t, err)

	err = stdin.Close()
	assert.NoError(t, err)

	stdout, _, err := cmd.RunAndGetOutput()
	assert.NoError(t, err)

	assert.Equal(t, "barfoo", string(stdout))
}

func TestCommandStdinAfterStart(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "bash", "-c", "cat; echo -n foo")

	c := cmd.Cmd()
	stdin, err := c.StdinPipe()
	assert.NoError(t, err)

	stdout := &bytes.Buffer{}
	c.Stdout = stdout

	err = cmd.Start()
	assert.NoError(t, err)

	_, err = stdin.Write([]byte("bar"))
	assert.NoError(t, err)

	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()
	time.Sleep(100 * time.Millisecond)

	select {
	case <-done:
		assert.Fail(t, "should not be done")
	default:
	}

	err = stdin.Close()
	assert.NoError(t, err)

	err = <-done
	assert.NoError(t, err)

	assert.Equal(t, "barfoo", stdout.String())
}

// Verify we can stream while command is running
func TestCommandStdoutPipe(t *testing.T) {
	ctx := context.Background()
	cmd := New(ctx, "cat")
	c := cmd.Cmd()

	stdin, err := c.StdinPipe()
	assert.NoError(t, err)

	stdout, err := c.StdoutPipe()
	assert.NoError(t, err)

	output := make(chan []byte)
	read := make(chan struct{})
	chanErr := make(chan error)
	go func() {
		for {
			<-read
			b := make([]byte, 10)
			n, err := stdout.Read(b)
			output <- b[:n]
			chanErr <- err
			if err == io.EOF {
				return
			}
		}
	}()

	err = cmd.Start()
	assert.NoError(t, err)

	read <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	select {
	case b := <-output:
		assert.NoError(t, <-chanErr)
		assert.Empty(t, b, "should not have received output")
	default:
	}

	_, err = stdin.Write([]byte("foo"))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	select {
	case b := <-output:
		assert.Equal(t, "foo", string(b))
		assert.NoError(t, <-chanErr)
	default:
		assert.Fail(t, "should have output")
	}

	_, err = stdin.Write([]byte("bar"))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	_, err = stdin.Write([]byte("baz"))
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)

	read <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	select {
	case b := <-output:
		assert.Equal(t, "barbaz", string(b))
		assert.NoError(t, <-chanErr)
	default:
		assert.Fail(t, "should have output")
	}

	err = stdin.Close()
	assert.NoError(t, err)

	read <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	err = cmd.Wait()
	assert.NoError(t, err)
}
