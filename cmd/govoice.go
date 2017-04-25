package main

import (
	"encoding/json"
	"fmt"
	"github.com/dannyvankooten/vat"
	"github.com/jung-kurt/gofpdf"
	"github.com/leekchan/accounting"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"flag"
	"github.com/nicksnyder/go-i18n/i18n"
	"io"
	"crypto/cipher"
	"encoding/hex"
	"crypto/aes"
	"crypto/rand"
)

var step float64 = 4

const (
	MARGIN_LEFT, MARGIN_TOP, MARGIN_RIGHT, MARGIN_BOTTOM float64 = 20,10,20,0

	BOX_FROM_X         float64 = MARGIN_LEFT
	BOX_FROM_Y         float64 = 28
	BOX_TO_X           float64 = MARGIN_LEFT
	BOX_TO_Y           float64 = 60
	BOX_INVOICE_Y      float64 = BOX_FROM_Y
	BOX_PAYMENT_X      float64 = MARGIN_LEFT
	BOX_PAYMENT_Y      float64 = 210
	BOX_NOTE_X         float64 = MARGIN_LEFT
	BOX_NOTE_Y         float64 = 240
	BOX_FULL_WIDTH     float64 = 0
	BOX_TABLE_X      float64 = MARGIN_LEFT
	BOX_TABLE_Y      float64 = 100
	DEFAULT_LINE_BREAK float64 = -1

	TEXT_ALIGN_LEFT       = "L"
	NO_FILL               = false
	FILL                  = true
	NO_BORDER             = "0"
	BORDER_BOTTOM         = "1B"
	BORDER_TOP 		= "T"
	TEXT_FONT             = "helvetica"
	FONT_STYLE_BOLD = "B"
	FONT_STYLE_NORMAL = ""
	TEXT_ALIGN_RIGHT_MID = "RM"
	TEXT_ALIGN_RIGHT_TOP = "RT"
	TEXT_ALIGN_RIGHT_BTM = "RB"
	TEXT_ALIGN_LEFT_MID = "LM"

	FONT_SIZE_H2 = 16
	LINE_HEIGHT_H2 = 16
	FONT_SIZE_H1 = 14
	LINE_HEIGHT_H1 = 6
	FONT_SIZE_NORMAL = 8
	LINE_HEIGHT_NORMAL = 3.7
	FONT_SIZE_SMALL = 6
	LINE_HEIGHT_SMALL = 3

	TABLE_COL1_W float64 = 60 // %
	TABLE_COL2_W,TABLE_COL3_W,TABLE_COL4_W float64= 13,13,13 // %
	TABLE_HEADER_H = 8
	TABLE_ROW_H = 6

	BLACK_R,BLACK_G, BLACK_B = 0,0,0
	WHITE_R,WHITE_G, WHITE_B = 255,255,255
)

func main() {
	// sample data

	// start writing invoice
	fmt.Println("Starting..")
	// program args
	var invoiceDescriptor = flag.String("d", "invoice.json", "load descriptor file, defaults invoice.json")
	var invoiceLang = flag.String("l", "en", "set the output language, default en")
	// parse args
	flag.Parse()
	// TODO read user input for password

	//read invoice descriptor
	invoice := readInvoiceDescriptor(*invoiceDescriptor)

	// load translations
	i18n.MustLoadTranslationFile(fmt.Sprintf("../i18n/%s.all.toml", *invoiceLang))
	T, _ := i18n.Tfunc(*invoiceLang)

	// create page
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(MARGIN_LEFT,MARGIN_TOP,MARGIN_RIGHT)

	// unicode font symbol (adding trailing space for better rendering)
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	invoice.Settings.CurrencySymbol = tr(invoice.Settings.CurrencySymbol + " ")

	// if Daylitime is enabled retrieve the content
	if invoice.Dailytime.Enabled {
		scanItemsFromDaily(&invoice)
	}

	// variables for boxes
	var boxTitle, boxContent string

	// add a page to the pdf
	pdf.AddPage()

	// get page size and margins
	w,_ := pdf.GetPageSize()
	ml,_,mr,_ := pdf.GetMargins()


	// title
	title := tr(strings.ToUpper(invoice.From.Name))
	pdf.SetFont(TEXT_FONT, FONT_STYLE_BOLD, FONT_SIZE_H1)
	pdf.CellFormat(BOX_FULL_WIDTH, LINE_HEIGHT_H1, title, BORDER_BOTTOM, 0, TEXT_ALIGN_RIGHT_MID, NO_FILL, 0, "")
	// calculate the title width to align the invoice box
	titleWidth := pdf.GetStringWidth(title);

	pdf.Ln(DEFAULT_LINE_BREAK)
	pdf.SetFont(TEXT_FONT, FONT_STYLE_NORMAL, FONT_SIZE_SMALL)
	pdf.CellFormat(BOX_FULL_WIDTH, LINE_HEIGHT_SMALL, invoice.From.Email, NO_BORDER, 0, TEXT_ALIGN_RIGHT_MID, NO_FILL, 0, "")

	pdf.SetFont(TEXT_FONT, FONT_STYLE_NORMAL, FONT_SIZE_NORMAL)

	// invoice data
	boxTitle = T("invoice")
	//boxContent = fmt.Sprintf("%s %s\n%s %s\n%s %s", "N.", invoice.Invoice.Number, "Date:" ,invoice.Invoice.Date, "Due:", invoice.Invoice.Due)
	boxContent = T("invoice_data", invoice.Invoice)
	boxInvoiceX := w-(titleWidth+mr+2) // calculate the offest of the invoice x
	printBlock(pdf, boxInvoiceX, BOX_INVOICE_Y, boxTitle, boxContent)

	// write from header
	boxTitle = T("from")
	//boxContent = fmt.Sprintf("%s\n%s\n%s %s\n%s\nSteuernummer: %s\nVAT: %s\n", invoice.From.Name, invoice.From.Address, invoice.From.AreaCode, invoice.From.City, invoice.From.Country, invoice.From.Steuernummer, invoice.From.VatNumber)
	boxContent = T("sender", invoice.From)
	printBlock(pdf, BOX_FROM_X, BOX_FROM_Y, boxTitle, boxContent)

	// write to header
	boxTitle = T("to")
	//boxContent = fmt.Sprintf("%s\n%s\n%s %s\n%s\nSteuernummer: %s\nVAT: %s\n", invoice.To.Name, invoice.To.Address, invoice.To.AreaCode, invoice.To.City, invoice.To.Country, invoice.To.Steuernummer, invoice.To.VatNumber)
	boxContent = T("recipient", invoice.To)
	printBlock(pdf, BOX_TO_X, BOX_TO_Y, boxTitle, boxContent)

	// TABLE itmes
	pdf.SetXY(BOX_TABLE_X, BOX_TABLE_Y)
	var c1v, c2v, c3v, c4v string // cell values

	// calculate the column widths
	table_max_width := w-(ml+ml)

	c1w := table_max_width * (TABLE_COL1_W/100)
	c2w := table_max_width * (TABLE_COL2_W/100)
	c3w := table_max_width * (TABLE_COL3_W/100)
	c4w := table_max_width * (TABLE_COL4_W/100)


	// write headers
	pdf.SetTextColor(WHITE_R, WHITE_G, WHITE_B)
	pdf.SetFillColor(BLACK_R, BLACK_G, BLACK_B)

	c1v = T("desc")
	c2v = T("quantity")
	c3v = T("rate")
	c4v = T("cost")

	pdf.CellFormat(c1w, TABLE_HEADER_H, c1v, NO_BORDER, 0, TEXT_ALIGN_LEFT_MID, FILL, 0, "")
	pdf.CellFormat(c2w, TABLE_HEADER_H, c2v, NO_BORDER, 0, TEXT_ALIGN_LEFT_MID, FILL, 0, "")
	pdf.CellFormat(c3w, TABLE_HEADER_H, c3v, NO_BORDER, 0, TEXT_ALIGN_LEFT_MID, FILL, 0, "")
	pdf.CellFormat(c4w, TABLE_HEADER_H, c4v, NO_BORDER, 0, TEXT_ALIGN_LEFT_MID, FILL, 0, "")
	pdf.Ln(TABLE_HEADER_H)
	pdf.SetTextColor(BLACK_R, BLACK_G, BLACK_B)

	// keep the subtotal
	var subtotal float64 = 0
	ac := accounting.Accounting{Symbol: invoice.Settings.CurrencySymbol, Precision: 2}
	//
	log.Print(invoice.Items)
	for _, it := range *invoice.Items {
		log.Print(it)

		c1v = tr(it.Description)
		c2v = strconv.FormatFloat(it.Quantity, 'f', 2, 64)
		// check if there is a per item rate
		itemPrice := invoice.Settings.BaseItemPrice
		if it.ItemPrice > 0{
			itemPrice = it.ItemPrice
		}
		// update sums
		cost := it.Quantity * itemPrice
		subtotal += cost
		// print column 3 and 4
		c3v = ac.FormatMoney(itemPrice)
		c4v = ac.FormatMoney(cost)

		pdf.CellFormat(c1w, TABLE_ROW_H, c1v, BORDER_BOTTOM, 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
		pdf.CellFormat(c2w, TABLE_ROW_H, c2v, BORDER_BOTTOM, 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
		pdf.CellFormat(c3w, TABLE_ROW_H, c3v, BORDER_BOTTOM, 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
		pdf.CellFormat(c4w, TABLE_ROW_H, c4v, BORDER_BOTTOM, 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
		pdf.Ln(TABLE_ROW_H)
	}
	pdf.Ln(TABLE_ROW_H)
	// subtotal
	c1v, c2v, c3v, c4v  = T("subtotal"), "", "", ac.FormatMoney(subtotal)

	pdf.CellFormat(c1w, TABLE_ROW_H, c1v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c2w, TABLE_ROW_H, c2v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c3w, TABLE_ROW_H, c3v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c4w, TABLE_ROW_H, c4v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.Ln(TABLE_ROW_H)
	// vat
	var vatRate float64 = 0
	if invoice.Settings.VatEnabled {
		c, _ := vat.GetCountryRates("NL")
		v, _ := c.GetRate("standard")
		vatRate = float64(v)
	}
	vatAmount := (vatRate / 100) * subtotal

	c1v, c2v, c3v, c4v = T("tax"), "", strconv.FormatFloat(vatRate, 'f', 2, 64)+" %", ac.FormatMoney(vatAmount)

	pdf.CellFormat(c1w, TABLE_ROW_H, c1v, "", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c2w, TABLE_ROW_H, c2v, "", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c3w, TABLE_ROW_H, c3v, "", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c4w, TABLE_ROW_H, c4v, "", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.Ln(TABLE_ROW_H)

	// total
	pdf.SetFont(TEXT_FONT, FONT_STYLE_BOLD, FONT_SIZE_NORMAL)
	total := subtotal + vatAmount

	c1v, c2v, c3v, c4v = T("total"), "", "", ac.FormatMoney(total)

	pdf.CellFormat(c1w, TABLE_ROW_H, c1v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c2w, TABLE_ROW_H, c2v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c3w, TABLE_ROW_H, c3v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")
	pdf.CellFormat(c4w, TABLE_ROW_H, c4v, "TB", 0, TEXT_ALIGN_LEFT_MID, NO_FILL, 0, "")

	// payment details
	boxTitle = T("payment_details")
	//boxContent = fmt.Sprintf("%s\n\nBank: %s\nIBAN: %s\nBIC: %s", invoice.PaymentDetails.AccountHolder, invoice.PaymentDetails.Bank, invoice.PaymentDetails.Iban, invoice.PaymentDetails.Bic)
	boxContent = T("payment_details_data", invoice.PaymentDetails)
	printBlock(pdf, BOX_PAYMENT_X, BOX_PAYMENT_Y, boxTitle, boxContent)

	// notes
	boxTitle = T("notes")
	boxContent = strings.Join(invoice.Notes, "\n")
	printBlock(pdf, BOX_NOTE_X, BOX_NOTE_Y, boxTitle, boxContent)

	// render pdf
	err := pdf.OutputFileAndClose(fmt.Sprintf("%s.pdf", invoice.Invoice.Number))
	if err != nil {
		log.Fatal("Error: ", err)
	}
	// disable dailytime if enabled and render the json
	invoice.Dailytime.Enabled = false
	doEncrypt := true // TODO read from settings

	writeInvoiceDescriptor(&invoice, invoice.Invoice.Number, doEncrypt)

	fmt.Println("Done!\n")

}

func printBlock(pdf *gofpdf.Fpdf, x, y float64, title, content string) {
	// write title
	pdf.SetXY(x, y)
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetFont(TEXT_FONT, FONT_STYLE_NORMAL, FONT_SIZE_H2)
	title = tr(title)
	pdf.MultiCell(BOX_FULL_WIDTH, LINE_HEIGHT_NORMAL, title, NO_BORDER, TEXT_ALIGN_LEFT, NO_FILL)

	// write content
	pdf.SetXY(x, y+ FONT_SIZE_H2 /4)
	pdf.SetFont(TEXT_FONT, FONT_STYLE_NORMAL, FONT_SIZE_NORMAL)
	content = tr(content)
	pdf.MultiCell(BOX_FULL_WIDTH, LINE_HEIGHT_NORMAL, content, NO_BORDER, TEXT_ALIGN_LEFT, NO_FILL)
}

// readInvoice parse the json file for an invoice
func readInvoiceDescriptor(path string) Invoice {

	rawJsonDescriptor, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if strings.HasSuffix(path, ".enc"){
		// file is encrypted
		rawJsonDescriptor = decrypterCFB("la mia password1", rawJsonDescriptor)

	}
	var i Invoice
	json.Unmarshal(rawJsonDescriptor, &i)
	return i
}

// writeInvoice write the invoice as descriptor
func writeInvoiceDescriptor(i *Invoice, name string, encrypt bool){
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil{
		ioutil.WriteFile(fmt.Sprintf("%s.json",name), content, os.FileMode(0660))
	}

	if(encrypt){
		encContent := encryptCFB("la mia password1", content)
		ioutil.WriteFile(fmt.Sprintf("%s.json.enc",name), encContent, os.FileMode(0660))
	}
}

func scanItemsFromDaily(i *Invoice) {
	dailyExportCommand := fmt.Sprintf(`tell application "Daily" to print json with report "summary" from (date("%s")) to (date("%s"))`, i.Dailytime.DateFrom, i.Dailytime.DateTo)
	//log.Println(dailyExportCommand)
	cmd := exec.Command("/usr/bin/osascript", "-e", dailyExportCommand)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	var dailyTimeAppItem []struct {
		Activity       string  `json:"activity"`
		DurationString string  `json:"durationString"`
		Percentage     float64 `json:"percentage"`
		Duration       int     `json:"duration"`
	}

	if err := json.NewDecoder(stdout).Decode(&dailyTimeAppItem); err != nil {
		log.Fatal("Error decoding json: ", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	for _, di := range dailyTimeAppItem {
		//log.Println(di.Activity, " - ", seconds2hours(&di.Duration), " (", di.DurationString, ")")
		for _, pr := range i.Dailytime.Projects {
			if pr.Name == di.Activity {
				i.PushItem(pr.ItemDescription, seconds2hours(&di.Duration))
			}
		}
	}
}

func seconds2hours(s *int) float64 {
	return (float64(*s) / 60) / 60
}

//Invoice contains all the information to generate an invoice
type Invoice struct {
	From struct {
		Name         string `json:"name"`
		Address      string `json:"address"`
		City         string `json:"city"`
		AreaCode     string `json:"area_code"`
		Country      string `json:"country"`
		TaxId        string `json:"tax_id"`
		VatNumber    string `json:"vat_number"`
		Email        string `json:"email"`
	} `json:"from"`
	To struct {
		Name         string `json:"name"`
		Address      string `json:"address"`
		City         string `json:"city"`
		AreaCode     string `json:"area_code"`
		Country      string `json:"country"`
		TaxId 	     string `json:"tax_id"`
		VatNumber    string `json:"vat_number"`
	} `json:"to"`
	PaymentDetails struct {
		AccountHolder string `json:"account_holder"`
		Bank   string `json:"account_bank"`
		Iban   string `json:"account_iban"`
		Bic    string `json:"account_bic"`
	} `json:"payment_details"`
	Invoice struct {
		Number string `json:"number"`
		Date   string `json:"date"`
		Due    string `json:"due"`
	} `json:"invoice"`
	Settings struct {
		BaseItemPrice  float64 `json:"base_item_price"`
		VatEnabled     bool    `json:"vat_enabled"`
		VatCountryCode string  `json:"vat_country_code"`
		CurrencySymbol string  `json:"currency_symbol"`
		QuantitySymbol string  `json:"quantity_symbol"`
	} `json:"settings"`
	Dailytime struct {
		Enabled  bool   `json:"enabled"`
		DateFrom string `json:"date_from,omitempty"`
		DateTo   string `json:"date_to",omitempty`
		Projects []struct {
			Name            string `json:"name"`
			ItemDescription string `json:"item_description"`
		} `json:"projects",omitempty`
	} `json:"dailytime"`
	Items *[]Item  `json:"items"`
	Notes []string `json:"notes"`
}

type Item struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	ItemPrice   float64 `json:"item_price,omitempty"`
}

func (i *Invoice) PushItem(description string, quantity float64) {
	*i.Items = append(*i.Items, Item{description, quantity, -1})
}


type Color struct {
	R, G, B int
}

func decrypterCFB(k string, data []byte)([]byte) {
	key := []byte(k)
	ciphertext, _ := hex.DecodeString(string(data))

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	fmt.Printf("%s", ciphertext)
	// Output: some plaintext
	return ciphertext
}

func encryptCFB(k string, plaintext []byte)([]byte) {
	key := []byte(k)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	cypertexthex := make([]byte, hex.EncodedLen(len(ciphertext)))
	hex.Encode(cypertexthex, ciphertext)
	return cypertexthex
}