package shell

import (
	"bytes"
	"os/exec"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/metrics"
)

type Supervisor interface {
	logger.Loggable

	Cmd() *exec.Cmd
	Start() error
	Run() error
	RunAndGetOutput() ([]byte, []byte, error)
	Wait() error
	Kill(signal syscall.Signal) error
}

func (w *wrapper) Cmd() *exec.Cmd {
	return w.cmd
}

func (w *wrapper) Start() error {
	return w.cmd.Start()
}

func (w *wrapper) Run() error {
	if err := w.Start(); err != nil {
		return err
	}
	return w.Wait()
}

func (w *wrapper) RunAndGetOutput() ([]byte, []byte, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := w.cmd
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := w.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

func (w *wrapper) Wait() (err error) {
	defer metrics.ShellCommandRun.StartTimer(w.ctx).SuccessFinish(&err)

	cmd := w.cmd

	if w.ctx == nil || !w.ctxCancellation {
		return cmd.Wait()
	}

	done := make(chan error, 1)

	go func() {
		err := cmd.Wait()
		if atomic.LoadUint32(&w.killedByCancel) > 0 {
			// Normally, the cmd would report it was "terminated".
			// In reality, the context was canceled and we gracefully stopped it, report as such.
			err = w.ctx.Err()
		}
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-w.ctx.Done():
		log(w.ctx, nil).Info("command was canceled; attempting graceful termination")

		atomic.AddUint32(&w.killedByCancel, 1)
		err := w.Kill(syscall.SIGTERM)
		if err != nil {
			log(w.ctx, err).Warn("unable to sigterm process; attempting killing")
			atomic.AddUint32(&w.killedByCancel, 1)
			if err := w.Kill(syscall.SIGKILL); err != nil {
				return errors.Wrap(err, "unable to kill process")
			}

			return <-done
		}

		select {
		case err := <-done:
			return err
		case <-time.After(w.gracefulTerminationOnCancelTimeout):
			log(w.ctx, nil).Warn("command was canceled and did not terminate in time; killing")
			atomic.AddUint32(&w.killedByCancel, 1)
			if err := w.Kill(syscall.SIGKILL); err != nil {
				return errors.Wrap(err, "unable to kill process")
			}
			return <-done
		}
	}
}

func (w *wrapper) Kill(signal syscall.Signal) error {
	pid := w.cmd.Process.Pid
	if w.sysProcAttr != nil && w.sysProcAttr.Setpgid {
		pid = -pid
	}
	log(w.ctx, nil).
		WithField("pid", pid).
		WithField("signal", signal.String()).
		Debug("sending signal")

	return syscall.Kill(pid, signal)
}
