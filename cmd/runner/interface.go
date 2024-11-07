package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prior-it/apollo/config"
)

var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#139DFF")).
			Align(lipgloss.Center)
	StyleError    = lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626"))
	StyleDefault  = lipgloss.NewStyle().Foreground(lipgloss.NoColor{})
	StyleEvent    = lipgloss.NewStyle().Foreground(lipgloss.Color("#525252"))
	StyleViewport = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
)

var keys = keyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "tab left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "tab right    "),
	),
	// Up is defined by the viewport, we're just putting it here for the help menu
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	// Down is defined by the viewport, we're just putting it here for the help menu
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down    "),
	),
	// PageUp is defined by the viewport, we're just putting it here for the help menu
	PageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "page up"),
	),
	// PageDown is defined by the viewport, we're just putting it here for the help menu
	PageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "page down    "),
	),
	ToTop: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "scroll to top"),
	),
	ToBottom: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "scroll to bottom    "),
	),
	Restart: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "restart"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help    "),
	),
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clear output    "),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	ForceQuit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "force quit    "),
	),
}

type RunnerUI struct {
	initialised     bool
	quitting        bool
	followBottom    bool
	ctxRunner       context.Context
	cancelRunner    context.CancelFunc
	err             error
	cmds            []*command
	activeCmd       int
	cmdsWg          *sync.WaitGroup
	cfg             *config.Config
	spinner         spinner.Model
	keys            keyMap
	help            help.Model
	tabs            *Tabs
	viewport        viewport.Model
	viewportContent []string
	viewportChan    chan string
	errChan         chan error
	outChan         chan string
}

func NewUI(cfg *config.Config) *RunnerUI {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = StyleTitle

	t := NewTabs(keys.Left, keys.Right)

	h := help.New()
	// This, as well as the spaces in the help descriptions above, are workarounds for a bug with
	// column spacing in the current version of bubbles. This will be fixed in v.0.22.
	h.Styles.FullSeparator = lipgloss.NewStyle().Width(0)

	viewportContent := WelcomeMessage()

	return &RunnerUI{
		initialised:     false,
		activeCmd:       0,
		quitting:        false,
		followBottom:    true,
		spinner:         s,
		cfg:             cfg,
		help:            h,
		keys:            keys,
		tabs:            t,
		viewportContent: viewportContent,
	}
}

func WelcomeMessage() []string {
	text := []string{
		"Welcome to the Apollo Runner test release!",
		"Press Q / CTRL+C to quit the runner.",
		"Press CTRL+Q to force-quit the runner. Note that this will not kill any subprocesses, things might keep files or ports open if you use this.",
		"Press R to restart the command in the currently active tab (or in the \"All\" tab to restart all commands)",
		"Press ? to view all commands.",
		"In case of issues, start the runner with the -d flag and contact robin.arys@prior-it.be to help debug (you can use Slack as well :P).",
		"glhf!",
	}
	formatted := make([]string, len(text))
	for i, t := range text {
		formatted[i] = StyleTitle.Render("  > "+t) + "\n"
	}
	// Append empty newline
	formatted[len(formatted)-1] += "\n"
	return formatted
}

//nolint:cyclop
func (ui *RunnerUI) StartRunner() {
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)
	outChan := make(chan string)
	ui.ctxRunner = ctx
	ui.cancelRunner = cancel
	ui.viewportChan = make(chan string)
	ui.outChan = outChan
	ui.errChan = errChan

	// Print all output to viewport
	go func() {
		for msg := range ui.viewportChan {
			ui.viewportContent = append(ui.viewportContent, msg)
			ui.UpdateViewportContent()
		}
	}()

	// Global listeners - cannot be combined with the viewportChan listener since these often send to
	// viewportChan as well
	go func() {
		list := newEventListener()
		for {
			select {
			case msg := <-outChan:
				if len(msg) > 0 {
					ui.viewportChan <- fmt.Sprintf("[system] %v", msg) + "\n"
				}
			case err := <-errChan:
				if err != nil {
					ui.viewportChan <- StyleError.Render(fmt.Sprintf("[error] %v", err)) + "\n"
				}

			case ev := <-list:
				ui.viewportChan <- StyleEvent.Render(fmt.Sprintf("[event] %v", ev.String())) + "\n"
			}
		}
	}()

	// Create commands - this might output to outChan or errChan
	cmds, wg, err := createCommands(ctx, errChan, outChan, ui.cfg)
	if err != nil {
		log.Fatalf("cannot create commands: %v", err)
	}
	ui.cmds = cmds
	ui.tabs = ui.tabs.setCommands(cmds)
	ui.cmdsWg = wg

	// Send stop signal if this context is cancelled
	go func() {
		<-ctx.Done()
		newStopEvent()
		ui.viewportChan <- "\n" + StyleDefault.Render("Shutting down...") + "\n"
		ui.viewportChan <- StyleDefault.Render(fmt.Sprintf("Time-out is set to %vs", ui.cfg.App.ShutdownTimeout)) + "\n"
	}()

	// Command listeners
	for _, cmd := range cmds {
		go func() {
			for {
				select {
				case msg := <-cmd.Out:
					if len(msg) > 0 {
						cmd.HasError = false
						ui.viewportChan <- cmd.Style.Width(ui.viewport.Width).Render(fmt.Sprintf("[%v] %v", cmd.Name, msg))
					}
				case msg := <-cmd.Err:
					if len(msg) > 0 {
						cmd.HasError = true
						ui.viewportChan <- StyleError.Width(ui.viewport.Width).Render(fmt.Sprintf("[%v] %v", cmd.Name, msg))
					}
				}
			}
		}()
	}

	// Actually start the runner now that all channels are handled correctly
	err = startRunner(cmds, ui.cfg, errChan, outChan, wg)
	if err != nil {
		log.Fatalf("cannot start runner: %v", err)
	}

	// Quit if all commands in the waitgroup are finished
	go func() {
		wg.Wait()
		ui.viewportChan <- "Done\n"
		ui.quitting = true
	}()
}

func (ui *RunnerUI) UpdateViewportContent() {
	var content string

	cmd := ui.tabs.activeCmd()
	if cmd == nil {
		content = strings.Join(ui.viewportContent, "")
	} else {
		// Filter on command
		content = ""
		cmdPrefix := fmt.Sprintf("[%v]", cmd.Name)
		for _, line := range ui.viewportContent {
			containsName := strings.Contains(line, cmd.Name)
			isGlobal := strings.Contains(line, "[event]") || strings.Contains(line, "[error]") || strings.Contains(line, "[system]")
			if (isGlobal && containsName) || strings.Contains(line, cmdPrefix) {
				content += line
			}
		}
	}
	content += "\n\n"

	str := lipgloss.NewStyle().Width(ui.viewport.Width).Render(content)
	ui.viewport.SetContent(str)
	if ui.followBottom {
		ui.viewport.GotoBottom()
	}
}

func (ui *RunnerUI) clearViewport() {
	ui.viewportContent = WelcomeMessage()
	ui.UpdateViewportContent()
}

func (ui *RunnerUI) Init() tea.Cmd {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	cmd = ui.spinner.Tick
	cmds = append(cmds, cmd)

	cmd = ui.tabs.Init()
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (ui *RunnerUI) RestartCurrentTab() {
	if cmd := ui.tabs.activeCmd(); cmd != nil {
		err := cmd.Restart(ui.ctxRunner, ui.cfg, ui.cmdsWg, ui.errChan, ui.outChan)
		if err != nil {
			ui.errChan <- err
		}
	} else {
		for _, cmd := range ui.cmds {
			err := cmd.Restart(ui.ctxRunner, ui.cfg, ui.cmdsWg, ui.errChan, ui.outChan)
			if err != nil {
				ui.errChan <- err
			}
		}
	}
}

//nolint:cyclop
func (ui *RunnerUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	if ui.quitting {
		return ui, tea.Quit
	}

	// Update tabs before key presses so we can update the viewport content more easily
	ui.tabs, cmd = ui.tabs.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Key presses
		switch {
		case key.Matches(msg, ui.keys.ForceQuit):
			return ui, tea.Quit
		case key.Matches(msg, ui.keys.Quit):
			ui.cancelRunner()
		case key.Matches(msg, ui.keys.Clear):
			ui.clearViewport()
		case key.Matches(msg, ui.keys.Help):
			ui.help.ShowAll = !ui.help.ShowAll
		case key.Matches(msg, ui.keys.Up):
			ui.followBottom = ui.viewport.AtBottom()
		case key.Matches(msg, ui.keys.Down):
			ui.followBottom = ui.viewport.AtBottom()
		case key.Matches(msg, ui.keys.Left):
			ui.UpdateViewportContent()
		case key.Matches(msg, ui.keys.Right):
			ui.UpdateViewportContent()
		case key.Matches(msg, ui.keys.ToTop):
			ui.followBottom = false
			ui.viewport.GotoTop()
		case key.Matches(msg, ui.keys.ToBottom):
			ui.followBottom = true
			ui.viewport.GotoBottom()
		case key.Matches(msg, ui.keys.Restart):
			go ui.RestartCurrentTab()
		}

	case tea.WindowSizeMsg:
		ui.help.Width = msg.Width
		headerHeight := lipgloss.Height(ui.HeaderView())
		footerHeight := lipgloss.Height(ui.FooterView())
		verticalMargin := headerHeight + footerHeight - 1
		if !ui.initialised {
			ui.viewport = viewport.New(msg.Width, msg.Height-verticalMargin)
			// ui.viewport.HighPerformanceRendering = true
			// ui.viewport.YPosition = headerHeight + 1
			ui.viewport.YPosition = headerHeight
			ui.UpdateViewportContent()
			ui.viewport.Style = StyleViewport
			ui.initialised = true
			ui.StartRunner()
		} else {
			ui.viewport.Width = msg.Width
			ui.viewport.Height = msg.Height - verticalMargin
		}
		// cmds = append(cmds, viewport.Sync(ui.viewport))

	case error:
		ui.err = msg

	}

	ui.spinner, cmd = ui.spinner.Update(msg)
	cmds = append(cmds, cmd)

	ui.viewport, cmd = ui.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return ui, tea.Batch(cmds...)
}

func (ui *RunnerUI) View() string {
	if ui.err != nil {
		return ui.err.Error()
	}
	if ui.quitting {
		return "\nBye!\n"
	}
	if !ui.initialised {
		return fmt.Sprintf("\n%vInitialising…\n", ui.spinner.View())
	}

	header := ui.HeaderView()
	viewport := ui.viewport.View()
	footer := ui.FooterView()
	return header + viewport + footer
}

// The view above the main output view
func (ui *RunnerUI) HeaderView() string {
	header := StyleTitle.Width(ui.viewport.Width).
		Render(fmt.Sprintf("%v Apollo Runner %v", ui.spinner.View(), ui.spinner.View())) +
		"\n"

	header += ui.tabs.View()

	return header + "\n"
}

// The view underneath the main output view
func (ui *RunnerUI) FooterView() string {
	helpView := ui.help.View(ui.keys)

	// Center the help view
	helpView = lipgloss.NewStyle().
		Width(ui.viewport.Width).
		AlignHorizontal(lipgloss.Center).
		Render(helpView)

	var errView string
	if ui.err != nil {
		errView = "\n" + StyleError.
			Width(ui.viewport.Width).
			AlignHorizontal(lipgloss.Center).
			Render(fmt.Sprintf("[ERROR] %v", ui.err.Error()))
	}
	return "\n" + helpView + errView
}

// KeyMap is basically only used to generate the help menu, we don't allow rebinding keys atm
type keyMap struct {
	Left      key.Binding
	Right     key.Binding
	Up        key.Binding
	Down      key.Binding
	PageUp    key.Binding
	PageDown  key.Binding
	ToTop     key.Binding
	ToBottom  key.Binding
	Restart   key.Binding
	Help      key.Binding
	Clear     key.Binding
	Quit      key.Binding
	ForceQuit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit, k.Restart, k.Help}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Left, k.Right},
		{k.Up, k.Down},
		{k.PageUp, k.PageDown},
		{k.ToTop, k.ToBottom},
		{k.Restart, k.Help},
		{k.Clear},
		{k.Quit, k.ForceQuit},
	}
}
