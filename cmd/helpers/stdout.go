package helpers

import (
	"os"

	"github.com/olekukonko/tablewriter"
)

type TableData struct {
	Header []string
	Data   [][]string
	Footer []string
}

func (td *TableData) AddRow(elements ...string) {
	td.Data = append(td.Data, elements)
}

func (td *TableData) SetHeader(titles ...string) {
	td.Header = titles
}

func (td *TableData) SetFooter(elements ...string) {
	td.Footer = elements
}

func RenderTable(dt *TableData) {
	// output results to console as a table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoFormatHeaders(false)
	table.SetHeader(dt.Header)
	table.SetAutoWrapText(false)
	table.AppendBulk(dt.Data)
	table.SetFooter(dt.Footer) // Add Footer
	// render the output
	table.Render()
}
