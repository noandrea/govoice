package invoice

import (
	"errors"
	"fmt"
	"github.com/nicksnyder/go-i18n/i18n"
	"strconv"
	"strings"
	"math"
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
		_, itemCost := it.GetCost(&i.Settings.ItemsPrice)
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
// The function also rounds the quantity to the next .5
func (i *Item) GetCost(basePrice *float64) (float64, float64) {
	qt := math.Ceil(i.Quantity * 2) / 2
	if i.Price > 0 {
		return i.Price, i.Price * qt
	}
	return *basePrice, *basePrice * qt
}

// FormatQuantity with a quantity symbol if present. it also rounds the quantity to the next .5
func (i *Item) FormatQuantity(quantitySymbol string) string {
	adjQt := math.Ceil(i.Quantity * 2) / 2
	qt := strconv.FormatFloat(adjQt, 'f', 2, 64)

	if i.QuantitySymbol != "" {
		quantitySymbol = i.QuantitySymbol
	}

	if quantitySymbol != "" {
		qt = fmt.Sprintf("%s %s", qt, i.QuantitySymbol)
	}
	return qt
}

//RenderInvoice render the master descriptor to a pdf file and create the encrypted descriptor of the invoice.
//The pdf and the descriptor are stored in the workspace folder in the format $INVOICE_NUMBER.pdf / $INVOICE_NUMBER.json.cfb
func RenderInvoice(c *Config, password string) (string, error) {
	// check if master exists
	descriptorPath, exists := c.GetMasterPath()
	if !exists {
		// file not exists, search for the encrypted version
		return "", errors.New("master descriptor not found!")
	}

	// read the master descriptor
	var i Invoice
	err := readInvoiceDescriptor(&descriptorPath, &i)
	if err != nil {
		return "", err
	}

	// load translations
	i18n.MustLoadTranslationFile(GetI18nTranslationPath(i.Settings.Language))
	T, _ := i18n.Tfunc(i.Settings.Language)

	// if Daylitime is enabled retrieve the content
	if i.Dailytime.Enabled {
		scanItemsFromDaily(&i)
	}
	// compute paths
	pdfPath, _ := c.GetInvoicePdfPath(i.Invoice.Number)
	descrPath, descrExists := c.GetInvoiceJsonPath(i.Invoice.Number)

	// add invoice to the index
	if err := addToSearchIndex(c, &i); err != nil {
		return i.Invoice.Number, err
	}

	// if the de
	if descrExists {
		reply := ReadUserInput(fmt.Sprint("invoice ", i.Invoice.Number, " already exists, overwrite? [yes/no] yes"))
		if reply != "" && reply != "yes" {
			return "", InvoiceDescriptorExists
		}
	}

	if strings.TrimSpace(i.Invoice.Number) == "" {
		return "", errors.New("missing invoice number in master descriptor")
	}

	RenderPDF(&i, &c.Layout, &pdfPath, T)
	// disable extensions in invoice
	i.DisableExtensions()
	// copy the date format if using the global one
	if i.Settings.DateInputFormat == "" {
		i.Settings.DateInputFormat = c.DateInputFormat
	}

	writeInvoiceDescriptorEncrypted(&i, &descrPath, &password)

	fmt.Println("encrypted descriptor created at", descrPath)
	fmt.Println("pdf created at", pdfPath)

	return i.Invoice.Number, nil
}

//RestoreInvoice restore the encrypted invoice descriptor into the master descriptor for editing.
//Overwrites the master descriptor without asking for confirmation.
func RestoreInvoice(c *Config, invoiceNumber, password string) error {
	var i Invoice

	// check if the invoice descriptor exists
	descriptorPath, exists := c.GetInvoiceJsonPath(invoiceNumber)
	if !exists {
		return errors.New(fmt.Sprint("Invoice ", invoiceNumber, " not found in ", c.Workspace))
	}

	// parse de invoice
	err := readInvoiceDescriptorEncrypted(&descriptorPath, &i, &password)
	if err != nil {
		return errors.New("invalid password")
	}
	// dump it on master descriptor
	masterDescriptorPath, _ := c.GetMasterPath()
	writeJsonToFile(masterDescriptorPath, i)
	return nil
}
