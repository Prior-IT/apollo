package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/prior-it/apollo/config"
)

type command struct {
	Cmd             *exec.Cmd
	Out, Err        chan string
	prOut, prErr    chan string
	Name            string
	Style           lipgloss.Style
	HasError        bool
	RequiresSigKill bool
	RunAtStart      bool
	FullCommand     string
	restartMutex    sync.Mutex
}

// Create a new command but do not run it yet.
// Call cmd.Run to start the command in a separate goroutine.
func newCommand(
	ctx context.Context,
	name string,
	style lipgloss.Style,
	requiresSigKill bool,
	runAtStart bool,
	cmd string,
	args ...any,
) (*command, error) {
	var c *exec.Cmd
	fullCommand := fmt.Sprintf(cmd, args...)
	// Watchers need to run without context so we will always kill their entire process group
	if requiresSigKill {
		c = exec.Command("/bin/sh", "-c", fullCommand)
	} else {
		c = exec.CommandContext(ctx, "/bin/sh", "-c", fullCommand)
	}

	// Child processes should get their own pgid
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to capture stderr: %w", err)
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to capture stdout: %w", err)
	}

	// Output and error channels
	outCh := make(chan string)
	errCh := make(chan string)

	// Intermediary channels that will get the actual command output
	prOutCh := make(chan string)
	prErrCh := make(chan string)

	// Stream command output to the intermediary channels
	go streamOutput(stdout, prOutCh)
	go streamOutput(stderr, prErrCh)

	// Redirect streams so we can actually use out vs err as intended
	go redirectStreams(name, prOutCh, prErrCh, outCh, errCh)

	return &command{
		Cmd:             c,
		Name:            name,
		FullCommand:     fullCommand,
		Style:           style,
		RequiresSigKill: requiresSigKill,
		RunAtStart:      runAtStart,
		Out:             outCh,
		Err:             errCh,
		prOut:           prOutCh,
		prErr:           prErrCh,
	}, nil
}

// Run the command and wait for it to finish in a separate goroutine.
// This will increment the waitgroup while the command is running and decrement it after.
// All command errors will be sent to the errChan, while start-up errors will simply be returned.
// OutChan will be used for sending runner-output, e.g. after the command has possibly already
// closed.
func (c *command) Run(
	wg *sync.WaitGroup,
	errChan chan<- error,
) error {
	if err := c.Cmd.Start(); err != nil {
		return err
	}
	c.Out <- fmt.Sprintf("running %q", c.FullCommand)
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Wait for the subprocess to finish
		if err := c.Cmd.Wait(); err != nil {
			e, ok := err.(*exec.ExitError)
			if ok && e.ExitCode() > 0 {
				// If it is an ExitError, check the exit code and pass along actual errors
				errChan <- fmt.Errorf("command %q exited with exit code %d, error: %w", c.Name, e.ExitCode(), err)
			} else if !ok {
				// If it is a different error, pass the error along
				errChan <- fmt.Errorf("command %q exited with error: %w", c.Name, err)
			}
		}
		newCommandDoneEvent(c.Name)
	}()
	return nil
}

// Returns true if the command has exited, and false otherwise.
func (c *command) Done() bool {
	return c.Cmd.ProcessState != nil
}

// Returns true if the command has started and not finished, and false otherwise.
func (c *command) IsRunning() bool {
	return c.Cmd.Process != nil && c.Cmd.ProcessState == nil
}

const restartKillAttempts = 5

// SigTermPGID will send SIGTERM to the entire process group for this command.
func (c *command) SigTermPGID() {
	if c.RequiresSigKill {
		return
	}

	if debug {
		c.Out <- "Will send SIGTERM to the entire process group"
	}

	// Get the process group ID
	pgid, err := syscall.Getpgid(c.Cmd.Process.Pid)
	// ESRCH = process already terminated
	if err != nil && err != syscall.ESRCH {
		c.Err <- fmt.Sprintf("Failed to get process group ID: %v (%#v)", err, err)
	}

	// Send SIGTERM to the entire process group
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil && err != syscall.ESRCH {
		c.Err <- fmt.Sprintf("Error while sending SIGTERM to process group: %v (%#v)", err, err)
	}
}

// WaitUntilKilled repeatedly tries to kill this command. If this did not work after
// maxNumberOfAttempts tries, this will return an error. If the process did die at some point, this will return
// nil.
func (c *command) WaitUntilKilled(
	maxNumberOfAttempts int,
	timeout time.Duration,
	errChan chan<- error,
	outChan chan<- string,
) error {
	attempts := 0
	for c.IsRunning() {
		c.Kill(errChan, outChan)
		if c.IsRunning() {
			if debug {
				c.Out <- "Command is still running, waiting for it to die so it can restart"
			}
			time.Sleep(timeout * time.Second)
			attempts++
			if attempts > maxNumberOfAttempts {
				return fmt.Errorf(
					"could not kill existing command %q, restart cancelled",
					c.Name,
				)
			}
		}
	}
	return nil
}

// Restart a command. If the command is done, this will immediately relaunch it. If not, it will
// kill the existing process and wait until it exits. This is done on the current thread, callers beware.
func (c *command) Restart(
	ctx context.Context,
	cfg *config.Config,
	wg *sync.WaitGroup,
	errChan chan<- error,
	outChan chan<- string,
) error {
	c.restartMutex.Lock()
	defer c.restartMutex.Unlock()
	if c.IsRunning() {
		if debug {
			c.Out <- "Restart requested but command still running"
		}

		// Sending SIGTERM to a group does not always seem to kill the subprocesses correctly for
		// some commands, although this only happens very infrequently or on specific OSes. In order
		// to favour the application always restarting correctly, we just immediately sigkill those
		// processes.
		c.SigTermPGID()

		time.Sleep(time.Duration(cfg.App.ShutdownTimeout) * time.Second)

		if !c.RequiresSigKill && c.IsRunning() {
			c.Out <- "Command seemingly did not respond to SIGTERM, sending SIGKILL now"
		}

		err := c.WaitUntilKilled(
			restartKillAttempts,
			time.Duration(cfg.App.ShutdownTimeout+1),
			errChan,
			outChan,
		)
		if err != nil {
			return err
		}
	} else {
		if debug {
			c.Out <- "Restart requested"
		}
	}

	newCmd, err := newCommand(
		ctx,
		c.Name,
		c.Style,
		c.RequiresSigKill,
		c.RunAtStart,
		c.FullCommand, //nolint:govet // we can be 100% sure this string has already been formatted correctly
	)
	if err != nil {
		return fmt.Errorf("cannot restart command %q: %w", c.Name, err)
	}

	go PipeChan(newCmd.Out, c.Out)
	go PipeChan(newCmd.Err, c.Err)

	// The old ones have been closed whenever the process died so we can just reassign
	c.prOut = newCmd.Out
	c.prErr = newCmd.Err

	err = newCmd.Run(wg, errChan)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("cannot run restarted %q command: %w", c.Name, err)
	}

	c.Cmd = newCmd.Cmd
	c.HasError = false

	return nil
}

// Forcefully kill the process and its subprocesses if they're still running.
// Only try this if SIGINT and SIGTERM did not work.
//
//nolint:cyclop
func (c *command) Kill(errChan chan<- error, outChan chan<- string) {
	if c.Cmd.ProcessState != nil && c.Cmd.ProcessState.Exited() && !c.RequiresSigKill {
		return
	}

	if c.Cmd.Process == nil {
		// Process never even started
		return
	}

	outChan <- fmt.Sprintf("Force-killing %s...", c.Name)

	// Get the process group ID
	pgid, err := syscall.Getpgid(c.Cmd.Process.Pid)
	// ESRCH = process already terminated
	if err != nil && err != syscall.ESRCH {
		errChan <- fmt.Errorf("failed to get process group ID for %q: %v (%#v)", c.Name, err, err)
	}

	// Send SIGKILL to the entire process group
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		errChan <- fmt.Errorf("failed to kill process group for %q: %v", c.Name, err)
	}

	if err := c.Cmd.Process.Kill(); err != nil && err != os.ErrProcessDone {
		errChan <- fmt.Errorf("cannot kill %q: %v", c.Name, err)
	} else {
		outChan <- fmt.Sprintf("%v has been killed", c.Name)
	}
}

// streamOutput reads from the provided pipe and sends each line to the output channel.
func streamOutput(pipe io.ReadCloser, outputChan chan<- string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		outputChan <- scanner.Text()
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, os.ErrClosed) {
		fmt.Fprintf(os.Stderr, "Error reading output: %v\n", err)
	}
}

func PipeChan(in chan string, out chan<- string) {
	for msg := range in {
		out <- msg
	}
}

// Since a lot of programs seem to misuse stdout and stderr, we need to manually filter and redirect
// the streams to get access to actual errors. For progams that are not supported by this, the
// streams won't be redirected and just passed through from the program channels to the respective
// runner channels.
//
//nolint:cyclop
func redirectStreams(
	name string,
	programOut, programErr chan string,
	runnerOut, runnerErr chan<- string,
) {
	switch name {
	case "templ":
		for out := range programErr {
			if strings.Contains(out, "(✗)") {
				runnerErr <- out
			} else {
				runnerOut <- out
			}
		}
	case "application":
		go PipeChan(programErr, runnerErr)
		for out := range programOut {
			if strings.Contains(out, "ERROR") ||
				strings.Contains(out, "Failed to open tcp listener") {
				runnerErr <- out
			} else {
				runnerOut <- out
			}
		}
	case "tailwind":
		for {
			select {
			case out, ok := <-programErr:
				if !ok {
					return
				}
				runnerOut <- out
			case out, ok := <-programOut:
				if !ok {
					return
				}
				runnerOut <- out
			}
		}
	default:
		go PipeChan(programOut, runnerOut)
		go PipeChan(programErr, runnerErr)
	}
}
