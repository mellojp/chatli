package ui

import (
	"strings"

	"github.com/mellojp/chatli/api"
	"github.com/mellojp/chatli/data"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

type uiState int

const (
	loginView uiState = iota
	registerView
	roomListView
	createRoomView
	chatView
	joinRoomView
)

type SocketError struct {
	Err error
}

type Model struct {
	State        uiState
	ChatsHistory map[string][]data.Message

	// Inputs específicos
	UsernameInput textinput.Model
	PasswordInput textinput.Model
	ChatInput     textarea.Model
	GenericInput  textarea.Model // Para Room ID ou Room Name

	InputIndex int // 0 para User, 1 para Password

	Viewport    viewport.Model
	Cursor      int
	CurrentRoom string
	Session     data.Session

	WindowHeight int
	WindowWidth  int
	ErrorMsg     string
	SuccessMsg   string
	WSConn       *websocket.Conn
}

func NewModel() *Model {
	// Configuração do Input de Username
	userIn := textinput.New()
	userIn.Placeholder = "Username"
	userIn.Focus()
	userIn.Prompt = ""

	// Configuração do Input de Password
	passIn := textinput.New()
	passIn.Placeholder = "Password"
	passIn.Prompt = ""
	passIn.EchoMode = textinput.EchoPassword
	passIn.EchoCharacter = '•'

	// Configuração do Chat Input
	chatIn := textarea.New()
	chatIn.Placeholder = "Type a message..."
	chatIn.Prompt = ""
	chatIn.SetHeight(1)
	chatIn.ShowLineNumbers = false

	// Configuração Genérica
	genIn := textarea.New()
	genIn.Prompt = ""
	genIn.SetHeight(1)
	genIn.ShowLineNumbers = false

	vp := viewport.New(80, 20)

	return &Model{
		State:         loginView,
		ChatsHistory:  make(map[string][]data.Message),
		UsernameInput: userIn,
		PasswordInput: passIn,
		ChatInput:     chatIn,
		GenericInput:  genIn,
		InputIndex:    0,
		Viewport:      vp,
	}
}

func (*Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, textinput.Blink)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		// Navegação via TAB e SETAS (Troca de Campo ou Navegação na Lista)
		case "tab", "shift+tab", "up", "down":
			switch m.State {
			case loginView, registerView:
				// Cicla entre 0 e 1
				m.InputIndex = (m.InputIndex + 1) % 2

				if m.InputIndex == 0 {
					m.UsernameInput.Focus()
					m.PasswordInput.Blur()
				} else {
					m.UsernameInput.Blur()
					m.PasswordInput.Focus()
				}
				return m, nil
			case roomListView:
				if msg.String() == "up" {
					if m.Cursor > 0 {
						m.Cursor--
					}
				} else {
					if m.Cursor < len(m.Session.JoinedRooms)-1 {
						m.Cursor++
					}
				}
				return m, nil
			}

		case "enter":
			m.ErrorMsg = ""
			m.SuccessMsg = ""
			switch m.State {
			case loginView:
				// Só tenta logar se o campo de password estiver focado
				if m.InputIndex == 1 {
					session, err := api.Login(m.UsernameInput.Value(), m.PasswordInput.Value())
					if err != nil {
						m.ErrorMsg = err.Error()
						return m, nil
					}
					m.Session = *session

					// Carrega as salas do usuário do servidor
					rooms, err := api.GetUserRooms(m.Session)
					if err == nil {
						m.Session.JoinedRooms = rooms
					}

					// Conecta WebSocket Global
					wsConn, err := api.ConnectWebSocket(m.Session)
					if err != nil {
						m.ErrorMsg = "Erro WS: " + err.Error()
						return m, nil
					}
					m.WSConn = wsConn

					m.State = roomListView
					m.UsernameInput.Reset()
					m.PasswordInput.Reset()
					return m, WaitForMessage(m.WSConn)
				}
				// Se ENTER for pressionado no campo de username, não faz nada
				return m, nil

			case registerView:
				// Só tenta registrar se o campo de password estiver focado
				if m.InputIndex == 1 {
					err := api.Register(m.UsernameInput.Value(), m.PasswordInput.Value())
					if err != nil {
						m.ErrorMsg = err.Error()
						return m, nil
					}
					m.State = loginView
					m.SuccessMsg = "Registrado! Faça login."
					m.InputIndex = 1 // Foca na senha pra agilizar no login
					m.UsernameInput.Blur()
					m.PasswordInput.Focus()
					return m, nil
				}
				// Se ENTER for pressionado no campo de username, não faz nada
				return m, nil
			case createRoomView:
				roomName := m.GenericInput.Value()
				if roomName == "" {
					return m, nil
				}
				newRoom, err := api.CreateRoom(m.Session, roomName)
				if err != nil {
					m.ErrorMsg = err.Error()
					return m, nil
				}
				m.Session.JoinedRooms = append(m.Session.JoinedRooms, *newRoom)
				m.State = roomListView
				m.GenericInput.Reset()
				return m, nil

			case joinRoomView:
				roomId := m.GenericInput.Value()
				err := api.JoinRoom(m.Session, roomId)
				if err != nil {
					m.ErrorMsg = err.Error()
					return m, nil
				}

				// Atualiza lista de salas para pegar o nome
				rooms, err := api.GetUserRooms(m.Session)
				if err == nil {
					m.Session.JoinedRooms = rooms
				}

				m.State = chatView
				m.CurrentRoom = roomId

				// Carrega histórico da nova sala
				v, _ := api.LoadChatMessages(m.Session, m.CurrentRoom)
				m.ChatsHistory[m.CurrentRoom] = v
				m.Viewport.SetContent(RenderChatView(m))
				m.Viewport.GotoBottom()

				m.GenericInput.Reset()
				m.ChatInput.Focus()
				return m, nil

			case chatView:
				content := m.ChatInput.Value()
				if content == "" {
					return m, nil
				}
				// Usa conexao global
				msg := data.Message{
					Type:    "chat",
					Content: content,
					RoomId:  m.CurrentRoom,
					// Id e SentAt são gerados no servidor
				}
				if err := m.WSConn.WriteJSON(msg); err != nil {
					m.ErrorMsg = "Erro ao enviar: " + err.Error()
					return m, nil
				}
				m.ChatInput.Reset()
				return m, nil
			case roomListView:
				if len(m.Session.JoinedRooms) == 0 {
					return m, nil
				}
				m.CurrentRoom = m.Session.JoinedRooms[m.Cursor].Id
				m.State = chatView

				v, _ := api.LoadChatMessages(m.Session, m.CurrentRoom)
				m.ChatsHistory[m.CurrentRoom] = v
				m.Viewport.SetContent(RenderChatView(m))
				m.Viewport.GotoBottom()
				m.ChatInput.Focus()

				return m, nil
			}

		case "n":
			if m.State == roomListView {
				m.State = createRoomView
				m.GenericInput.Placeholder = "Room Name"
				m.GenericInput.Reset()
				m.GenericInput.Focus()
				return m, nil
			}
		case "e":
			if m.State == roomListView {
				m.State = joinRoomView
				m.GenericInput.Placeholder = "Room UUID"
				m.GenericInput.Reset()
				m.GenericInput.Focus()
				m.ErrorMsg = ""
				return m, nil
			}
		case "esc":
			m.ErrorMsg = ""
			m.SuccessMsg = ""
			switch m.State {
			case roomListView:
				m.State = loginView
				m.UsernameInput.Reset()
				m.PasswordInput.Reset()
				m.InputIndex = 0
				m.UsernameInput.Focus()
				m.PasswordInput.Blur()
			case loginView:
				m.State = registerView
				m.ErrorMsg = ""
				m.SuccessMsg = ""
				m.UsernameInput.Reset()
				m.PasswordInput.Reset()
				m.InputIndex = 0
				m.UsernameInput.Focus()
				m.PasswordInput.Blur()
				return m, nil
			case registerView:
				m.State = loginView
				m.InputIndex = 0
				m.UsernameInput.Focus()
				m.PasswordInput.Blur()
			case joinRoomView, chatView, createRoomView:
				m.State = roomListView
			}
		case "pgup", "pgdown":
			if m.State == chatView {
				m.Viewport, cmd = m.Viewport.Update(msg)
				return m, cmd
			}
		case "left", "right":
			if m.State == chatView {
				m.Viewport, cmd = m.Viewport.Update(msg)
				return m, cmd
			}
		}

	case tea.WindowSizeMsg:
		m.WindowHeight = msg.Height
		m.WindowWidth = msg.Width
		m.ChatInput.SetWidth(m.WindowWidth)
		// Input ocupa largura total menos margem estimada (labels etc)
		inputWidth := m.WindowWidth - 20
		if inputWidth < 10 {
			inputWidth = 10
		}
		m.UsernameInput.Width = inputWidth
		m.PasswordInput.Width = inputWidth
		m.GenericInput.SetWidth(inputWidth)
		m.Viewport.Height = m.WindowHeight - 4 - m.ChatInput.Height() - 1 // Ajusta para o chat input
		m.Viewport.Width = m.WindowWidth

	case data.Message:
		m.ChatsHistory[msg.RoomId] = append(m.ChatsHistory[msg.RoomId], msg)
		if m.CurrentRoom == msg.RoomId {
			m.Viewport.SetContent(RenderChatView(m))
			m.Viewport.GotoBottom()
		}
		return m, WaitForMessage(m.WSConn)
	}

	// Update dos TextAreas
	switch m.State {
	case loginView, registerView:
		m.UsernameInput, cmd = m.UsernameInput.Update(msg)
		cmds = append(cmds, cmd)
		m.PasswordInput, cmd = m.PasswordInput.Update(msg)
		cmds = append(cmds, cmd)
	case chatView:
		m.ChatInput, cmd = m.ChatInput.Update(msg)
		cmds = append(cmds, cmd)
	case createRoomView, joinRoomView:
		m.GenericInput, cmd = m.GenericInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update Viewport
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var s string
	switch m.State {
	case loginView:
		s = RenderLogin(m)
	case registerView:
		s = RenderRegister(m)
	case roomListView:
		s = RenderRoomsList(m)
	case createRoomView:
		s = RenderCreateRoom(m)
	case joinRoomView:
		s = RenderJoinRoom(m)
	case chatView:
		var roomName string
		for _, r := range m.Session.JoinedRooms {
			if r.Id == m.CurrentRoom {
				roomName = r.Name
				break
			}
		}
		if roomName == "" {
			roomName = "Unknown Room"
		}

		title := RoomTitleStyle.Render("Room: " + roomName)
		back := HelpStyle.Render("[esc] back")
		gap := m.WindowWidth - lipgloss.Width(title) - lipgloss.Width(back)
		if gap < 0 {
			gap = 0
		}

		header := title + strings.Repeat(" ", gap) + back + "\n"
		subheader := RoomIdStyle.Render("ID: "+m.CurrentRoom) + "\n"
		separator := HelpStyle.Render(strings.Repeat("─", m.WindowWidth)) + "\n"
		body := m.Viewport.View()
		prompt := SystemStyle.Render("$ ") + InputStyle.Render(m.ChatInput.View())
		s = header + subheader + separator + body + "\n" + separator + prompt
		if m.ErrorMsg != "" {
			s += "\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
		}
	}
	return lipgloss.Place(m.WindowWidth, m.WindowHeight, lipgloss.Left, lipgloss.Top, AppStyle.Render(s))
}

func WaitForMessage(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		msg := data.Message{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			return SocketError{Err: err}
		}
		return msg
	}
}
