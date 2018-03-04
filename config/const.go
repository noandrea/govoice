package config

// file extensions
const (
	ExtPdf           = "pdf"
	ExtJson          = "json"
	ExtToml          = "toml"
	ExtTemplate      = "tpl.toml"
	ExtJsonEncripted = "json.cfb"
	ExtCfb           = ".cfb"
)

// templates
const (
	DefaultTemplateName = "default"
)

// preview
const (
	PreviewFileName = "PREVIEW"
)

// searcing
const (
	FieldNumber   = "Number"
	FieldCustomer = "Customer"
	FieldAmount   = "Amount"
	FieldDate     = "Date"
	FieldText     = "Text"

	QueryDateFormat      = "2006-01-02"
	QueryDefaultDateFrom = "1970-01-01"
	QueryDefaultAmountGE = float64(0)
	QueryDefaultAmountLE = float64(1000000000000)
	QueryDefaultCustomer = "none"
	QueryDefaultText     = ""
)
