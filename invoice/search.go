package invoice

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"gitlab.com/almost_cc/govoice/config"
)

// -------- types --------

// InvoiceQuery is the query object to search for invoices
type InvoiceQuery struct {
	Customer string
	Text     string
	AmountGE float64
	AmountLE float64
	DateFrom time.Time
	DateTo   time.Time
}

func (q *InvoiceQuery) String() string {
	var f []string
	if len(q.Customer) != 0 {
		f = append(f, fmt.Sprint("customer like '", q.Customer, "'"))
	}
	if len(q.Text) != 0 {
		f = append(f, fmt.Sprint("text like '", q.Text, "'"))
	}
	ddf, _ := time.Parse(config.QueryDateFormat, config.QueryDefaultDateFrom)
	if !isSameDate(ddf, q.DateFrom) {
		f = append(f, fmt.Sprint("date > ", q.DateFrom.Format(config.QueryDateFormat)))

	}
	if !isSameDate(time.Now(), q.DateTo) {
		f = append(f, fmt.Sprint("date > ", q.DateFrom.Format(config.QueryDateFormat)))
	}
	// add range on amount if necessary
	if q.AmountGE != config.QueryDefaultAmountGE {
		f = append(f, fmt.Sprint("amount > ", q.AmountGE))
	}
	// add range on amount if necessary
	if q.AmountLE != config.QueryDefaultAmountLE {
		f = append(f, fmt.Sprint("amount < ", q.AmountLE))
	}
	return strings.Join(f, " and ")
}

//InvoiceQuery is the object indexed by bleve
type InvoiceEntry struct {
	Number   string
	Customer string
	Amount   float64
	Date     time.Time
	Text     string
}

// -------- exported functions --------

//DefaultInvoiceQuery return the default invoice query object
func DefaultInvoiceQuery() InvoiceQuery {
	df, _ := time.Parse(config.QueryDateFormat, config.QueryDefaultDateFrom)
	return InvoiceQuery{
		Customer: config.QueryDefaultCustomer,
		Text:     config.QueryDefaultText,
		AmountGE: float64(config.QueryDefaultAmountGE),
		AmountLE: float64(config.QueryDefaultAmountLE),
		DateFrom: df,
		DateTo:   time.Now(),
	}
}

//CreateSearchIndex create a new empty search index in the GetSearchIndexFilePath folder
func CreateSearchIndex() error {

	// index
	indexPath, exists := config.GetSearchIndexFilePath()
	// if exists do nothing
	if exists {
		return nil
	}

	index, err := initBleveIndex(indexPath)
	if err != nil {
		return err
	}
	defer index.Close()
	return nil
}

//RebuildSearchIndex rebuild search index
func RebuildSearchIndex(password string) (int, time.Duration, error) {
	var counter int = -1
	var elapsed time.Duration
	// delete old index if exists
	indexPath, exists := config.GetSearchIndexFilePath()
	if exists {
		if err := os.RemoveAll(indexPath); err != nil {
			return counter, elapsed, err
		}
	}

	// create a new index
	index, err := initBleveIndex(indexPath)
	if err != nil {
		return counter, elapsed, errors.New("cannot open the search index")
	}
	defer index.Close()
	// track the time spended to create the search index
	start := time.Now()

	// create a new batch
	b := index.NewBatch()

	// scan the descriptor files
	files, _ := ioutil.ReadDir(config.Govoice.Workspace)

	counter = 0
	for _, f := range files {
		if path.Ext(f.Name()) == config.ExtCfb {
			descriptorPath := path.Join(config.Govoice.Workspace, f.Name())
			if invoice, err := readInvoiceDescriptorEncrypted(descriptorPath, password); err == nil {
				// build the IndexEntry
				amount, _ := invoice.GetTotals()
				df := dateFormatToLayout(invoice.Settings.DateInputFormat)
				invd, _ := time.Parse(df, invoice.Invoice.Date)
				// write the text in the bleve index
				var fulldescr strings.Builder
				for _, t := range *invoice.Items {
					fulldescr.WriteString(strings.ToLower(t.Description))
					fulldescr.WriteString(" ")
				}

				ie := InvoiceEntry{
					Number:   invoice.Invoice.Number,
					Customer: invoice.To.Name,
					Amount:   amount,
					Date:     invd,
					Text:     fulldescr.String(),
				}
				// add the invoice to the index
				b.Index(invoice.Invoice.Number, ie)

				counter++
				if counter%100 == 0 {
					index.Batch(b)
				}
			} else {
				fmt.Println("error decrypting ", f.Name(), ", the invoice will not be searchable")
			}
		}
	}
	index.Batch(b)
	elapsed = time.Since(start)
	docsCount, _ := index.DocCount()
	// return the numer of entries indexed
	return int(docsCount), elapsed, nil
}

//SearchInvoice search for an invoice using InvoiceQuery object
// returns the entries found (first 50), the total number of hits, the duration of the query and error
func SearchInvoice(q InvoiceQuery) (entries []InvoiceEntry, found uint64, elapsed time.Duration, amount float64, err error) {

	// index
	indexPath, exists := config.GetSearchIndexFilePath()
	// if the index exists delete it
	if !exists {
		err = errors.New("search index does not exists, run govoice index to create the index")
		return
	}
	// open  index
	index, err := bleve.Open(indexPath)
	if err != nil {
		err = errors.New("error opening the search inde")
		return
	}
	defer index.Close()

	query := bleve.NewConjunctionQuery()
	// add phrase match for customer
	if len(q.Customer) > 0 {
		subq := bleve.NewFuzzyQuery(strings.ToLower(q.Customer))
		subq.Fuzziness = 2
		subq.SetField(config.FieldCustomer)
		query.AddQuery(subq)
	}
	// add phrase match for text
	if len(q.Text) > 0 {
		subq := bleve.NewFuzzyQuery(strings.ToLower(q.Text))
		subq.Fuzziness = 4
		subq.SetField(config.FieldText)
		query.AddQuery(subq)
	}
	// add range on amount if necessary
	if q.AmountLE != config.QueryDefaultAmountLE || q.AmountGE != config.QueryDefaultAmountGE {
		subq := bleve.NewNumericRangeQuery(&q.AmountGE, &q.AmountLE)
		subq.SetField(config.FieldAmount)
		query.AddQuery(subq)
	}
	// add range query on date
	ddf, _ := time.Parse(config.QueryDateFormat, config.QueryDefaultDateFrom)
	if !isSameDate(ddf, q.DateFrom) || !isSameDate(time.Now(), q.DateTo) {
		subq := bleve.NewDateRangeQuery(q.DateFrom, q.DateTo)
		subq.SetField(config.FieldDate)
		query.AddQuery(subq)
	}

	if err = query.Validate(); err != nil {
		err = errors.New("invalid query")
		return
	}

	search := bleve.NewSearchRequest(query)
	search.Fields = []string{config.FieldNumber, config.FieldCustomer, config.FieldAmount, config.FieldDate}
	search.SortBy([]string{"-" + config.FieldDate, config.FieldNumber})
	search.Size = 50
	results, err := index.Search(search)
	if err != nil {
		err = errors.New("error running the search")
		return
	}

	// record also the total amount
	for _, res := range results.Hits {

		d, _ := time.Parse(time.RFC3339, res.Fields[config.FieldDate].(string))
		amount += res.Fields[config.FieldAmount].(float64)
		ie := InvoiceEntry{
			Number:   res.Fields[config.FieldNumber].(string),
			Customer: res.Fields[config.FieldCustomer].(string),
			Amount:   res.Fields[config.FieldAmount].(float64),
			Date:     d,
		}
		entries = append(entries, ie)
	}
	found = results.Total
	elapsed = results.Took
	return
}

// -------- private methods ---------

//initBleveIndex initialize the bleve index and mappings
func initBleveIndex(dbPath string) (bleve.Index, error) {
	var index bleve.Index
	var err error
	// open a new index
	mapping := bleve.NewIndexMapping()

	dm := bleve.NewDocumentMapping()
	// text mapping for the invoce number
	inm := bleve.NewTextFieldMapping()
	dm.AddFieldMappingsAt(config.FieldNumber, inm)
	// text mapping for the customer name
	cfm := bleve.NewTextFieldMapping()
	dm.AddFieldMappingsAt(config.FieldCustomer, cfm)
	// text mapping for the invoice text
	cft := bleve.NewTextFieldMapping()
	dm.AddFieldMappingsAt(config.FieldText, cft)
	// numeric mapping for the subtotal
	afm := bleve.NewNumericFieldMapping()
	dm.AddFieldMappingsAt(config.FieldAmount, afm)
	// numeric mapping for the date
	dfm := bleve.NewDateTimeFieldMapping()
	dm.AddFieldMappingsAt(config.FieldDate, dfm)
	// add document mapping
	mapping.AddDocumentMapping("invoice", dm)

	// create a new index
	index, err = bleve.New(dbPath, mapping)
	if err != nil {
		err = errors.New("cannot open or create the search index")
		return index, err
	}
	return index, err
}

//addToSearchIndex add an invoice to the existing search index
func addToSearchIndex(i *Invoice) error {
	// index
	indexPath, exists := config.GetSearchIndexFilePath()
	if !exists {
		return errors.New("index does not exists")
	}

	// open the index
	index, err := bleve.Open(indexPath)
	if err != nil {
		return errors.New("cannot open the search index")
	}
	defer index.Close()
	// parse the date format
	df := dateFormatToLayout(i.Settings.DateInputFormat)
	// parse the time
	date, err := time.Parse(df, i.Invoice.Date)
	if err != nil {
		return fmt.Errorf("date %s doesen't match the format %s", i.Invoice.Date, df)
	}
	// create the index entry
	amount, _ := i.GetTotals()
	ie := InvoiceEntry{
		Number:   i.Invoice.Number,
		Customer: i.To.Name,
		Amount:   amount,
		Date:     date,
	}
	// insert the entry in the index
	if err := index.Index(i.Invoice.Number, ie); err != nil {
		return errors.New(fmt.Sprint("error inserting", i.Invoice.Number, "to the search index"))
	}

	return nil

}

// -------- Utilities --------

//isSameDate utility function to check if two time.Time have the same date
func isSameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	if ay == by && am == bm && ad == bd {
		return true
	}
	return false
}

//dateFormatToLayout convert a date format %y,%m,%d to the funny golang layout 2006,01,02
func dateFormatToLayout(format string) string {
	format = strings.Replace(format, "%d", "02", 1)
	format = strings.Replace(format, "%m", "01", 1)
	format = strings.Replace(format, "%y", "2006", 1)
	return format
}
