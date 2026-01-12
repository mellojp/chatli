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
	s := SystemStyle.Render("user@terminal:~$ ") + HighlightTitleStyle.Render("login") + "\n\n"

	if m.SuccessMsg != "" {
		s += SuccessStyle.Render(m.SuccessMsg) + "\n\n"
	}

	var userPrefix, passPrefix, userLabel, passLabel string

	// Configuração Visual do Username
	if m.InputIndex == 0 {
		userPrefix = "> "
		userLabel = ActiveLabelStyle.Render("Username:")
	} else {
		userPrefix = "  "
		userLabel = InactiveLabelStyle.Render("Username:")
	}

	// Configuração Visual do Password
	if m.InputIndex == 1 {
		passPrefix = "> "
		passLabel = ActiveLabelStyle.Render("Password:")
	} else {
		passPrefix = "  "
		passLabel = InactiveLabelStyle.Render("Password:")
	}

	// Montagem das linhas com alinhamento manual
	// %-10s não é usado no label para não bagunçar com os códigos de cor ANSI do Lipgloss
	// Então alinhamos visualmente com espaços fixos se necessário, ou deixamos fluido.
	// Vamos usar um layout simples: PREFIX LABEL INPUT
	
	s += fmt.Sprintf("%s%s %s\n", userPrefix, userLabel, m.UsernameInput.View())
	s += fmt.Sprintf("%s%s %s\n\n", passPrefix, passLabel, m.PasswordInput.View())

	s += lipgloss.PlaceHorizontal(m.WindowWidth, lipgloss.Center, HelpStyle.Render("[tab/arrows] switch fields | [enter] login | [esc] register new account"))

	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}

	return renderAsciiHeader(m) + s
}

func RenderRegister(m *Model) string {
	s := SystemStyle.Render("user@terminal:~$ ") + HighlightTitleStyle.Render("register") + "\n\n"

	var userPrefix, passPrefix, userLabel, passLabel string

	if m.InputIndex == 0 {
		userPrefix = "> "
		userLabel = ActiveLabelStyle.Render("Username:")
	} else {
		userPrefix = "  "
		userLabel = InactiveLabelStyle.Render("Username:")
	}

	if m.InputIndex == 1 {
		passPrefix = "> "
		passLabel = ActiveLabelStyle.Render("Password:")
	} else {
		passPrefix = "  "
		passLabel = InactiveLabelStyle.Render("Password:")
	}

	s += fmt.Sprintf("%s%s %s\n", userPrefix, userLabel, m.UsernameInput.View())
	s += fmt.Sprintf("%s%s %s\n\n", passPrefix, passLabel, m.PasswordInput.View())

	s += lipgloss.PlaceHorizontal(m.WindowWidth, lipgloss.Center, HelpStyle.Render("[tab/arrows] switch fields | [enter] create account | [esc] back to login"))

	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	
	return renderAsciiHeader(m) + s
}

func RenderRoomsList(m *Model) string {
	s := SystemStyle.Render(fmt.Sprintf("%s@terminal:~/chatli/rooms$ ls -lh", m.Session.Username)) + "\n\n"

	// Definição de Larguras
	width := m.WindowWidth
	if width < 0 {
		width = 0
	}
	
	// Layout: [Prefix 2] [Name Dynamic] [Gap 1] [Date 15] [Gap 1] [ID 36]
	// Total Fixed = 2 + 1 + 15 + 1 + 36 = 55 chars
	// Name Width = WindowWidth - 55
	
	fixedWidth := 55
	nameWidth := width - fixedWidth
	if nameWidth < 10 {
		nameWidth = 10
	}

	// Cabeçalho
	// Usamos estilos para forçar largura
	nameHeader := lipgloss.NewStyle().Width(nameWidth).Render("NAME")
	dateHeader := lipgloss.NewStyle().Width(15).Render("CREATED")
	idHeader := lipgloss.NewStyle().Width(36).Render("ID")
	
	headerStr := fmt.Sprintf("  %s %s %s", nameHeader, dateHeader, idHeader)
	s += ListHeaderStyle.Width(width).Render(headerStr) + "\n"

	if len(m.Session.JoinedRooms) == 0 {
		s += HelpStyle.Render("\n  (empty directory - use 'n' to create a room)") + "\n"
	}

	for i, room := range m.Session.JoinedRooms {
		dateStr := room.CreatedAt.Format("02/01 15:04")
		
		// Trunca nome se necessário (embora lipgloss oculte, é bom cortar)
		name := room.Name
		if len(name) > nameWidth {
			name = name[:nameWidth-1] + "…"
		}

		// Renderiza colunas
		colName := lipgloss.NewStyle().Width(nameWidth).Render(name)
		colDate := lipgloss.NewStyle().Width(15).Render(dateStr)
		colId := lipgloss.NewStyle().Width(36).Render(room.Id)

		lineContent := fmt.Sprintf("%s %s %s", colName, colDate, colId)
		
		var renderedLine string

		if i == m.Cursor {
			// Item selecionado
			renderedLine = ListSelectedRowStyle.Copy().Width(width).Render("> " + lineContent)
		} else {
			// Item normal
			renderedLine = ListNormalRowStyle.Copy().Width(width).Render("  " + lineContent)
		}
		s += renderedLine + "\n"
	}

	// Footer de ajuda (Centralizado)
	helpText := "[up/down] nav | [n] new room | [e] enter room id | [enter] select | [esc] logout"
	s += "\n" + lipgloss.PlaceHorizontal(width, lipgloss.Center, HelpStyle.Render(helpText))
	
	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	
	return renderAsciiHeader(m) + s
}

func RenderCreateRoom(m *Model) string {
	s := SystemStyle.Render(fmt.Sprintf("%s@terminal:~/chatli/rooms$ mkdir", m.Session.Username)) + "\n\n"
	
	// Input único sempre focado
	prefix := "> "
	label := ActiveLabelStyle.Render("Room Name:")
	
	s += fmt.Sprintf("%s%s %s", prefix, label, m.GenericInput.View())

	s += "\n\n" + lipgloss.PlaceHorizontal(m.WindowWidth, lipgloss.Center, HelpStyle.Render("[enter] create | [esc] cancel"))

	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}
	
	return renderAsciiHeader(m) + s
}

func RenderJoinRoom(m *Model) string {
	s := SystemStyle.Render(fmt.Sprintf("%s@terminal:~/chatli/rooms$ join", m.Session.Username)) + "\n\n"
	
	// Input único sempre focado
	prefix := "> "
	label := ActiveLabelStyle.Render("Room ID:")
	
	s += fmt.Sprintf("%s%s %s", prefix, label, m.GenericInput.View())

	s += "\n\n" + lipgloss.PlaceHorizontal(m.WindowWidth, lipgloss.Center, HelpStyle.Render("[enter] join | [esc] cancel"))

	if m.ErrorMsg != "" {
		s += "\n\n" + ErrorStyle.Render("error: "+m.ErrorMsg)
	}

	return renderAsciiHeader(m) + s
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
