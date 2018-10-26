package main

import (
	"fmt"
)

// AdocWriter is Asciidoc writer
type AdocWriter struct{}

func (w *AdocWriter) h1(txt string) string {
	return "= " + txt + "\n"
}

func (w *AdocWriter) h2(txt string) string {
	return "== " + txt + "\n"
}

func (w *AdocWriter) para(txt string) string {
	if txt == "" {
		return txt
	}
	return txt + "\n"
}

func (w *AdocWriter) tableHeader(fields []string) string {
	out := "|===\n"
	for _, f := range fields {
		out += "|" + f
	}
	return out + "\n"
}

func (w *AdocWriter) tableEnd() string {
	return "|===\n"
}

func (w *AdocWriter) tableRow(fields []string) string {
	out := ""
	for _, f := range fields {
		out += "|" + f + "\n"
	}
	return out
}

func generateMetricsDoc() {
	// Asciidoc writer
	writer := AdocWriter{}

	fmt.Println(writer.h1("Metrics Exported by Gluster Prometheus exporter"))
	for _, m := range metrics {
		fmt.Println(writer.h2(m.Name))
		desc := m.LongHelp
		if desc == "" {
			desc = m.Help
		}
		fmt.Println(writer.para(desc))
		if len(m.Labels) > 0 {
			fmt.Println(writer.tableHeader([]string{"Label", "Description"}))
			for _, lbl := range m.Labels {
				fmt.Println(writer.tableRow([]string{lbl.Name, lbl.Help}))
			}
			fmt.Println(writer.tableEnd())
		}
	}
}
