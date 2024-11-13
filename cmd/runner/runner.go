package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bep/debounce"

	"github.com/charmbracelet/lipgloss"
	"github.com/prior-it/apollo/config"
)

// Launch a new command that notifies the templ proxy of any changes. The new command will be added
// to the existing waitgroup and its output will be automatically redirected to the channels of the specified command.
// This will return all errors that occur during startup but output runtime errors to errChan.
func notifyTempl(
	ctx context.Context,
	cmd *command,
	cfg *config.Config,
	wg *sync.WaitGroup,
	errChan chan<- error,
) error {
	notify, err := newCommand(
		ctx,
		"notify",
		cmd.Style,
		false,
		false,
		"go run github.com/a-h/templ/cmd/templ@%v generate --notify-proxy",
		cfg.Tools.Templ,
	)
	if err != nil {
		return fmt.Errorf("cannot create templ notify: %w", err)
	}

	cmd.Out <- fmt.Sprintf("Notifying templ of %s change", cmd.Name)
	go PipeChan(notify.Out, cmd.Out)
	go PipeChan(notify.Err, cmd.Err)

	err = notify.Run(wg, errChan)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("cannot notify templ after %v changes: %w", cmd.Name, err)
	}
	return nil
}

//nolint:cyclop
func createCommands(
	ctx context.Context,
	errChan chan<- error,
	outChan chan<- string,
	cfg *config.Config,
) ([]*command, *sync.WaitGroup, error) {
	go func() {
		err := runFileWatcher(ctx, cfg, outChan)
		if err != nil {
			log.Fatalf("cannot create file watcher: %v\n", err)
		}
	}()

	var templ, tailwind, sqlc *command
	var err error
	var wg sync.WaitGroup

	if len(cfg.Tools.Templ) > 0 {
		// Start the templ proxy server
		templ, err = newCommand(
			ctx,
			"templ",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#CC8925")),
			true,
			true,
			"go run github.com/a-h/templ/cmd/templ@%v generate --watch --proxy=\"http://127.0.0.1:%v\" --open-browser=%v",
			cfg.Tools.Templ,
			cfg.App.Port,
			cfg.Tools.OpenBrowser,
		)
		if err != nil {
			return []*command{}, nil, fmt.Errorf("Cannot create templ: %w", err)
		}

		// Notify templ proxy whenever the tailwind output file changes
		debouncer := debounce.New(time.Duration(cfg.Tools.Debounce) * time.Millisecond)
		go onEvent(evFileChanged, func(ev event) bool {
			if strings.HasSuffix(ev.payload, cfg.Tools.TailwindOutput) {
				debouncer(func() {
					templ.Out <- "Tailwind file change detected, notifying proxy"
					err := notifyTempl(ctx, tailwind, cfg, &wg, errChan)
					if err != nil {
						errChan <- err
					}
				})
			}
			return true
		})

	} else if debug {
		outChan <- "Templ disabled in config"
	}

	if len(cfg.Tools.Tailwind) > 0 {

		// Compile tailwind output
		tailwind, err = newCommand(
			ctx,
			"tailwind",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#7DD3FC")),
			false,
			true,
			"npx tailwindcss@%v -i %v -o %v",
			cfg.Tools.Tailwind,
			cfg.Tools.TailwindInput,
			cfg.Tools.TailwindOutput,
		)
		if err != nil {
			errChan <- fmt.Errorf("Cannot create tailwind: %w", err)
		}

		// Rebuild whenever a templ file changes
		debouncer := debounce.New(time.Duration(cfg.Tools.Debounce) * time.Millisecond)
		go onEvent(evFileChanged, func(ev event) bool {
			if strings.HasSuffix(ev.payload, ".templ") {
				debouncer(func() {
					tailwind.Out <- "Templ file change detected, rerunning command"
					err := tailwind.Restart(ctx, cfg, &wg, errChan, outChan)
					if err != nil {
						tailwind.Err <- err.Error()
					}
				})
			} else if strings.Contains(ev.payload, "tailwind.config") {
				debouncer(func() {
					tailwind.Out <- "Tailwind configuration changed, rerunning command"
					err := tailwind.Restart(ctx, cfg, &wg, errChan, outChan)
					if err != nil {
						tailwind.Err <- err.Error()
					}
				})
			}
			return true
		})
	} else if debug {
		outChan <- "Tailwind disabled in config"
	}

	if len(cfg.Tools.SQLC) > 0 {
		// Compile sqlc output
		sqlc, err = newCommand(
			ctx,
			"sqlc",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FCD34D")),
			false,
			true,
			"go run github.com/sqlc-dev/sqlc/cmd/sqlc@%v generate",
			cfg.Tools.SQLC,
		)
		if err != nil {
			return []*command{}, nil, fmt.Errorf("Cannot compile sqlc: %w", err)
		}

		// Rebuild whenever a sql file changes
		debouncer := debounce.New(time.Duration(cfg.Tools.Debounce) * time.Millisecond)
		go onEvent(evFileChanged, func(ev event) bool {
			if strings.HasSuffix(ev.payload, ".sql") {
				debouncer(func() {
					sqlc.Out <- "SQL file change detected, rerunning command"
					err := sqlc.Restart(ctx, cfg, &wg, errChan, outChan)
					if err != nil {
						sqlc.Err <- err.Error()
					}
				})
			} else if strings.Contains(ev.payload, "sqlc.yaml") {
				debouncer(func() {
					sqlc.Out <- "SQLC configuration changed, rerunning command"
					err := sqlc.Restart(ctx, cfg, &wg, errChan, outChan)
					if err != nil {
						sqlc.Err <- err.Error()
					}
				})
			}
			return true
		})
	} else if debug {
		outChan <- "SQLC disabled in config"
	}

	// Build go code
	compiler, err := newCommand(
		ctx,
		"go",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#c084fc")),
		false,
		templ == nil, // If templ is running, that will trigger a compilation, so we don't need to start compiling immediately
		"go build -o %v/main %v",
		cfg.Tools.BuildDir,
		cfg.Tools.MainCmd,
	)
	if err != nil {
		return []*command{}, nil, fmt.Errorf("Cannot create go builder: %w", err)
	}

	// Rebuild whenever a go file changes
	compilerDebouncer := debounce.New(time.Duration(cfg.Tools.Debounce) * time.Millisecond)
	go onEvent(evFileChanged, func(ev event) bool {
		if strings.HasSuffix(ev.payload, ".go") {
			compilerDebouncer(func() {
				compiler.Out <- "Go file change detected, rerunning command"
				err := compiler.Restart(ctx, cfg, &wg, errChan, outChan)
				if err != nil {
					compiler.Err <- err.Error()
				}
			})
		}
		return true
	})

	// Execute the app
	application, err := newCommand(
		ctx,
		"application",
		lipgloss.NewStyle(),
		true,
		false, // don't run at start, wait for the first compile to finish
		"%v/main",
		cfg.Tools.BuildDir,
	)
	if err != nil {
		return []*command{}, nil, fmt.Errorf("Cannot create application runner: %w", err)
	}

	// Restart application whenever it rebuilds
	appDebouncer := debounce.New(time.Duration(cfg.Tools.Debounce) * time.Millisecond)
	go onEvent(evCommandDone, func(ev event) bool {
		if ev.payload == "go" {
			appDebouncer(func() {
				application.Out <- "Application rebuilt, restarting"
				err := application.Restart(ctx, cfg, &wg, errChan, outChan)
				if err != nil {
					application.Err <- err.Error()
				}
			})
		}
		return true
	})

	// Restart application if a configuration file changes
	go onEvent(evFileChanged, func(ev event) bool {
		if strings.Contains(ev.payload, "/config.toml") || strings.Contains(ev.payload, "/.env") {
			appDebouncer(func() {
				application.Out <- "Configuration changed, restarting"
				err := application.Restart(ctx, cfg, &wg, errChan, outChan)
				if err != nil {
					application.Err <- err.Error()
				}
			})
		}
		return true
	})

	allCmds := []*command{
		application,
		compiler,
		templ,
		tailwind,
		sqlc,
	}

	var cmds []*command
	for _, c := range allCmds {
		if c != nil {
			cmds = append(cmds, c)
		}
	}
	return cmds, &wg, nil
}

func startRunner(
	cmds []*command,
	cfg *config.Config,
	errChan chan<- error,
	outChan chan<- string,
	wg *sync.WaitGroup,
) error {
	for _, c := range cmds {
		if !c.RunAtStart {
			continue
		}
		err := c.Run(wg, errChan)
		if err != nil {
			return fmt.Errorf("cannot start %q: %w", c.Name, err)
		}
	}

	// Set up signal capturing for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a Stop event and pass it along to the stop channel
	go onEvent(evStop, func(_ event) bool {
		sigs <- syscall.SIGINT
		return true
	})

	go func() {
		// Wait for an interrupt signal
		<-sigs

		if debug {
			outChan <- "Received shutdown signal"
		}

		// Give the subprocesses some time to exit gracefully
		time.Sleep(time.Duration(cfg.App.ShutdownTimeout) * time.Second)

		// Forcefully kill the processes if they're still running
		for _, c := range cmds {
			c.Kill(errChan, outChan)
		}
	}()

	return nil
}
