package invoice

import (
	"github.com/jung-kurt/gofpdf"
	"github.com/leekchan/accounting"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/olekukonko/tablewriter"
	"log"
	"os"
	"strconv"
	"strings"
)

var step float64 = 4

const (
	BOX_FULL_WIDTH     float64 = 0
	DEFAULT_LINE_BREAK float64 = -1

	TEXT_ALIGN_LEFT   = "L"
	NO_FILL           = false
	FILL              = true
	NO_BORDER         = "0"
	BORDER_BOTTOM     = "1B"
	BORDER_TOP_BOTTOM = "TB"

	FONT_STYLE_BOLD      = "B"
	FONT_STYLE_NORMAL    = ""
	TEXT_ALIGN_RIGHT_MID = "RM"
	TEXT_ALIGN_RIGHT_TOP = "RT"
	TEXT_ALIGN_RIGHT_BTM = "RB"
	TEXT_ALIGN_LEFT_MID  = "LM"

	BLACK_R, BLACK_G, BLACK_B = 0, 0, 0
	WHITE_R, WHITE_G, WHITE_B = 255, 255, 255
)

func RenderPDF(invoice *Invoice, layout *Layout, pdfPath *string, T i18n.TranslateFunc) {

	// compute layout defaults
	l := layoutComputeDefaults(*layout)

	// create page
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(l.Style.Margins.Left, l.Style.Margins.Top, l.Style.Margins.Right)
	defer pdf.Close()
	// unicode font symbol (adding trailing space for better rendering)
	utf8 := pdf.UnicodeTranslatorFromDescriptor("")
	currencySymbol := utf8(invoice.Settings.CurrencySymbol + " ")

	// variables for boxes
	var boxTitle, boxContent string

	// add a page to the pdf
	pdf.AddPage()

	// get page size and margins
	w, _ := pdf.GetPageSize()
	ml, _, _, _ := pdf.GetMargins()

	// title
	title := utf8(strings.ToUpper(invoice.From.Name))
	pdf.SetFont(l.Style.FontFamily, FONT_STYLE_BOLD, l.Style.FontSizeH1)
	pdf.CellFormat(BOX_FULL_WIDTH, l.Style.LineHeightH1, title, BORDER_BOTTOM, 0, TEXT_ALIGN_RIGHT_MID, NO_FILL, 0, "")

	pdf.Ln(DEFAULT_LINE_BREAK)
	pdf.SetFont(l.Style.FontFamily, FONT_STYLE_NORMAL, l.Style.FontSizeSmall)
	pdf.CellFormat(BOX_FULL_WIDTH, l.Style.FontSizeNormal, invoice.From.Email, NO_BORDER, 0, TEXT_ALIGN_RIGHT_MID, NO_FILL, 0, "")

	pdf.SetFont(l.Style.FontFamily, FONT_STYLE_NORMAL, l.Style.FontSizeNormal)

	// invoice data
	boxTitle = T("invoice")
	//boxContent = fmt.Sprintf("%s %s\n%s %s\n%s %s", "N.", invoice.Invoice.Number, "Date:" ,invoice.Invoice.Date, "Due:", invoice.Invoice.Due)
	boxContent = T("invoice_data", invoice.Invoice)
	renderBlock(pdf, &l.Invoice, boxTitle, boxContent, &l.Style)

	// write from header
	boxTitle = T("from")
	//boxContent = fmt.Sprintf("%s\n%s\n%s %s\n%s\nSteuernummer: %s\nVAT: %s\n", invoice.From.Name, invoice.From.Address, invoice.From.AreaCode, invoice.From.City, invoice.From.Country, invoice.From.Steuernummer, invoice.From.VatNumber)
	boxContent = T("sender", invoice.From)
	renderBlock(pdf, &l.From, boxTitle, boxContent, &l.Style)

	// write to header
	boxTitle = T("to")
	//boxContent = fmt.Sprintf("%s\n%s\n%s %s\n%s\nSteuernummer: %s\nVAT: %s\n", invoice.To.Name, invoice.To.Address, invoice.To.AreaCode, invoice.To.City, invoice.To.Country, invoice.To.Steuernummer, invoice.To.VatNumber)
	boxContent = T("recipient", invoice.To)
	renderBlock(pdf, &l.To, boxTitle, boxContent, &l.Style)

	// TABLE itmes
	pdf.SetXY(l.Items.Position.X, l.Items.Position.Y)
	var c1v, c2v, c3v, c4v string // cell values

	// print table in console
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	// calculate the column widths
	table_max_width := w - (ml + ml)

	c1w := table_max_width * (l.Style.TableCol1W / 100)
	c2w := table_max_width * (l.Style.TableCol2W / 100)
	c3w := table_max_width * (l.Style.TableCol3W / 100)
	c4w := table_max_width * (l.Style.TableCol4W / 100)
	// create th row styles
	headerRowStyle := RowStyle{[]float64{c1w, c2w, c3w, c4w}, l.Style.TableHeadHeight, NO_BORDER, TEXT_ALIGN_LEFT_MID, FILL}
	normalRowStyle := RowStyle{[]float64{c1w, c2w, c3w, c4w}, l.Style.TableRowHeight, BORDER_BOTTOM, TEXT_ALIGN_LEFT_MID, NO_FILL}

	// write headers
	pdf.SetTextColor(WHITE_R, WHITE_G, WHITE_B)
	pdf.SetFillColor(BLACK_R, BLACK_G, BLACK_B)

	c1v = T("desc")
	c2v = T("quantity")
	c3v = T("rate")
	c4v = T("cost")
	data := []string{c1v, c2v, c3v, c4v}
	// table header console
	table.SetHeader(data)
	renderRow(pdf, &headerRowStyle, data)

	pdf.SetTextColor(BLACK_R, BLACK_G, BLACK_B)

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
		c2v = strconv.FormatFloat(it.Quantity, 'f', 2, 64)
		// get the price and the cost of the item
		// price can be global or per item
		itemPrice, itemCost := it.GetCost(&invoice.Settings.BaseItemPrice)
		// print column 3 and 4
		c3v = ac.FormatMoney(itemPrice)
		c4v = ac.FormatMoney(itemCost)

		data = []string{c1v, c2v, c3v, c4v}
		// append data for the console output
		table.Append(data)
		// render pdf row
		renderRow(pdf, &normalRowStyle, data)

	}
	pdf.Ln(l.Style.TableRowHeight)
	// total and subtotal
	subtotal, total := invoice.GetTotals()

	// subtotal
	c1v, c2v, c3v, c4v = T("subtotal"), "", "", ac.FormatMoney(subtotal)
	data = []string{c1v, c2v, c3v, c4v}
	// append data for the console output
	table.Append(data)
	// render pdf
	renderRow(pdf, &normalRowStyle, data)

	// vat
	c1v, c2v, c3v, c4v = T("tax"), "", strconv.FormatFloat(invoice.Settings.VatRate, 'f', 2, 64)+" %", ac.FormatMoney(total-subtotal)
	data = []string{c1v, c2v, c3v, c4v}
	// append data for the console output
	table.Append([]string{c1v, c2v, c3v, c4v})
	renderRow(pdf, &normalRowStyle, data)

	// total
	pdf.SetFont(l.Style.FontFamily, FONT_STYLE_BOLD, l.Style.FontSizeNormal)

	c1v, c2v, c3v, c4v = T("total"), "", "", ac.FormatMoney(total)
	data = []string{c1v, c2v, c3v, c4v}
	// append data for the console output
	table.Append(data)
	renderRow(pdf, &normalRowStyle, data)

	// render console table
	table.Render()

	// payment details
	boxTitle = T("payment_details")
	//boxContent = fmt.Sprintf("%s\n\nBank: %s\nIBAN: %s\nBIC: %s", invoice.PaymentDetails.AccountHolder, invoice.PaymentDetails.Bank, invoice.PaymentDetails.Iban, invoice.PaymentDetails.Bic)
	boxContent = T("payment_details_data", invoice.PaymentDetails)
	renderBlock(pdf, &l.Payments, boxTitle, boxContent, &l.Style)

	// notes
	boxTitle = T("notes")
	boxContent = strings.Join(invoice.Notes, "\n")
	renderBlock(pdf, &l.Notes, boxTitle, boxContent, &l.Style)

	// render pdf
	err := pdf.OutputFileAndClose(*pdfPath)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

// layoutComputeDefaults All the X values that are < 0 will be defaulted to Margin.X
func layoutComputeDefaults(l Layout) Layout {
	if l.From.Position.X < 0 {
		l.From.Position.X = l.Style.Margins.Left
	}

	if l.To.Position.X < 0 {
		l.To.Position.X = l.Style.Margins.Left
	}

	if l.Notes.Position.X < 0 {
		l.Notes.Position.X = l.Style.Margins.Left
	}

	if l.Payments.Position.X < 0 {
		l.Payments.Position.X = l.Style.Margins.Left
	}

	if l.Invoice.Position.X < 0 {
		l.Invoice.Position.X = l.Style.Margins.Left
	}

	if l.Items.Position.X < 0 {
		l.Items.Position.X = l.Style.Margins.Left
	}
	return l
}

func renderBlock(pdf *gofpdf.Fpdf, b *Block, title, content string, s *Style) {
	// write title
	pdf.SetXY(b.Position.X, b.Position.Y)
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetFont(s.FontFamily, FONT_STYLE_NORMAL, s.FontSizeH2)
	title = tr(title)
	pdf.MultiCell(BOX_FULL_WIDTH, s.LineHeightNormal, title, NO_BORDER, TEXT_ALIGN_LEFT, NO_FILL)

	// write content
	pdf.SetXY(b.Position.X, b.Position.Y+s.FontSizeH2/4)
	pdf.SetFont(s.FontFamily, FONT_STYLE_NORMAL, s.FontSizeNormal)
	content = tr(content)
	pdf.MultiCell(BOX_FULL_WIDTH, s.LineHeightNormal, content, NO_BORDER, TEXT_ALIGN_LEFT, NO_FILL)
}

func renderRow(pdf *gofpdf.Fpdf, s *RowStyle, colValues []string) {
	// there are more columns than data
	for i, w := range s.ColWidths {
		pdf.CellFormat(w, s.Height, colValues[i], s.Border, 0, s.TextAlign, s.Fill, 0, "")
	}
	pdf.Ln(s.Height)
}

type RowStyle struct {
	ColWidths []float64
	Height    float64
	Border    string
	TextAlign string
	Fill      bool
}
