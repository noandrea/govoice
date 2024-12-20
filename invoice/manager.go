package invoice

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gitlab.com/almost_cc/govoice/config"
)

//Errors
var InvoiceDescriptorExists = errors.New("Invoice descriptor already exists")

//============== INVOICE ================

//Invoice contains all the information to generate an invoice
type Invoice struct {
	From           Recipient       `json:"from"`
	To             Recipient       `json:"to"`
	PaymentDetails BankCoordinates `json:"payment_details"`
	Invoice        InvoiceData     `json:"invoice"`
	Settings       InvoiceSettings `json:"settings"`
	Dailytime      Daily           `json:"dailytime"`
	Items          *[]Item         `json:"items"`
	Notes          []string        `json:"notes"`
}

// PushItem push an item to the list of the items of the invoice
func (i *Invoice) PushItem(description string, quantity, price float64, quantitySymbol string) {
	*i.Items = append(*i.Items, Item{description, quantity, price, quantitySymbol})
}

// DisableExtensions disable the extensions of the invoices
// TODO extenesions should be treathed as a list
func (i *Invoice) DisableExtensions() {
	i.Dailytime.Enabled = false
}

// GetTotals calculate and retrieve the subtotal (without tax) and the total (with tax)
// if the tax rate is 0 the subtotal and total are the same
func (i *Invoice) GetTotals() (float64, float64) {
	subtotal := 0.0
	for _, it := range *i.Items {
		_, itemCost := it.GetCost(&i.Settings.ItemsPrice, &i.Settings.RoundQuantity)
		subtotal += itemCost
	}
	total := subtotal
	if i.Settings.VatRate > 0 {
		total += total * (i.Settings.VatRate / 100)
	}
	return subtotal, total
}

type Daily struct {
	Enabled  bool           `json:"enabled"`
	DateFrom string         `json:"date_from,omitempty"`
	DateTo   string         `json:"date_to",omitempty`
	Projects []DailyProject `json:"projects",omitempty`
}

type DailyProject struct {
	Name            string  `json:"name"`
	ItemDescription string  `json:"item_description"`
	ItemPrice       float64 `json:"item_price,omitempty"`
}

type InvoiceSettings struct {
	ItemsPrice          float64 `json:"items_price"`
	ItemsQuantitySymbol string  `json:"items_quantity_symbol"`
	VatRate             float64 `json:"vat_rate"`
	CurrencySymbol      string  `json:"currency_symbol"`
	Language            string  `json:"lang"`
	DateInputFormat     string  `json:"date_format",omitempty`
	RoundQuantity       bool    `json:"round_quantity",omitempty`
}

type InvoiceData struct {
	Number string `json:"number"`
	Date   string `json:"date"`
	Due    string `json:"due"`
}

type BankCoordinates struct {
	AccountHolder string `json:"account_holder"`
	Bank          string `json:"account_bank"`
	Iban          string `json:"account_iban"`
	Bic           string `json:"account_bic"`
}

type Recipient struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	City      string `json:"city"`
	AreaCode  string `json:"area_code"`
	Country   string `json:"country"`
	TaxId     string `json:"tax_id"`
	VatNumber string `json:"vat_number"`
	Email     string `json:"email"`
}

type Item struct {
	Description    string  `json:"description"`
	Quantity       float64 `json:"quantity"`
	Price          float64 `json:"price,omitempty"`
	QuantitySymbol string  `json:"quantity_symbol,omitempty"`
}

//GetCost return the cost of an item, that is the ItemPrice multiplied the ItemQuantity.
//if the ItemPrice of the item is 0 then the global item price will be used.
// The function also rounds the quantity to the next .5 if it is specified in settings
func (i *Item) GetCost(basePrice *float64, roundQuantity *bool) (unitCost, cost float64) {
	qt := i.Quantity
	if *roundQuantity {
		qt = math.Ceil(i.Quantity*2) / 2
	}
	if i.Price > 0 {
		return i.Price, i.Price * qt
	}
	return *basePrice, *basePrice * qt
}

// FormatQuantity with a quantity symbol if present. it also rounds the quantity to the next .5
// if it specified in the settings
func (i *Item) FormatQuantity(quantitySymbol string, roundQuantity bool) string {

	// round quantity only if is requested
	adjQt := i.Quantity
	if roundQuantity {
		adjQt = math.Ceil(i.Quantity*2) / 2
	}
	qt := strconv.FormatFloat(adjQt, 'f', 2, 64)

	if i.QuantitySymbol != "" {
		quantitySymbol = i.QuantitySymbol
	}

	if quantitySymbol != "" {
		qt = fmt.Sprintf("%s %s", qt, i.QuantitySymbol)
	}
	return qt
}

// PreviewInvoice same as RenderInvoice but for previews
func PreviewInvoice(templatePath string) (invoiceNumber string, err error) {
	// check if master exists
	descriptorPath, exists := config.GetMasterPath()
	if !exists {
		// file not exists, search for the encrypted version
		err = errors.New("master descriptor not found")
		return
	}
	// read the master descriptor
	invoice, err := readInvoiceDescriptor(descriptorPath)
	if err != nil {
		return
	}
	// set the return invoice number
	invoiceNumber = invoice.Invoice.Number
	// load template
	template, err := readInvoiceTemplate(templatePath)

	// if Daylitime is enabled retrieve the content
	if invoice.Dailytime.Enabled {
		scanItemsFromDaily(&invoice)
	}
	// compute paths
	pdfPath, _ := config.GetInvoicePdfPath(config.PreviewFileName)
	RenderPDF(&invoice, pdfPath, &template)

	fmt.Println("pdf created at", pdfPath)
	return
}

//RenderInvoice render the master descriptor to a pdf file and create the encrypted descriptor of the invoice.
//The pdf and the descriptor are stored in the workspace folder in the format $INVOICE_NUMBER.pdf / $INVOICE_NUMBER.json.cfb
func RenderInvoice(password, templatePath string) (invoiceNumber string, err error) {
	// check if master exists
	descriptorPath, exists := config.GetMasterPath()
	if !exists {
		// file not exists, search for the encrypted version
		err = errors.New("master descriptor not found")
		return
	}

	// read the master descriptor
	invoice, err := readInvoiceDescriptor(descriptorPath)
	if err != nil {
		return
	}
	// set the return invoice number
	invoiceNumber = invoice.Invoice.Number

	// load template
	template, err := readInvoiceTemplate(templatePath)

	// if Daylitime is enabled retrieve the content
	if invoice.Dailytime.Enabled {
		scanItemsFromDaily(&invoice)
	}
	// compute paths
	pdfPath, _ := config.GetInvoicePdfPath(invoice.Invoice.Number)
	descrPath, descrExists := config.GetInvoiceJsonPath(invoice.Invoice.Number)

	// add invoice to the index
	if err = addToSearchIndex(&invoice); err != nil {
		return
	}

	// if the de
	if descrExists {
		reply := ReadUserInput(fmt.Sprint("invoice ", invoice.Invoice.Number, " already exists, overwrite? [yes/no] yes"))
		if reply != "" && reply != "yes" {
			err = InvoiceDescriptorExists
			return
		}
	}

	if strings.TrimSpace(invoice.Invoice.Number) == "" {
		err = errors.New("missing invoice number in master descriptor")
		return
	}

	RenderPDF(&invoice, pdfPath, &template)
	// disable extensions in invoice
	invoice.DisableExtensions()
	// copy the date format if using the global one
	if invoice.Settings.DateInputFormat == "" {
		invoice.Settings.DateInputFormat = config.Govoice.DateInputFormat
	}

	writeInvoiceDescriptorEncrypted(&invoice, descrPath, password)

	fmt.Println("encrypted descriptor created at", descrPath)
	fmt.Println("pdf created at", pdfPath)

	return
}

//RestoreInvoice restore the encrypted invoice descriptor into the master descriptor for editing.
//Overwrites the master descriptor without asking for confirmation.
func RestoreInvoice(invoiceNumber, password string) (err error) {
	// check if the invoice descriptor exists
	descriptorPath, exists := config.GetInvoiceJsonPath(invoiceNumber)
	if !exists {
		return errors.New(fmt.Sprint("Invoice ", invoiceNumber, " not found in ", descriptorPath))
	}

	// parse de invoice
	invoice, err := readInvoiceDescriptorEncrypted(descriptorPath, password)
	if err != nil {
		err = errors.New("invalid password")
		return
	}
	// dump it on master descriptor
	masterDescriptorPath, _ := config.GetMasterPath()
	err = writeJsonToFile(masterDescriptorPath, invoice)
	return
}
