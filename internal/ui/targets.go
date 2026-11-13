package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oddship/wg-tui/internal/api"
	cfgpkg "github.com/oddship/wg-tui/internal/config"
	sshpkg "github.com/oddship/wg-tui/internal/ssh"
)

// browsePaneHeights includes pane borders and reserves space for the chrome.
func (m Model) browsePaneHeights() (int, int) {
	height := m.height
	if height == 0 {
		height = 34
	}
	chrome := lipgloss.Height(m.viewTopChrome("wgt", m.browseTitleText(), m.browseHelpBindings().short))
	body := max(8, height-chrome-3)
	if m.cfg.UI.DetailsPane == "right" && m.width >= 72 {
		return body, body
	}
	details := min(10, max(6, body/3))
	return body - details, details
}

func targetCell(t api.Target, field string, recent bool) string {
	value := ""
	switch field {
	case "name":
		value = t.Name
	case "description":
		value = t.Description
	case "group":
		value = t.Group.Name
	case "kind":
		value = t.Kind
	}
	value = strings.Join(strings.Fields(value), " ")
	if field == "name" && recent {
		value += " " + recentTargetMarker
	}
	if value == "" {
		value = "-"
	}
	return value
}

// Columns share widths across all results, so paging does not shift their edges.
func renderTargetTable(targets []api.Target, width, selected int, recent func(string) bool, ui cfgpkg.UIConfig) []string {
	fields := []string{}
	metadata := ui.TargetFields
	if strings.TrimSpace(metadata) == "" {
		metadata = cfgpkg.Default().UI.TargetFields
	}
	if metadata != "none" {
		for _, raw := range strings.Split(metadata, ",") {
			field := strings.TrimSpace(raw)
			if field == "name" || field == "description" || field == "group" || field == "kind" {
				fields = append(fields, field)
			}
		}
	}
	hasName := false
	for _, field := range fields {
		if field == "name" {
			hasName = true
		}
	}
	if !hasName {
		fields = append([]string{"name"}, fields...)
	}
	labels := map[string]string{"name": "Name", "description": "Description", "group": "Group", "kind": "Kind"}
	minimum := map[string]int{"name": 12, "description": 11, "group": 8, "kind": 6}
	caps := map[string]int{"name": 48, "description": 48, "group": 24, "kind": 12}
	// On narrow panes hide trailing columns, always retaining Name.
	for len(fields) > 1 {
		needed := 2 + 3*(len(fields)-1)
		for _, field := range fields {
			needed += minimum[field]
		}
		if needed <= width {
			break
		}
		remove := len(fields) - 1
		if fields[remove] == "name" {
			remove--
		}
		fields = append(fields[:remove], fields[remove+1:]...)
	}
	widths := make([]int, len(fields))
	for i, field := range fields {
		widths[i] = max(lipgloss.Width(labels[field]), minimum[field])
		for _, target := range targets {
			widths[i] = max(widths[i], min(caps[field], lipgloss.Width(targetCell(target, field, recent(target.Name)))))
		}
	}
	available := max(0, width-2-3*(len(fields)-1))
	total := func() int {
		n := 0
		for _, w := range widths {
			n += w
		}
		return n
	}
	for total() > available {
		largest := 0
		for i := range widths {
			if widths[i] > widths[largest] {
				largest = i
			}
		}
		if widths[largest] == 0 {
			break
		}
		widths[largest]--
	}
	if extra := available - total(); extra > 0 {
		grow := 0
		for i, field := range fields {
			if field == "description" {
				grow = i
			}
		}
		widths[grow] += extra
	}
	render := func(t api.Target, header, active bool) string {
		cells := make([]string, len(fields))
		for i, field := range fields {
			value := labels[field]
			if !header {
				value = targetCell(t, field, recent(t.Name))
			}
			value = truncate(value, widths[i])
			cells[i] = value + strings.Repeat(" ", max(0, widths[i]-lipgloss.Width(value)))
			if !header && !active && field != "name" && ui.TargetContrast == "subtle" {
				cells[i] = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "250"}).Render(cells[i])
			}
		}
		prefix := "  "
		if active {
			prefix = "• "
		}
		line := truncate(prefix+strings.Join(cells, " │ "), max(0, width))
		if header {
			return headerStyle.Render(line)
		}
		if active {
			return selectedStyle.Render(line)
		}
		return line
	}
	rows := []string{render(api.Target{}, true, false)}
	for i, target := range targets {
		rows = append(rows, render(target, false, i == selected))
	}
	return rows
}

// stackedDetailRows keeps every metadata value available through scrolling.
func (m Model) stackedDetailRows(width int) []string {
	t, ok := m.selectedTarget()
	if !ok {
		return nil
	}
	values := []string{"Name: " + t.Name, "Kind: " + t.Kind, "Group: " + t.Group.Name}
	if t.ExternalHost != "" {
		values = append(values, "Host: "+t.ExternalHost)
	}
	if t.DefaultDatabaseName != "" {
		values = append(values, "DB: "+t.DefaultDatabaseName)
	}
	if t.Description != "" {
		values = append(values, "Description: "+t.Description)
	}
	return strings.Split(lipgloss.NewStyle().Width(max(1, width)).Render(strings.Join(values, "\n")), "\n")
}

// The command stays pinned while metadata scrolls inside the bottom pane.
func (m Model) stackedDetailView(width int) string {
	selected, ok := m.selectedTarget()
	if !ok {
		return headerStyle.Render("Details") + "\nNo target selected. Use up/down to choose."
	}
	_, height := m.browsePaneHeights()
	available := max(1, height-4)
	metadata := m.stackedDetailRows(width)
	offset := clamp(m.detailScroll, 0, max(0, len(metadata)-available))
	end := min(len(metadata), offset+available)
	heading := "Details"
	if len(metadata) > available {
		heading += fmt.Sprintf(" %d-%d/%d · alt+up/down scroll", offset+1, end, len(metadata))
	}
	rows := []string{headerStyle.Render(heading)}
	rows = append(rows, metadata[offset:end]...)
	bin, args := sshpkg.ShellCommand(m.cfg, selected.Name)
	rows = append(rows, truncate("SSH: "+commandText(bin, args), width))
	return strings.Join(rows, "\n")
}

func renderBoundedPanel(content string, width, height int) string {
	lines := strings.Split(content, "\n")
	lines = lines[:min(len(lines), max(1, height-2))]
	for i := range lines {
		lines[i] = truncate(lines[i], max(1, width-4))
	}
	return panelStyle.Width(max(1, width-2)).Height(max(1, height-2)).Render(strings.Join(lines, "\n"))
}
