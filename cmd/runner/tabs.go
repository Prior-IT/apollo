package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	StyleSucces  = lipgloss.NewStyle().Foreground(lipgloss.Color("#1EA97C"))
	StyleSpinner = lipgloss.NewStyle().Foreground(lipgloss.Color("#139DFF"))
)

type Tabs struct {
	spinner           spinner.Model
	items             []*command
	active            int
	keyLeft, keyRight key.Binding
}

func NewTabs(
	keyLeft, keyRight key.Binding,
) *Tabs {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = StyleSpinner

	return &Tabs{
		spinner:  s,
		items:    []*command{},
		active:   0,
		keyLeft:  keyLeft,
		keyRight: keyRight,
	}
}

func (m *Tabs) setCommands(cmds []*command) *Tabs {
	// Leave the first spot nil to denote "All"
	items := make([]*command, len(cmds)+1)
	for i, cmd := range cmds {
		items[i+1] = cmd
	}

	m.items = items

	return m
}

func (m *Tabs) activeCmd() *command {
	if len(m.items) == 0 {
		return nil
	}
	return m.items[m.active]
}

func (m *Tabs) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Tabs) Update(msg tea.Msg) (*Tabs, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if len(m.items) > 0 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keyLeft):
				m.active--
				if m.active < 0 {
					m.active = len(m.items) - 1
				}
			case key.Matches(msg, m.keyRight):
				m.active++
				if m.active > len(m.items)-1 {
					m.active = 0
				}
			}
		}
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Tabs) View() string {
	line := ""
	for idx, c := range m.items {
		var item string
		if c == nil {
			item = "All"
		} else if c.Cmd.ProcessState != nil && c.Cmd.ProcessState.Exited() {
			if c.Cmd.ProcessState.ExitCode() == 0 {
				item = fmt.Sprintf(
					"%v %v (%v)",
					StyleSucces.Render("✓"),
					c.Name,
					c.Cmd.ProcessState.String(),
				)
			} else {
				item = fmt.Sprintf("%v %v (%v)", StyleError.Render("✗"), c.Name, c.Cmd.ProcessState.String())
			}
		} else if c.HasError {
			item = fmt.Sprintf("%v %v", StyleError.Render("✗"), c.Name)
		} else {
			item = fmt.Sprintf("%v %v", m.spinner.View(), c.Name)
		}

		if idx == m.active {
			line += fmt.Sprintf(" [%v] ", item)
		} else {
			line += fmt.Sprintf("  %v  ", item)
		}
	}

	return line
}
