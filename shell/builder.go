package shell

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

type Builder interface {
	WithOSEnv() Builder
	WithEnv(env Env) Builder
	WithSysProcAttr(attr *syscall.SysProcAttr) Builder
	WithContextCancellation(gracefulTimeout time.Duration) Builder
	Prepare() Supervisor
}

func (w *wrapper) WithOSEnv() Builder {
	w.osEnv = true
	return w
}

func (w *wrapper) WithEnv(env Env) Builder {
	if w.env == nil {
		w.env = []string{}
	}
	for key, val := range env {
		w.env = append(w.env, key+"="+val)
	}
	return w
}

func (w *wrapper) WithSysProcAttr(attr *syscall.SysProcAttr) Builder {
	if w.sysProcAttr == nil {
		w.sysProcAttr = attr
	} else if err := mergo.Merge(w.sysProcAttr, *attr); err != nil {
		panic(errors.Wrap(err, "unable to merge SysProcAttr"))
	}
	return w
}

func (w *wrapper) WithContextCancellation(gracefulTimeout time.Duration) Builder {
	w.ctxCancellation = true
	w.gracefulTerminationOnCancelTimeout = gracefulTimeout
	return w
}

func (w *wrapper) Prepare() Supervisor {
	log(w.ctx, nil).Debug("preparing shell command")

	cmd := exec.Command(w.path, w.args...)
	cmd.SysProcAttr = w.sysProcAttr

	// Avoid modifying the w.env, so it doesn't log the sensitive OS environment
	env := w.env
	if w.osEnv {
		env = append(env, os.Environ()...)
	}
	cmd.Env = env

	w.cmd = cmd

	return w
}
