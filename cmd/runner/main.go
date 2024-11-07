package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prior-it/apollo/config"
)

var (
	debug       bool
	showVersion bool
	runHeadless bool
)

func init() {
	args := os.Args[1:]
	flag.Usage = helpMessage
	flag.BoolVar(&debug, "d", false, "Debug mode")
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.BoolVar(&runHeadless, "nogui", false, "Headless mode")
	if err := flag.CommandLine.Parse(args); err != nil {
		log.Fatal(err)
	}
}

func helpMessage() {
	cmdName := os.Args[0]
	output := flag.CommandLine.Output()
	fmt.Fprintf(output, "Usage of %s:\n\n", cmdName)
	fmt.Fprintln(
		output,
		"This tool will run an Apollo dev server that watches your source files and live-reloads the browser page.",
	)

	fmt.Fprintln(output, "Flags:")
	flag.PrintDefaults()
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not access the current working directory: %v\n", err)
	}
	wd := os.DirFS(cwd)
	if debug {
		fmt.Printf("Working directory: %q\n", wd)
	}
	cfg, err := config.Load(wd)
	if err != nil {
		log.Fatalf("Could not find your Apollo configuration: %v\n", err)
	}
	if debug {
		fmt.Println("Apollo configuration loaded successfully")
	}

	initializeEventSystem()

	if showVersion {
		fmt.Println("Not implemented.")
	} else if runHeadless {
		if debug {
			fmt.Println("Running in headless mode")
		}
		HeadlessMode(cfg)
	} else {
		program := tea.NewProgram(
			NewUI(cfg),
			tea.WithAltScreen(),       // use the full size of the terminal in its "alternate screen buffer"
			tea.WithMouseCellMotion(), // turn on mouse support so we can track the mouse wheel
		)
		if _, err := program.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

//nolint:cyclop
func HeadlessMode(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error)
	outCh := make(chan string)

	// Global listeners
	go func() {
		list := newEventListener()
		for {
			select {
			case msg := <-outCh:
				if len(msg) > 0 {
					fmt.Fprintln(os.Stdout, msg)
				}
			case err := <-errCh:
				if err != nil {
					fmt.Fprintln(
						os.Stderr,
						StyleError.Render(fmt.Sprintf("Error: %v", err.Error())),
					)
				}
			case ev := <-list:
				if debug {
					fmt.Fprintln(
						os.Stdout,
						StyleEvent.Render(fmt.Sprintf("[event] %v", ev.String())),
					)
				}
			}
		}
	}()

	go func() {
		<-ctx.Done()
		newStopEvent()
	}()

	// Start the runner
	cmds, wg, err := createCommands(ctx, errCh, outCh, cfg)
	if err != nil {
		log.Fatalf("cannot create commands: %v\n", err)
	}

	// Command listeners
	for _, cmd := range cmds {
		go func() {
		loop:
			for {
				select {
				case msg, ok := <-cmd.Out:
					if !ok {
						fmt.Fprintf(os.Stdout, "[%v] Out channel closed\n", cmd.Name)
						break loop
					}
					if len(msg) > 0 {
						fmt.Fprintln(os.Stdout, cmd.Style.Render(fmt.Sprintf("[%v] %v", cmd.Name, msg)))
					}

				case msg, ok := <-cmd.Err:
					if !ok {
						fmt.Fprintf(os.Stdout, "[%v] Err channel closed\n", cmd.Name)
						break loop
					}
					if len(msg) > 0 {
						fmt.Fprintln(os.Stdout, StyleError.Render(fmt.Sprintf("[%v] %v", cmd.Name, msg)))
					}
				case <-ctx.Done():
					break loop
				}
			}
		}()
	}

	err = startRunner(cmds, cfg, errCh, outCh, wg)
	if err != nil {
		log.Fatalf("cannot start runner: %v", err)
	}

	wg.Wait()
	fmt.Println("Bye!")
}
