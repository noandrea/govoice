package invoice

import (
	"os"
	"path"
	"testing"
)

// ReadInvoice parse the json file for an invoice
func TestReadInvoiceDescriptor(t *testing.T) {
	cwd, _ := os.Getwd()
	path := path.Join(cwd, "_testresources", "0001.json")
	i, err := readInvoiceDescriptor(path)

	if err != nil {
		t.Error("expected nil, found", err)
	}

	ivoiceNumber := "0001"
	if i.Invoice.Number != ivoiceNumber {
		t.Error("expected", ivoiceNumber, "found", i.Invoice.Number)
	}
}

func TestReadInvoiceDescriptorEncrypted(t *testing.T) {
	cwd, _ := os.Getwd()
	path := path.Join(cwd, "_testresources", "0001.json.cfb")
	var i Invoice
	wrongPass := "                     xxxxxxxxxxx"
	i, err := readInvoiceDescriptorEncrypted(path, wrongPass)
	if err == nil {
		t.Error("unexpected", nil, "as error")
	}

	rightPass := "                        12345678"
	i, err = readInvoiceDescriptorEncrypted(path, rightPass)
	if err != nil {
		t.Error("unexpected", err, "as error")
	}

	ivoiceNumber := "0001"
	if i.Invoice.Number != ivoiceNumber {
		t.Error("expected", ivoiceNumber, "found", i.Invoice.Number)
	}
}
