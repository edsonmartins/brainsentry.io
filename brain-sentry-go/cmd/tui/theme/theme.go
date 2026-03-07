package theme

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha palette
var (
	Primary    = lipgloss.Color("#CBA6F7") // Mauve
	Secondary  = lipgloss.Color("#A6E3A1") // Green
	Accent     = lipgloss.Color("#F9E2AF") // Yellow
	Danger     = lipgloss.Color("#F38BA8") // Red
	Info       = lipgloss.Color("#89B4FA") // Blue
	Peach      = lipgloss.Color("#FAB387") // Peach
	Pink       = lipgloss.Color("#F5C2E7") // Pink
	Teal       = lipgloss.Color("#94E2D5") // Teal
	Sky        = lipgloss.Color("#89DCEB") // Sky
	Lavender   = lipgloss.Color("#B4BEFE") // Lavender

	Base       = lipgloss.Color("#1E1E2E") // Background
	Mantle     = lipgloss.Color("#181825") // Darker bg
	Crust      = lipgloss.Color("#11111B") // Darkest bg
	Surface0   = lipgloss.Color("#313244") // Surface
	Surface1   = lipgloss.Color("#45475A") // Surface alt
	Surface2   = lipgloss.Color("#585B70") // Surface alt2
	Overlay0   = lipgloss.Color("#6C7086") // Overlay
	Overlay1   = lipgloss.Color("#7F849C") // Overlay alt
	Subtext0   = lipgloss.Color("#A6ADC8") // Dim text
	Subtext1   = lipgloss.Color("#BAC2DE") // Dim text alt
	Text       = lipgloss.Color("#CDD6F4") // Main text
)

// Category colors
var CategoryColors = map[string]lipgloss.Color{
	"KNOWLEDGE":    Info,
	"INSIGHT":      Secondary,
	"DECISION":     Accent,
	"WARNING":      Danger,
	"PATTERN":      Primary,
	"ANTIPATTERN":  Pink,
	"BUG":          Peach,
	"OPTIMIZATION": Teal,
	"CONTEXT":      Lavender,
	"REFERENCE":    Sky,
	"ACTION":       Secondary,
	"DOMAIN":       Primary,
	"INTEGRATION":  Sky,
}

// Importance colors
var ImportanceColors = map[string]lipgloss.Color{
	"CRITICAL":  Danger,
	"IMPORTANT": Accent,
	"MINOR":     Secondary,
}

// Status colors
var StatusColors = map[string]lipgloss.Color{
	"ACTIVE":    Secondary,
	"PAUSED":    Accent,
	"COMPLETED": Overlay0,
	"EXPIRED":   Danger,
}

// Reusable styles
var (
	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Subtext0).
			Bold(true)

	// Card with rounded border
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Surface1).
			Padding(1, 2)

	// Active card (highlighted border)
	ActiveCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// Tag badge
	TagStyle = lipgloss.NewStyle().
			Background(Surface0).
			Foreground(Secondary).
			Padding(0, 1)

	// Selected row
	SelectedStyle = lipgloss.NewStyle().
			Background(Surface0).
			Foreground(Text).
			Bold(true)

	// Hint text (keybindings at bottom)
	HintStyle = lipgloss.NewStyle().
			Foreground(Overlay0)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Dim text
	DimStyle = lipgloss.NewStyle().
			Foreground(Overlay0)

	// Value text
	ValueStyle = lipgloss.NewStyle().
			Foreground(Text)

	// Label text
	LabelStyle = lipgloss.NewStyle().
			Foreground(Subtext0).
			Bold(true)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Background(Surface0).
			Foreground(Text).
			Padding(0, 1)

	StatusBarAccentStyle = lipgloss.NewStyle().
				Background(Primary).
				Foreground(Crust).
				Padding(0, 1).
				Bold(true)

	// Section header
	SectionStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	// Box for content display
	ContentBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Surface1).
			Padding(1, 2).
			Foreground(Text)

	// Logo style
	LogoStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)
)

// ColorForCategory returns the color for a memory category.
func ColorForCategory(cat string) lipgloss.Color {
	if c, ok := CategoryColors[cat]; ok {
		return c
	}
	return Overlay0
}

// ColorForImportance returns the color for an importance level.
func ColorForImportance(imp string) lipgloss.Color {
	if c, ok := ImportanceColors[imp]; ok {
		return c
	}
	return Overlay0
}

// ColorForStatus returns the color for a session status.
func ColorForStatus(status string) lipgloss.Color {
	if c, ok := StatusColors[status]; ok {
		return c
	}
	return Overlay0
}

// StyledCategory returns a styled category string.
func StyledCategory(cat string) string {
	return lipgloss.NewStyle().
		Foreground(ColorForCategory(cat)).
		Render(cat)
}

// StyledImportance returns a styled importance string.
func StyledImportance(imp string) string {
	return lipgloss.NewStyle().
		Foreground(ColorForImportance(imp)).
		Bold(true).
		Render(imp)
}

// StyledStatus returns a styled status string.
func StyledStatus(status string) string {
	return lipgloss.NewStyle().
		Foreground(ColorForStatus(status)).
		Bold(true).
		Render(status)
}

// Badge renders a text with background color like a tag badge.
func Badge(text string, fg, bg lipgloss.Color) string {
	return lipgloss.NewStyle().
		Background(bg).
		Foreground(fg).
		Padding(0, 1).
		Bold(true).
		Render(text)
}
