package gui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type helpPage struct {
	title    string
	sections []helpSection
}

type helpSection struct {
	heading string
	items   []helpItem
}

type helpItem struct {
	key  string
	desc string
}

// buildHelp documents the LIVE keys — every entry must have a handler in
// gui.go's Update, and every handled key should appear here (r1 finding 24).
func buildHelp() []helpPage {
	return []helpPage{
		{
			title: "Global",
			sections: []helpSection{
				{
					heading: "Navigation",
					items: []helpItem{
						{"Tab", "Next panel"},
						{"Shift+Tab", "Previous panel"},
						{"1-7", "Jump to panel"},
						{"j/k", "Scroll list"},
						{"[/]", "Switch tabs"},
					},
				},
				{
					heading: "Actions",
					items: []helpItem{
						{"R", "Refresh all data"},
						{"/", "Search packages"},
						{"C", "Toggle command log"},
						{"q", "Quit lazybrew"},
						{"?", "This help"},
					},
				},
			},
		},
		{
			title: "Formulae",
			sections: []helpSection{
				{
					items: []helpItem{
						{"i", "Install selected"},
						{"x", "Uninstall selected"},
						{"r", "Reinstall selected"},
						{"u", "Upgrade selected"},
						{"U", "Upgrade all"},
						{"p", "Pin/Unpin selected"},
						{"L", "Toggle leaves filter"},
						{"F", "Fetch selected"},
					},
				},
			},
		},
		{
			title: "Casks",
			sections: []helpSection{
				{
					items: []helpItem{
						{"i", "Install"},
						{"x", "Uninstall"},
						{"X", "Zap uninstall"},
						{"r", "Reinstall"},
						{"u", "Upgrade"},
						{"U", "Upgrade all"},
						{"F", "Fetch"},
					},
				},
			},
		},
		{
			title: "Taps",
			sections: []helpSection{
				{
					items: []helpItem{
						{"a", "Add tap"},
						{"x", "Remove tap"},
						{"t", "Trust config"},
						{"r", "Repair tap"},
					},
				},
			},
		},
		{
			title: "Services",
			sections: []helpSection{
				{
					items: []helpItem{
						{"s", "Start"},
						{"S", "Stop"},
						{"r", "Restart"},
						{"f", "Run (foreground)"},
						{"c", "Cleanup stale services"},
					},
				},
			},
		},
		{
			title: "Outdated",
			sections: []helpSection{
				{
					items: []helpItem{
						{"u", "Upgrade selected"},
						{"U", "Upgrade all"},
						{"Space", "Toggle selection"},
						{"a", "Select/deselect all"},
					},
				},
			},
		},
		{
			title: "Status",
			sections: []helpSection{
				{
					items: []helpItem{
						{"R", "Refresh all data"},
						{"c", "Run brew cleanup"},
						{"d", "Run brew doctor"},
						{"A", "Run brew autoremove"},
						{"B", "Brewfile actions"},
						{"v", "Vulnerability check"},
						{"m", "Missing dependencies"},
						{"[ ]", "Switch tabs"},
					},
				},
			},
		},
		{
			title: "Search",
			sections: []helpSection{
				{
					heading: "Actions",
					items: []helpItem{
						{"/", "Type search query"},
						{"Enter", "Execute search"},
						{"i", "Install selected result"},
						{"F", "Fetch selected result"},
					},
				},
			},
		},
	}
}

func renderHelp(width int) string {
	pages := buildHelp()
	colWidth := (width - 4) / 3

	var allRows []string
	var currentRow []string
	col := 0

	for _, page := range pages {
		title := style.PanelTitle.Render(page.title)
		content := title

		for _, section := range page.sections {
			if section.heading != "" {
				content += "\n" + style.SubtleText.Render(section.heading)
			}
			for _, item := range section.items {
				content += "\n" + style.HintKey.Render("  "+item.key) + "  " + style.HintDesc.Render(item.desc)
			}
		}

		currentRow = append(currentRow, lipgloss.NewStyle().Width(colWidth).Render(content))
		col++

		if col >= 3 {
			allRows = append(allRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
			currentRow = nil
			col = 0
		}
	}
	if len(currentRow) > 0 {
		for i := col; i < 3; i++ {
			currentRow = append(currentRow, "")
		}
		allRows = append(allRows, lipgloss.JoinHorizontal(lipgloss.Top, currentRow...))
	}

	return lipgloss.JoinVertical(lipgloss.Top, allRows...)
}
