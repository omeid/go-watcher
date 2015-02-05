package watcher

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
)

type Builder struct {
	runner  *Runner
	watcher *Watcher
}

func NewBuilder(w *Watcher, r *Runner) *Builder {
	return &Builder{watcher: w, runner: r}
}

// Build listens watch events from Watcher and sends messages to Runner
// when new changes are built.
func (b *Builder) Build(p *Params) error {
	go b.registerSignalHandler()
	go func() {
		b.watcher.update <- true
	}()

	for <-b.watcher.Wait() {
		fileName := p.createBinaryName()

		pkg := p.GetPackage()

		log.Printf(color.CyanString("Building %s...\n", pkg))

		// build package
		cmd, err := runCommand("go", "build", "-o", fileName, pkg)
		if err != nil {
			return fmt.Errorf("Could not run 'go build' command: %s", err)
		}

		if err := cmd.Wait(); err != nil {
			if err := interpretError(err); err != nil {
				log.Printf(color.RedString("An error occurred while building: %s", err))
			} else {
				log.Printf(color.RedString("A build error occurred. Please update your code..."))
			}

			continue
		}

		// and start the new process
		b.runner.restart(fileName)
	}

	return nil
}

func (b *Builder) registerSignalHandler() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals)
		for {
			signal := <-signals
			switch signal {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSTOP:
				b.watcher.Close()
				b.runner.Close()
			}
		}
	}()
}

// interpretError checks the error, and returns nil if it is
// an exit code 2 error. Otherwise error is returned as it is.
// when a compilation error occurres, it returns with code 2.
func interpretError(err error) error {
	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return err
	}

	status, ok := exiterr.Sys().(syscall.WaitStatus)
	if !ok {
		return err
	}

	if status.ExitStatus() == 2 {
		return nil
	}

	return err
}
