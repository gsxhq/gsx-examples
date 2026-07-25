package main

import "github.com/gsxhq/gsx-examples/streaming-partial/views"

// panelData returns deterministic, boring content for each panel name. Kept
// fixed rather than randomized so the ordering test can assert on it.
func panelData(name string) views.Panel {
	switch name {
	case "revenue":
		return views.Panel{
			Name:  "revenue",
			Title: "Revenue",
			Value: "$48,290",
			Note:  "+12% vs last month",
		}
	case "orders":
		return views.Panel{
			Name:  "orders",
			Title: "Recent orders",
			Rows: []views.Row{
				{Label: "#1042", Value: "$120.00"},
				{Label: "#1041", Value: "$64.50"},
				{Label: "#1040", Value: "$212.10"},
			},
		}
	case "products":
		return views.Panel{
			Name:  "products",
			Title: "Top products",
			Rows: []views.Row{
				{Label: "Widget A", Value: "312 sold"},
				{Label: "Widget B", Value: "204 sold"},
				{Label: "Widget C", Value: "98 sold"},
			},
		}
	default:
		return views.Panel{Name: name, Title: name}
	}
}
