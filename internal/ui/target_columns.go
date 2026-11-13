package ui

import "strings"

type targetColumnChoice struct {
	field string
	shown bool
}

func newTargetColumnsField(value string) formField {
	f := newField("ui.target_fields", "Target Columns", value, "up/down: select | space: show/hide | -/+: move up/down | tab: next")
	if strings.TrimSpace(value) == "" {
		value = "description, group, kind"
	}
	seen := map[string]bool{}
	order := strings.Split(value, ",")
	hasName := false
	for _, field := range order {
		if strings.TrimSpace(field) == "name" {
			hasName = true
		}
	}
	if !hasName {
		order = append([]string{"name"}, order...)
	}
	for _, raw := range order {
		field := strings.TrimSpace(raw)
		switch field {
		case "name", "description", "group", "kind":
			if !seen[field] {
				f.columns = append(f.columns, targetColumnChoice{field, true})
				seen[field] = true
			}
		}
	}
	for _, field := range []string{"name", "description", "group", "kind"} {
		if !seen[field] {
			f.columns = append(f.columns, targetColumnChoice{field, false})
		}
	}
	return f
}

func (f *formField) updateTargetColumns(key string) {
	// Bubble Tea models are passed by value; keep the previous model immutable.
	f.columns = append([]targetColumnChoice(nil), f.columns...)
	switch key {
	case "up":
		f.columnIndex = max(0, f.columnIndex-1)
	case "down":
		f.columnIndex = min(len(f.columns)-1, f.columnIndex+1)
	case " ", "space":
		column := &f.columns[f.columnIndex]
		if column.field != "name" {
			column.shown = !column.shown
		}
	case "-", "+", "=":
		delta := 1
		if key == "-" {
			delta = -1
		}
		next := f.columnIndex + delta
		if next >= 0 && next < len(f.columns) {
			f.columns[f.columnIndex], f.columns[next] = f.columns[next], f.columns[f.columnIndex]
			f.columnIndex = next
		}
	}
	fields := []string{}
	for _, column := range f.columns {
		if column.shown {
			fields = append(fields, column.field)
		}
	}
	f.input.SetValue(strings.Join(fields, ", "))
}

func (f formField) targetColumnsView(active bool) string {
	labels := map[string]string{"name": "Name", "description": "Description", "group": "Group", "kind": "Kind"}
	rows := []string{}
	for i, column := range f.columns {
		check := "[ ]"
		if column.shown {
			check = "[x]"
		}
		cursor := "  "
		if active && i == f.columnIndex {
			cursor = "> "
		}
		row := cursor + check + " " + labels[column.field]
		if column.field == "name" {
			row += " (always shown)"
		}
		if active && i == f.columnIndex {
			row = selectedStyle.Render(row)
		}
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}
