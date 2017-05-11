package invoice

import (
	"errors"
	"fmt"
	"github.com/blevesearch/bleve"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const (
	FIELD_NUMBER   = "Number"
	FIELD_CUSTOMER = "Customer"
	FIELD_AMOUNT   = "Amount"
	FIELD_DATE     = "Date"

	QUERY_DATE_FORMAT       = "2006-01-02"
	QUERY_DEFAULT_DATE_FROM = "1970-01-01"
	QUERY_DEFAULT_AMOUNT_GE = float64(0)
	QUERY_DEFAULT_AMOUNT_LE = float64(1000000000000)
	QUERY_DEFAULT_CUSTOMER  = "none"
)

// -------- types --------

//InvoiceQuery is the query object to search for invoices
type InvoiceQuery struct {
	Customer string
	AmountGE float64
	AmountLE float64
	DateFrom time.Time
	DateTo   time.Time
}

func (q *InvoiceQuery) String() string {
	var f []string
	if q.Customer != QUERY_DEFAULT_CUSTOMER{
		f = append(f, fmt.Sprint("customer like ",q.Customer))
	}
	ddf, _ := time.Parse(QUERY_DATE_FORMAT, QUERY_DEFAULT_DATE_FROM)
	if !isSameDate(ddf, q.DateFrom) {
		f = append(f, fmt.Sprint("date > ",q.DateFrom.Format(QUERY_DATE_FORMAT)))

	}
	if !isSameDate(time.Now(), q.DateTo) {
		f = append(f, fmt.Sprint("date > ",q.DateFrom.Format(QUERY_DATE_FORMAT)))
	}
	// add range on amount if necessary
	if q.AmountGE != QUERY_DEFAULT_AMOUNT_GE {
		f = append(f, fmt.Sprint("amount > ",q.AmountGE))
	}
	// add range on amount if necessary
	if q.AmountLE != QUERY_DEFAULT_AMOUNT_LE {
		f = append(f, fmt.Sprint("amount < ",q.AmountLE))
	}
	return strings.Join(f, " and ")
}

//InvoiceQuery is the object indexed by bleve
type InvoiceEntry struct {
	Number   string
	Customer string
	Amount   float64
	Date     time.Time
}

// -------- exported functions --------

//DefaultInvoiceQuery return the default invoice query object
func DefaultInvoiceQuery() InvoiceQuery {
	df, _ := time.Parse(QUERY_DATE_FORMAT, QUERY_DEFAULT_DATE_FROM)
	return InvoiceQuery{
		Customer: QUERY_DEFAULT_CUSTOMER,
		AmountGE: float64(QUERY_DEFAULT_AMOUNT_GE),
		AmountLE: float64(QUERY_DEFAULT_AMOUNT_LE),
		DateFrom: df,
		DateTo:   time.Now(),
	}
}

//CreateSearchIndex create a new empty search index in the GetSearchIndexFilePath folder
func CreateSearchIndex() error {

	// index
	indexPath, exists := GetSearchIndexFilePath()
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
func RebuildSearchIndex(c *Config, password *string) (int, time.Duration, error) {
	var i Invoice
	var counter int = -1
	var elapsed time.Duration
	// delete old index if exists
	indexPath, exists := GetSearchIndexFilePath()
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
	files, _ := ioutil.ReadDir(c.Workspace)

	counter = 0
	for _, f := range files {
		if path.Ext(f.Name()) == EXT_CFB {
			descriptorPath := path.Join(c.Workspace, f.Name())
			if err := readInvoiceDescriptorEncrypted(&descriptorPath, &i, password); err == nil {
				// build the IndexEntry
				amount, _ := i.GetTotals()
				df := dateFormatToLayout(i.Settings.DateInputFormat)
				invd, _ := time.Parse(df, i.Invoice.Date)
				ie := InvoiceEntry{
					Number:   i.Invoice.Number,
					Customer: i.To.Name,
					Amount:   amount,
					Date:     invd,
				}
				// add the invoice to the index
				b.Index(i.Invoice.Number, ie)

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
func SearchInvoice(q InvoiceQuery) ([]InvoiceEntry, uint64, time.Duration, error) {
	var entries []InvoiceEntry
	var found uint64
	var elapsed time.Duration
	var err error
	// index
	indexPath, exists := GetSearchIndexFilePath()
	// if the index exists delete it
	if !exists {
		err = errors.New("search index does not exists, run govoice index to create the index")
		return entries, found, elapsed, err
	}
	// open  index
	index, err := bleve.Open(indexPath)
	if err != nil {
		err = errors.New("error opening the search inde")
		return entries, found, elapsed, err
	}
	defer index.Close()

	query := bleve.NewConjunctionQuery()
	// add phrase match for customer
	if q.Customer != QUERY_DEFAULT_CUSTOMER {
		subq := bleve.NewFuzzyQuery(q.Customer)
		subq.Fuzziness = 2
		subq.SetField(FIELD_CUSTOMER)
		query.AddQuery(subq)
	}
	// add range on amount if necessary
	if q.AmountLE != QUERY_DEFAULT_AMOUNT_LE || q.AmountGE != QUERY_DEFAULT_AMOUNT_GE {
		subq := bleve.NewNumericRangeQuery(&q.AmountGE, &q.AmountLE)
		subq.SetField(FIELD_AMOUNT)
		query.AddQuery(subq)
	}
	// add range query on date
	ddf, _ := time.Parse(QUERY_DATE_FORMAT, QUERY_DEFAULT_DATE_FROM)
	if !isSameDate(ddf, q.DateFrom) || !isSameDate(time.Now(), q.DateTo) {
		subq := bleve.NewDateRangeQuery(q.DateFrom, q.DateTo)
		subq.SetField(FIELD_DATE)
		query.AddQuery(subq)
	}

	if err := query.Validate(); err != nil {
		err = errors.New("invalid query")
		return entries, found, elapsed, err
	}

	search := bleve.NewSearchRequest(query)
	search.Fields = []string{FIELD_NUMBER, FIELD_CUSTOMER, FIELD_AMOUNT, FIELD_DATE}
	search.Size = 50
	results, err := index.Search(search)
	if err != nil {
		err = errors.New("error running the search")
		return entries, found, elapsed, err
	}

	for _, res := range results.Hits {

		d, _ := time.Parse(time.RFC3339, res.Fields[FIELD_DATE].(string))
		ie := InvoiceEntry{
			Number:   res.Fields[FIELD_NUMBER].(string),
			Customer: res.Fields[FIELD_CUSTOMER].(string),
			Amount:   res.Fields[FIELD_AMOUNT].(float64),
			Date:     d,
		}
		entries = append(entries, ie)
	}
	found = results.Total
	elapsed = results.Took
	return entries, found, elapsed, err
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
	dm.AddFieldMappingsAt(FIELD_NUMBER, inm)
	// text mapping for the customer name
	cfm := bleve.NewTextFieldMapping()
	dm.AddFieldMappingsAt(FIELD_CUSTOMER, cfm)
	// numeric mapping for the subtotal
	afm := bleve.NewNumericFieldMapping()
	dm.AddFieldMappingsAt(FIELD_AMOUNT, afm)
	// numeric mapping for the date
	dfm := bleve.NewDateTimeFieldMapping()
	dm.AddFieldMappingsAt(FIELD_DATE, dfm)
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
func addToSearchIndex(c *Config, i *Invoice) error {
	// index
	indexPath, exists := GetSearchIndexFilePath()
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
