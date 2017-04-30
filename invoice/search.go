package invoice

import (
	"os"
	"github.com/blevesearch/bleve"
	"fmt"
	"io/ioutil"
	"path"
	"errors"
	"time"
)

type InvoiceEntry struct {
	Number   string
	Customer string
	Amount   float64
	Date     string
}

const(
	FIELD_NUMBER   ="Number"
	FIELD_CUSTOMER ="Customer"
	FIELD_AMOUNT   ="Amount"
	FIELD_DATE     ="Date"
)

//CreateIndex creates the bleve index
func CreateSearchIndex(c *Config, password *string)(int,error){
	var i Invoice
	// open a new index
	mapping := bleve.NewIndexMapping()

	dm := bleve.NewDocumentMapping()
	// text mapping for the invoce number
	inm := bleve.NewTextFieldMapping()
	dm.AddFieldMappingsAt(FIELD_NUMBER,inm)
	// text mapping for the customer name
	cfm := bleve.NewTextFieldMapping()
	dm.AddFieldMappingsAt(FIELD_CUSTOMER,cfm)
	// numeric mapping for the subtotal
	afm := bleve.NewNumericFieldMapping()
	dm.AddFieldMappingsAt(FIELD_AMOUNT,afm)
	// numeric mapping for the date
	dfm := bleve.NewDateTimeFieldMapping()
	dm.AddFieldMappingsAt(FIELD_DATE,dfm)
	// add document mapping
	mapping.AddDocumentMapping("invoice", dm)

	// index
	indexPath, exists := GetSearchIndexFilePath()
	// if the index exists delete it
	if exists {
		os.RemoveAll(indexPath)
	}
	// create a new index
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		return -1, errors.New("cannot open or create the search index")

	}
	defer index.Close()
	// create a new batch
	b := index.NewBatch()

	// scan the descriptor files
	files, _ := ioutil.ReadDir(c.Workspace)

	counter := 0
	for _, f := range files {
		if path.Ext(f.Name()) == EXT_CFB {
			descriptorPath := path.Join(c.Workspace, f.Name())
			if err := readInvoiceDescriptorEncrypted(&descriptorPath, &i, password); err == nil {

				amount, _ := i.GetTotals()
				ie := InvoiceEntry{
					i.Invoice.Number,
					i.To.Name,
					amount,
					i.Invoice.Date,
				}
				// add the invoice to the index
				b.Index(i.Invoice.Number, ie)

				counter++
				if counter % 100 == 0 {
					index.Batch(b)
				}
			} else {
				fmt.Println("error decrypting ", f.Name(), ", the invoice will not be searchable")
			}
		}
	}
	index.Batch(b)
	// return the numer of entries indexed
	return counter,nil
}


func SearchInvoice(queryString string)([]InvoiceEntry,uint64,time.Duration,error){
	var entries []InvoiceEntry
	// index
	indexPath, exists := GetSearchIndexFilePath()
	// if the index exists delete it
	if !exists {
		return entries,0,time.Duration(0), errors.New("search index does not exists, run govoice index to create the index")
	}
	// open  index
	index, err := bleve.Open(indexPath)
	if err != nil {
		return entries,0,time.Duration(0), errors.New("error opening the search inde")
	}
	defer index.Close()

	query := bleve.NewQueryStringQuery(queryString)
	if err := query.Validate(); err != nil{
		return entries,0,time.Duration(0), errors.New("invalid query")
	}
	search := bleve.NewSearchRequest(query)
	search.Fields = []string{FIELD_NUMBER,FIELD_CUSTOMER,FIELD_AMOUNT,FIELD_DATE}
	search.Size = 10
	results, err := index.Search(search)
	if err != nil{
		return entries,0,time.Duration(0), errors.New("error running the search")
	}

	for _, res := range results.Hits {

		ie := InvoiceEntry{
			res.Fields[FIELD_NUMBER].(string),
			res.Fields[FIELD_CUSTOMER].(string),
			res.Fields[FIELD_AMOUNT].(float64),
			res.Fields[FIELD_DATE].(string),

		}
		entries = append(entries, ie)
	}
	return entries,results.Total,results.Took,nil
}
