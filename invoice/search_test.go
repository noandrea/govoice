package invoice

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"gitlab.com/almost_cc/govoice/config"
)

const SAMPLE_SIZE = 1009

func remap(value, start1, stop1, start2, stop2 float64) (float64, error) {
	outgoing := start2 + (stop2-start2)*((value-start1)/(stop1-start1))
	if math.IsNaN(outgoing) {
		return 0, errors.New("NAN")

	} else if math.Abs(outgoing) == math.MaxFloat64 {
		return 0, errors.New("infinity")
	}
	return outgoing, nil
}

func invoices() []Invoice {
	cwd, _ := os.Getwd()
	p := path.Join(cwd, "_testresources", "breweries.csv")
	r, _ := os.Open(p)
	c := csv.NewReader(r)
	// skip header
	c.Read()

	var invoices []Invoice

	from := Recipient{"Mathis Hecht", "880 Whispering Half", "Hamburg", "67059", "Deutsheland", "9999999", "DE99999999999", "mh@ex.com"}

	countdown := SAMPLE_SIZE

	startDate, _ := time.Parse(config.QueryDateFormat, "2016-06-27")

	for {
		record, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Errorf("error: %s", err)
		}
		if countdown <= 0 {
			break
		}

		// 0: id,
		// 1: name
		// 2: address1
		// 3: address2
		// 4: city
		// 5: state
		// 6: code
		// 7: country
		// 8: phone
		// 9: website
		// 10: filepath,
		// 11: descript
		// 12: add_user
		// 13: last_mod

		to := Recipient{
			Name:      record[1],
			Address:   record[2],
			City:      record[4],
			Country:   record[7],
			AreaCode:  record[6],
			Email:     record[9],
			TaxId:     strings.Repeat(record[0], 8),
			VatNumber: fmt.Sprintf("%s %s", strings.ToUpper(record[7][0:2]), strings.Repeat(record[0], 8)),
		}

		numOfDays, _ := remap(float64(countdown), 0, float64(SAMPLE_SIZE), 1, 365)
		td := startDate.AddDate(0, 0, int(numOfDays))

		ins := strconv.Itoa(countdown)

		invd := InvoiceData{
			Number: fmt.Sprintf("%s%s", strings.Repeat("0", 7-len(ins)), ins),
			Date:   td.Format(config.QueryDateFormat),
			Due:    td.AddDate(0, 0, 30).Format(config.QueryDateFormat),
		}

		i := Invoice{
			From:           from,
			To:             to,
			PaymentDetails: BankCoordinates{"Mathis Hecht", "B Bank", "DE 1111 1111 1111 1111 11", "XXXXXXXX"},
			Invoice:        invd,
			Settings:       InvoiceSettings{45, "", 19, "€", "en", "%y-%m-%d"},
			Dailytime:      Daily{Enabled: false},
			Items:          &[]Item{Item{"web dev", float64(1 * countdown), 0, ""}, Item{"training", float64(2 * countdown), 5, ""}},
			Notes:          []string{"first note", "second note"},
		}
		invoices = append(invoices, i)

		//fmt.Println(i)

		countdown--
	}

	return invoices
}

func TestSearchInvoice(t *testing.T) {
	tmpHome, tmpWorkspace := makeTmpHome()
	defer os.RemoveAll(tmpHome)

	t.Log("home is ", os.Getenv("HOME"))

	// check if config path and master path are set
	cfp, mp, e := Setup(tmpWorkspace)
	if e != nil {
		t.Error(e)
	}

	t.Log(cfp, mp)

	config.Main = config.MainConfig{
		Workspace:        tmpWorkspace,
		MasterDescriptor: "_master",
	}

	pass := "abcd^^D[é123"
	pass = strings.Repeat(" ", 32-len(pass)) + pass
	// this creates the search index and add the invoices to the index
	for _, i := range invoices() {
		p, e := config.GetInvoiceJsonPath(i.Invoice.Number)
		if e {
			t.Error(p, "should not exists")
		}
		writeInvoiceDescriptorEncrypted(&i, p, pass)
		RestoreInvoice(i.Invoice.Number, pass)
		RenderInvoice(pass, "default")
	}

	runQueries(t)
}

func TestRebuildSearchIndex(t *testing.T) {
	tmpHome, tmpWorkspace := makeTmpHome()
	defer os.RemoveAll(tmpHome)

	t.Log("home is ", os.Getenv("HOME"))
	// check if config path and master path are set
	cfp, mp, e := Setup(tmpWorkspace)
	if e != nil {
		t.Error(e)
	}

	t.Log(cfp, mp)

	config.Main = config.MainConfig{
		Workspace:        tmpWorkspace,
		MasterDescriptor: "_master",
	}

	pass := "abcd^^D[é123"
	pass = strings.Repeat(" ", 32-len(pass)) + pass
	// this creates the search index and add the invoices to the index
	for _, i := range invoices() {
		p, _ := config.GetInvoiceJsonPath(i.Invoice.Number)
		writeInvoiceDescriptorEncrypted(&i, p, pass)
		RestoreInvoice(i.Invoice.Number, pass)
		RenderInvoice(pass, "default")
	}

	// this recreates the seaarch index should have the same results as above
	counter, elapsed, _ := RebuildSearchIndex(pass)
	t.Log("indexed", counter, "invoices in", elapsed)

	if counter != SAMPLE_SIZE {
		t.Error("expected", SAMPLE_SIZE, "got", counter)
	}

	runQueries(t)

}

func runQueries(t *testing.T) {
	var expected, found uint64
	var entries []InvoiceEntry
	var elapsed time.Duration
	var q InvoiceQuery
	var err error

	// #####################  first search
	q = DefaultInvoiceQuery()

	q.DateFrom, _ = time.Parse(config.QueryDateFormat, "2017-01-01")
	q.DateTo, _ = time.Parse(config.QueryDateFormat, "2017-01-10")

	// 25, 25, xx, nil
	entries, found, elapsed, err = SearchInvoice(q)
	expected = 25

	check(t, q, expected, entries, found, elapsed, err)

	// #####################  second search
	q = DefaultInvoiceQuery()

	q.DateFrom, _ = time.Parse(config.QueryDateFormat, "2017-01-01")
	q.DateTo, _ = time.Parse(config.QueryDateFormat, "2017-02-10")
	// 110, 110, xx, nil
	expected = 111

	entries, found, elapsed, err = SearchInvoice(q)
	check(t, q, expected, entries, found, elapsed, err)

	// #####################  filter by amount
	q = DefaultInvoiceQuery()

	q.AmountGE = 10
	q.AmountLE = 1000
	// 18, 18, xx, nil
	expected = 18

	entries, found, elapsed, err = SearchInvoice(q)
	check(t, q, expected, entries, found, elapsed, err)

	// #####################  filter by date and amount
	q = DefaultInvoiceQuery()

	q.AmountGE = 10
	q.AmountLE = 1000
	q.DateFrom, _ = time.Parse(config.QueryDateFormat, "2016-06-29")
	q.DateTo, _ = time.Parse(config.QueryDateFormat, "2016-07-01")
	// 6, 6, xx, nil
	expected = 6

	entries, found, elapsed, err = SearchInvoice(q)
	check(t, q, expected, entries, found, elapsed, err)

	// #####################  filter by date, amount, customer
	q = DefaultInvoiceQuery()

	q.AmountGE = 1000
	q.AmountLE = 20000
	q.DateFrom, _ = time.Parse(config.QueryDateFormat, "2016-06-27")
	q.DateTo, _ = time.Parse(config.QueryDateFormat, "2017-07-10")
	q.Customer = "pizz"
	expected = 4

	// 5, 5, xx, nil
	entries, found, elapsed, err = SearchInvoice(q)
	check(t, q, expected, entries, found, elapsed, err)
}

func check(t *testing.T, q InvoiceQuery, expected uint64, entries []InvoiceEntry, found uint64, elapsed time.Duration, err error) {
	t.Log("query", q)
	t.Log("found", found, "entries in ", elapsed)
	if err != nil {
		t.Error("error", err)
	}
	if found != expected {
		t.Error("expected", expected, "got", found)
	}
	for _, r := range entries {
		if r.Date.After(q.DateTo) || r.Date.Before(q.DateFrom) {
			t.Error("expected", r.Date, "between", q.DateFrom, "and", q.DateTo)
		}
		if r.Amount > q.AmountLE || r.Amount < q.AmountGE {
			t.Error("expected", r.Amount, "between", q.AmountLE, "and", q.AmountGE)
		}
		t.Log("number:", r.Number, "date:", r.Date, "amount:", r.Amount, "customer:", r.Customer)
	}
}
