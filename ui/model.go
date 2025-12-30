package ui

import (
	"github.com/mellojp/chatli/api"
	"github.com/mellojp/chatli/data"

	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

type uiState int

const (
	loginView uiState = iota
	roomListView
	chatView
	joinRoomView
)

type SocketError struct {
	RoomId string
	Err    error
}

type ReconnectMsg struct {
	RoomId string
}

type Model struct {
	State        uiState
	ChatsHistory map[string][]data.Message
	TextArea     textarea.Model
	Viewport     viewport.Model
	Cursor       int
	CurrentRoom  string
	Session      data.Session
	WindowHeight int
	WindowWidth  int
	ErrorMsg     string
	SocketsConns map[string]*websocket.Conn
}

func NewModel() *Model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.SetHeight(1)

	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(80, 20)

	return &Model{
		State:        loginView,
		ChatsHistory: map[string][]data.Message{},
		TextArea:     ta,
		Viewport:     vp,
		Cursor:       0,
		CurrentRoom:  "",
		ErrorMsg:     "",
		SocketsConns: map[string]*websocket.Conn{},
	}
}

func (*Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd, vpCmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			m.ErrorMsg = ""
			switch m.State {
			case loginView:
				name := m.TextArea.Value()
				if name == "" {
					return m, nil
				}
				session, err := api.CreateSession(name)
				if err != nil {
					m.ErrorMsg = err.Error()
					return m, nil
				}
				m.Session = *session
				m.State = roomListView
				m.TextArea.Reset()
				return m, nil

			case roomListView:
				if len(m.Session.JoinedRooms) == 0 {
					return m, nil
				}
				m.CurrentRoom = m.Session.JoinedRooms[m.Cursor]
				m.State = chatView
				m.TextArea.Reset()
				api.JoinRoom(m.Session, m.CurrentRoom)
				v, _ := api.LoadChatMessages(m.Session, m.CurrentRoom, 50)
				m.ChatsHistory[m.CurrentRoom] = v
				m.Viewport.SetContent(RenderChatView(m))
				m.Viewport.GotoBottom()

				conn, ok := m.SocketsConns[m.CurrentRoom]
				if !ok || conn == nil {
					wsConn, err := api.ConnectWebSocket(m.Session, m.CurrentRoom)
					if err != nil {
						m.ErrorMsg = err.Error()
						return m, nil
					}
					m.SocketsConns[m.CurrentRoom] = wsConn
					return m, WaitForMessage(wsConn, m.CurrentRoom)
				}
				return m, nil

			case joinRoomView:
				roomId := m.TextArea.Value()
				err := api.JoinRoom(m.Session, roomId)
				if err != nil {
					m.ErrorMsg = err.Error()
					return m, nil
				}
				wsConn, _ := api.ConnectWebSocket(m.Session, roomId)
				m.SocketsConns[roomId] = wsConn
				m.Session.JoinedRooms = append(m.Session.JoinedRooms, roomId)
				m.State = chatView
				m.CurrentRoom = roomId
				m.TextArea.Reset()
				return m, WaitForMessage(wsConn, roomId)

			case chatView:
				content := m.TextArea.Value()
				if content == "" {
					return m, nil
				}
				conn := m.SocketsConns[m.CurrentRoom]
				msg := data.Message{
					Type: "chat", User: m.Session.Username,
					Timestamp: time.Now().Format("2006-01-02 15:04:05"),
					Content:   content, RoomId: m.CurrentRoom,
				}
				conn.WriteJSON(msg)
				m.TextArea.Reset()
				return m, nil
			}
		case "n":
			if m.State == roomListView {
				newRoom, _ := api.CreateRoom(m.Session)
				m.Session.JoinedRooms = append(m.Session.JoinedRooms, newRoom.Id)
				return m, nil
			}
		case "e":
			if m.State == roomListView {
				m.State = joinRoomView
				m.TextArea.Reset()
				m.ErrorMsg = ""
				return m, nil
			}
		case "up":
			if m.Cursor > 0 && m.State == roomListView {
				m.Cursor--
			}
		case "down":
			if m.Cursor < len(m.Session.JoinedRooms)-1 && m.State == roomListView {
				m.Cursor++
			}
		case "esc":
			m.ErrorMsg = ""
			switch m.State {
			case roomListView:
				m.State = loginView
			case joinRoomView, chatView:
				m.State = roomListView
			}
			m.TextArea.Reset()
		}
	case tea.WindowSizeMsg:
		m.WindowHeight = msg.Height
		m.WindowWidth = msg.Width
		m.TextArea.SetWidth(msg.Width)
		m.Viewport.Height = msg.Height - 4
		m.Viewport.Width = msg.Width

	case data.Message:
		m.ChatsHistory[msg.RoomId] = append(m.ChatsHistory[msg.RoomId], msg)
		m.Viewport.SetContent(RenderChatView(m))
		m.Viewport.GotoBottom()
		return m, WaitForMessage(m.SocketsConns[msg.RoomId], msg.RoomId)
	}

	m.TextArea, cmd = m.TextArea.Update(msg)
	if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "k" || key.String() == "j") {
		return m, cmd
	}
	m.Viewport, vpCmd = m.Viewport.Update(msg)
	return m, tea.Batch(cmd, vpCmd)
}

func (m *Model) View() string {
	var s string
	switch m.State {
	case loginView:
		s = RenderLogin(m)
	case roomListView:
		s = RenderRoomsList(m)
	case joinRoomView:
		s = RenderJoinRoom(m)
	case chatView:
		header := RoomTitleStyle.Render(fmt.Sprintf("%s@terminal:~$ /chatli/rooms/%s tail", m.Session.Username, m.CurrentRoom)) + "\n"
		body := m.Viewport.View()
		prompt := SystemStyle.Render("$ ") + InputStyle.Render(m.TextArea.View())
		s = header + body + "\n" + prompt
		if m.ErrorMsg != "" {
			s += "\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
		}
	}
	return lipgloss.Place(m.WindowWidth, m.WindowHeight, lipgloss.Left, lipgloss.Top, AppStyle.Render(s))
}

func WaitForMessage(conn *websocket.Conn, roomId string) tea.Cmd {
	return func() tea.Msg {
		msg := data.Message{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			return SocketError{RoomId: roomId, Err: err}
		}
		return msg
	}
}
