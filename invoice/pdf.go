package invoice

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"gitlab.com/almost_cc/govoice/config"

	"github.com/davecgh/go-spew/spew"
	"github.com/jung-kurt/gofpdf"
	"github.com/leekchan/accounting"
	"github.com/olekukonko/tablewriter"
)

var step float64 = 4

const (
	boxFullWidth     float64 = 0
	defaultLineBreak float64 = -1

	textAlignLeft   = "L"
	noFill          = false
	fill            = true
	borderNone      = "0"
	borderBottom    = "1B"
	borderTopBottom = "TB"

	fontStyleBold     = "B"
	fontStyleNormal   = ""
	textAlignRightMid = "RM"
	textAlignRightTop = "RT"
	textAlignRightBtm = "RB"
	textAlignLeftMid  = "LM"

	blackR, blackG, blackB = 0, 0, 0
	whiteR, whiteG, whiteB = 255, 255, 255

	sectionInvoice  = "invoice"
	sectionFrom     = "from"
	sectionTo       = "to"
	sectionPayments = "payments"
	sectionNotes    = "notes"
	sectionDetails  = "details"
)

func applyTemplate(s *Section, data interface{}) (err error) {
	// workaround remove tab from template
	s.Template = strings.Replace(s.Template, "\t", "", -1)
	t, err := template.New(s.Title).Parse(s.Template)
	if err != nil {
		return
	}
	var out bytes.Buffer
	if err = t.Execute(&out, data); err != nil {
		return
	}
	s.Content = out.String()
	if config.DebugEnabled {
		log.Println("d: template", s.Template, "data: ", spew.Sdump(data), "output", s.Content)
	}
	return
}

func computeCoordinates(s *Section, margins *Margins) {
	if s.X <= 0 {
		s.X = 0
	}

	if s.Y <= 0 {
		s.Y = 0
	}

	s.X += margins.Left
	s.Y += margins.Top

}

func computeColors(rgb []int, defaultR, defaultG, defaultB int) (r, g, b int) {
	r, g, b = defaultR, defaultG, defaultB
	if len(rgb) != 3 {
		return
	}
	for _, v := range rgb {
		if v < 0 || v > 255 {
			return
		}
	}
	return rgb[0], rgb[1], rgb[2]
}

func RenderPDF(invoice *Invoice, pdfPath string, tpl *InvoiceTemplate) {

	// create page
	pdf := gofpdf.New(tpl.Page.Orientation, "mm", tpl.Page.Size, "")

	pdf.SetMargins(tpl.Page.Margins.Left,
		tpl.Page.Margins.Top,
		tpl.Page.Margins.Right)
	defer pdf.Close()
	// unicode font symbol (adding trailing space for better rendering)
	utf8 := pdf.UnicodeTranslatorFromDescriptor("")
	currencySymbol := utf8(invoice.Settings.CurrencySymbol + " ")

	// add a page to the pdf
	pdf.AddPage()
	// get page size and margins
	w, h := pdf.GetPageSize()
	ml, _, _, _ := pdf.GetMargins()
	// draw the background
	fr, fg, fb := pdf.GetFillColor()
	pdf.SetFillColor(computeColors(tpl.Page.BackgroundColor, whiteR, whiteG, whiteB))
	pdf.Rect(.0, .0, w, h, "FD")
	// restore the fill colors
	pdf.SetFillColor(fr, fg, fb)
	// set the font color
	pdf.SetTextColor(computeColors(tpl.Page.FontColor, blackR, blackG, blackB))

	// title
	title := utf8(strings.ToUpper(invoice.From.Name))
	pdf.SetFont(tpl.Page.Font.Family, fontStyleBold, tpl.Page.Font.SizeH1)
	pdf.CellFormat(boxFullWidth, tpl.Page.Font.LineHeightH1, title, borderBottom, 0, textAlignRightMid, noFill, 0, "")
	pdf.Ln(defaultLineBreak)
	pdf.SetFont(tpl.Page.Font.Family, fontStyleNormal, tpl.Page.Font.SizeSmall)
	pdf.CellFormat(boxFullWidth, tpl.Page.Font.LineHeightSmall, invoice.From.Email, borderNone, 0, textAlignRightMid, noFill, 0, "")

	pdf.SetFont(tpl.Page.Font.Family, fontStyleNormal, tpl.Page.Font.SizeNormal)

	var section Section
	// invoice data
	section = tpl.Sections[sectionInvoice]
	applyTemplate(&section, invoice.Invoice)
	renderBlock(pdf, &section, &tpl.Page)

	// write from header
	section = tpl.Sections[sectionFrom]
	applyTemplate(&section, invoice.From)
	renderBlock(pdf, &section, &tpl.Page)

	// write to header
	section = tpl.Sections[sectionTo]
	applyTemplate(&section, invoice.To)
	renderBlock(pdf, &section, &tpl.Page)

	// TABLE itmes
	section = tpl.Sections[sectionDetails]
	renderBlock(pdf, &section, &tpl.Page)

	var c1v, c2v, c3v, c4v string // cell values

	// print table in console
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	// calculate the column widths
	tableMaxWidth := w - (section.X + ml)

	c1w := tableMaxWidth * (tpl.Page.Table.Col1W / 100)
	c2w := tableMaxWidth * (tpl.Page.Table.Col2W / 100)
	c3w := tableMaxWidth * (tpl.Page.Table.Col3W / 100)
	c4w := tableMaxWidth * (tpl.Page.Table.Col4W / 100)
	// create th row styles
	headerRowStyle := RowStyle{[]float64{c1w, c2w, c3w, c4w}, tpl.Page.Table.HeadHeight, borderNone, textAlignLeftMid, fill}
	normalRowStyle := RowStyle{[]float64{c1w, c2w, c3w, c4w}, tpl.Page.Table.RowHeight, borderBottom, textAlignLeftMid, noFill}

	// write headers
	tr, tg, tb := pdf.GetTextColor()
	fr, fg, fb = pdf.GetFillColor()

	pdf.SetTextColor(computeColors(tpl.Page.Table.HeaderFontColor, tr, tg, tb))
	pdf.SetFillColor(computeColors(tpl.Page.Table.HeaderBackgroundColor, fr, fg, fb))

	data := tpl.Page.Table.Header
	// table header console
	table.SetHeader(data)
	renderRow(pdf, &section, &headerRowStyle, data)

	// restore the colors
	pdf.SetTextColor(tr, tg, tb)
	pdf.SetFillColor(fr, fg, fb)

	// keep the subtotal
	ac := accounting.Accounting{Symbol: currencySymbol, Precision: 2}

	//  log.Print(invoice.Items)
	if invoice.Items == nil {
		items := []Item{}
		invoice.Items = &items
	}

	for _, it := range *invoice.Items {
		//log.Print(it)
		c1v = utf8(it.Description)
		c2v = utf8(it.FormatQuantity(invoice.Settings.ItemsQuantitySymbol, invoice.Settings.RoundQuantity))

		// get the price and the cost of the item
		// price can be global or per item
		itemPrice, itemCost := it.GetCost(&invoice.Settings.ItemsPrice, &invoice.Settings.RoundQuantity)
		// print column 3 and 4
		c3v = ac.FormatMoney(itemPrice)
		c4v = ac.FormatMoney(itemCost)

		data = []string{c1v, c2v, c3v, c4v}
		// append data for the console output
		table.Append(data)
		// render pdf row
		renderRow(pdf, &section, &normalRowStyle, data)
	}
	pdf.Ln(tpl.Page.Table.RowHeight)
	// total and subtotal
	subtotal, total := invoice.GetTotals()

	// subtotal
	c1v, c2v, c3v, c4v = tpl.Page.Table.LabelSubtotal, "", "", ac.FormatMoney(subtotal)
	data = []string{c1v, c2v, c3v, c4v}
	// append data for the console output
	table.Append(data)
	// render pdf
	pdf.SetX(section.X)
	renderRow(pdf, &section, &normalRowStyle, data)

	// vat
	c1v, c2v, c3v, c4v = tpl.Page.Table.LabelTax, "", strconv.FormatFloat(invoice.Settings.VatRate, 'f', 2, 64)+" %", ac.FormatMoney(total-subtotal)
	data = []string{c1v, c2v, c3v, c4v}
	// append data for the console output
	table.Append([]string{c1v, c2v, c3v, c4v})
	renderRow(pdf, &section, &normalRowStyle, data)

	// total
	pdf.SetFont(tpl.Page.Font.Family, fontStyleBold, tpl.Page.Font.SizeNormal)

	c1v, c2v, c3v, c4v = tpl.Page.Table.LabelTotal, "", "", ac.FormatMoney(total)
	data = []string{c1v, c2v, c3v, c4v}
	// append data for the console output
	table.Append(data)
	renderRow(pdf, &section, &normalRowStyle, data)

	// render console table
	table.Render()

	// payment details
	section = tpl.Sections[sectionPayments]
	applyTemplate(&section, invoice.PaymentDetails)
	renderBlock(pdf, &section, &tpl.Page)

	// notes
	section = tpl.Sections[sectionNotes]
	applyTemplate(&section, invoice.Notes)
	renderBlock(pdf, &section, &tpl.Page)

	// render pdf
	err := pdf.OutputFileAndClose(pdfPath)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

// renderBlock renders a block in the pdf
func renderBlock(pdf *gofpdf.Fpdf, s *Section, page *Page) {
	// adjust x/y
	computeCoordinates(s, &page.Margins)
	// copy the x,y values
	x, y := s.X, s.Y
	// this is necessary to handle unicode string
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	//
	pdf.SetXY(x, y)
	if len(s.Title) > 0 {
		// write title
		pdf.SetFont(page.Font.Family, fontStyleNormal, page.Font.SizeH2)
		pdf.MultiCell(boxFullWidth, page.Font.LineHeightH2, tr(s.Title), borderNone, textAlignLeft, noFill)
		// update x,y
		x, y = s.X, s.Y+page.Font.LineHeightH2
	}
	// reset the font to normal
	pdf.SetFont(page.Font.Family, fontStyleNormal, page.Font.SizeNormal)

	if len(s.Content) > 0 {
		// write content
		pdf.SetXY(x, y)
		pdf.MultiCell(boxFullWidth, page.Font.LineHeightNormal, tr(s.Content), borderNone, textAlignLeft, noFill)
	}
}

func renderRow(pdf *gofpdf.Fpdf, s *Section, rs *RowStyle, colValues []string) {
	// there are more columns than data
	pdf.SetX(s.X)
	for i, w := range rs.ColWidths {
		pdf.CellFormat(w, rs.Height, colValues[i], rs.Border, 0, rs.TextAlign, rs.Fill, 0, "")
	}
	pdf.Ln(rs.Height)
}

type RowStyle struct {
	ColWidths []float64
	Height    float64
	Border    string
	TextAlign string
	Fill      bool
}
