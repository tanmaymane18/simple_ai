package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"simple_ai/client"
)

const gap = "\n\n"

var ac simple_ai.AdkClient

type errMsg error

type noop struct{}

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	aiStyle     lipgloss.Style
	err         error
}

func initModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "│ "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to hte chat room!
Type a message and press Enter to send.`)
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ac = simple_ai.InitAdkClient()
	return model{
		viewport:    vp,
		textarea:    ta,
		messages:    []string{},
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		aiStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("202")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return ac.CreateSession
}

func (m model) View() string {

	return fmt.Sprintf(
		"%s%s%s",
		m.viewport.View(),
		gap,
		m.textarea.View(),
	)
}

func noopCmd() tea.Msg {
	return noop{}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		saCmd tea.Cmd
	)
	saCmd = noopCmd
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case simple_ai.MsgStatus:
		if bool(msg) {
			fmt.Println("Session init properly")
		} else {
			fmt.Println("Session init failed!")
		}
	case simple_ai.MsgSuccessResponse:
		msg = simple_ai.MsgSuccessResponse(msg)
		m.messages = append(m.messages, m.aiStyle.Render("AI: ")+msg.Response)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.textarea.Reset()
		m.viewport.GotoBottom()
	case simple_ai.MsgFailureResponse:
		msg = simple_ai.MsgFailureResponse(msg)
		m.messages = append(m.messages, m.aiStyle.Render("AI: ")+msg.FailureMsg)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.textarea.Reset()
		m.viewport.GotoBottom()
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)
		if len(m.messages) > 0 {
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}
		m.viewport.GotoBottom()
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			saCmd = func() tea.Msg {
				return ac.MakeRequest(m.textarea.Value())
			}

		}
	case errMsg:
		m.err = msg
		return m, nil
	}
	return m, tea.Batch(tiCmd, vpCmd, saCmd)
}

func main() {
	p := tea.NewProgram(initModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
