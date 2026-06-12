package gui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/thiago/lazybrew/internal/gui/style"
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
						{"p", "Pin/Unpin selected"},
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
						{"p", "Pin/Unpin"},
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
