package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const ascii = `
 ________  ___  ___  ________  _________  ___       ___     
|\   ____\|\  \|\  \|\   __  \|\___   ___\\  \     |\  \    
\ \  \___|\ \  \\\  \ \  \|\  \|___ \  \_\ \  \    \ \  \   
 \ \  \    \ \   __  \ \   __  \   \ \  \ \ \  \    \ \  \  
  \ \  \____\ \  \ \  \ \  \ \  \   \ \  \ \ \  \____\ \  \ 
   \ \_______\ \__\ \__\ \__\ \__\   \ \__\ \ \_______\ \__\
    \|_______|\|__|\|__|\|__|\|__|    \|__|  \|_______|\|__|
`

func renderAsciiHeader(m *Model) string {
	return lipgloss.PlaceHorizontal(m.WindowWidth, lipgloss.Center, SystemStyle.Render(ascii)) + "\n\n"
}

func RenderLogin(m *Model) string {
	s := renderAsciiHeader(m)
	s += SystemStyle.Render("user@terminal:~$ login") + "\n\n"

	usernameField := m.UsernameInput.View()
	passwordField := m.PasswordInput.View()

	// Adiciona estilo de foco
	if m.InputIndex == 0 {
		usernameField = SelectedItemStyle.Render(usernameField)
	} else {
		passwordField = SelectedItemStyle.Render(passwordField)
	}

	s += "Username: " + usernameField + "\n"
	s += "Password: " + passwordField + "\n\n"

	s += HelpStyle.Render("[tab/arrows] switch fields | [enter] login | [esc] register new account")

	if m.SuccessMsg != "" {
		s += "\n\n" + SuccessStyle.Render(m.SuccessMsg)
	}
	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	return s
}

func RenderRegister(m *Model) string {
	s := renderAsciiHeader(m)
	s += SystemStyle.Render("user@terminal:~$ register") + "\n\n"

	usernameField := m.UsernameInput.View()
	passwordField := m.PasswordInput.View()

	// Adiciona estilo de foco
	if m.InputIndex == 0 {
		usernameField = SelectedItemStyle.Render(usernameField)
	} else {
		passwordField = SelectedItemStyle.Render(passwordField)
	}

	s += "Username: " + usernameField + "\n"
	s += "Password: " + passwordField + "\n\n"

	s += HelpStyle.Render("[tab/arrows] switch fields | [enter] create account | [esc] back to login")

	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	return s
}

func RenderRoomsList(m *Model) string {
	s := fmt.Sprintf("%s@terminal:~$ /chatli/rooms ls", m.Session.Username)
	s = SystemStyle.Render(s) + "\n"

	if len(m.Session.JoinedRooms) == 0 {
		s += HelpStyle.Render("(Nenhuma sala recente)") + "\n"
	}

	for i, room := range m.Session.JoinedRooms {
		// Formatação: "Nome da Sala" (ID)
		line := fmt.Sprintf("%s %s", room.Name, RoomIdStyle.Render("("+room.Id+")"))

		if i == m.Cursor {
			s += SelectedItemStyle.Render("> "+line) + "\n"
		} else {
			s += UnselectedItemStyle.Render("  "+line) + "\n"
		}
	}
	s += "\n" + HelpStyle.Render("[up/down] nav | [n] new room | [e] enter room | [enter] select | [esc] logout")
	if m.ErrorMsg != "" {
		s += "\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	return s
}

func RenderCreateRoom(m *Model) string {
	s := fmt.Sprintf("%s@terminal:~$ create-room", m.Session.Username)
	s = SystemStyle.Render(s) + "\n"
	s += SystemStyle.Render("Room Name: ") + InputStyle.Render(m.GenericInput.View())

	s += "\n\n" + HelpStyle.Render("[enter] create | [esc] cancel")

	if m.ErrorMsg != "" {
		s += "\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	return s
}

func RenderJoinRoom(m *Model) string {
	s := fmt.Sprintf("%s@terminal:~$ join-room --id", m.Session.Username)
	s = SystemStyle.Render(s) + "\n"
	s += SystemStyle.Render("Room ID: ") + InputStyle.Render(m.GenericInput.View())

	s += "\n\n" + HelpStyle.Render("[enter] join | [esc] cancel")

	if m.ErrorMsg != "" {
		s += "\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	return s
}

func RenderChatView(m *Model) string {
	var b strings.Builder
	renderWidth := m.WindowWidth - 4 // Margem de segurança

	for _, val := range m.ChatsHistory[m.CurrentRoom] {
		if val.Content == "" {
			continue
		}

		displayTime := val.SentAt.Format("15:04")

		var line string
		var alignedLine string

		if val.UserId == m.Session.UserId {
			// Mensagem do próprio usuário (Direita)
			senderName := "Você"
			styledUser := RoomTitleStyle.Render(senderName) // Verde (cor 2)

			line = fmt.Sprintf("%s <%s> [%s]", val.Content, styledUser, TimeStyle.Render(displayTime))
			alignedLine = lipgloss.NewStyle().Width(renderWidth).Align(lipgloss.Right).Render(line)
		} else {
			// Mensagem de outros (Esquerda)
			senderName := val.SenderUsername
			if senderName == "" {
				senderName = val.UserId // Fallback se não tiver username
			}
			styledUser := SenderStyle.Render(senderName) // Verde escuro (cor 22)

			line = fmt.Sprintf("[%s] <%s> %s", TimeStyle.Render(displayTime), styledUser, val.Content)
			alignedLine = lipgloss.NewStyle().Width(renderWidth).Align(lipgloss.Left).Render(line)
		}

		b.WriteString(alignedLine + "\n")
	}
	return b.String()
}
