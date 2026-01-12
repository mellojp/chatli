package ui

import (
	"github.com/charmbracelet/lipgloss"
)

var ColorMap = map[int]string{
	1: "2", 2: "10", 3: "34", 4: "35", 5: "40", 6: "41",
}

var AppStyle = lipgloss.NewStyle().
	Margin(0, 0).
	Padding(0, 0)

var SystemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("22"))

var InputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

var SenderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("22"))

var TimeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

var RoomTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))

var HighlightTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1eff00"))

var RoomIdStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

var UnselectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("22"))

var SelectedItemStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("10")).
	Bold(true).
	Border(lipgloss.NormalBorder(), false, false, false, true). // Adiciona um border para indicar foco
	BorderForeground(lipgloss.Color("10"))                      // Cor do border

var HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

var ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Vermelho para erro

var SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true) // Verde para sucesso

// Estilos para a Lista de Salas
var ListHeaderStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240")).
	Bold(true)

// Estilo Genérico de Foco (Invertido: Texto Preto em Fundo Verde)
var FocusedRowStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("232")). // Preto
	Background(lipgloss.Color("10")).  // Verde brilhante
	Bold(true)

// Estilo Genérico Normal (Texto Verde)
var NormalRowStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("10")) // Verde padrão

// Estilos para Formulários (Login/Registro/Inputs)
var ActiveLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("10")). // Verde brilhante
	Bold(true)

var InactiveLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240")) // Cinza escuro

// Mantendo compatibilidade com código anterior (alias)
var ListSelectedRowStyle = FocusedRowStyle
var ListNormalRowStyle = NormalRowStyle

func HashColor(username string, colors map[int]string) string {
	var hashCode int
	for _, v := range username {
		hashCode += int(v)
	}
	hashCode = (hashCode % len(colors)) + 1
	return colors[hashCode]
}
