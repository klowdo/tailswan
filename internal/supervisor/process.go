package supervisor

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

type Process struct {
	name    string
	cmd     *exec.Cmd
	mu      sync.Mutex
	started time.Time
}

func (p *Process) Start(name string, args ...string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.name = name
	p.cmd = exec.Command(name, args...)
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", name, err)
	}

	p.started = time.Now()
	slog.Info("Started process", "name", name, "pid", p.cmd.Process.Pid)
	return nil
}

func (p *Process) Wait() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Wait()
}

func (p *Process) Kill() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}

	slog.Info("Stopping process", "name", p.name, "pid", p.cmd.Process.Pid)
	return p.cmd.Process.Signal(syscall.SIGTERM)
}

func (p *Process) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}

	err := p.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}
