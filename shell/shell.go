package shell

import (
	"context"
	"os/exec"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/statsd"
)

var log = logger.New("shell")

const defaultGracefulTerminationOnCancelPeriod = 3 * time.Second

type Env map[string]string

func New(ctx context.Context, path string, args ...string) Supervisor {
	return NewDefaultBuilder(ctx, path, args...).Prepare()
}

func NewDefaultBuilder(ctx context.Context, path string, args ...string) Builder {
	return NewBuilder(ctx, path, args...).
		WithOSEnv().
		WithSysProcAttr(&syscall.SysProcAttr{Setpgid: true}).
		WithContextCancellation(defaultGracefulTerminationOnCancelPeriod)
}

func NewBuilder(ctx context.Context, path string, args ...string) Builder {
	w := &wrapper{
		path: path,
		args: args,
	}
	w.ctx = logger.WithLoggable(ctx, w)
	return w
}

type wrapper struct {
	args        []string
	env         []string
	ctx         context.Context
	path        string
	dir         string
	cmd         *exec.Cmd
	sysProcAttr *syscall.SysProcAttr

	// When a context is provided and it is canceled while the process is
	// running, we send SIGTERM to the process. if, after this period, the
	// process is still running, we send SIGKILL. If left unspecified, the
	// default is 3 seconds.
	gracefulTerminationOnCancelTimeout time.Duration

	// Number of times the command was "killed" after the context was canceled.
	killedByCancel uint32

	osEnv           bool
	ctxCancellation bool
}

func (w *wrapper) StatsTags() statsd.Tags {
	return statsd.Tags{
		// Only report the command since the rest may contain sensitive information.
		"cmdPath": w.path,
	}
}

func (w *wrapper) LogFields() logrus.Fields {
	return logrus.Fields{
		"cmdPath": w.path,
		"cmdArgs": w.args,
		"cmdEnv":  w.env,
	}
}
