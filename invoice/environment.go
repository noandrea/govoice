package invoice

import (
	"os"

	"gitlab.com/almost_cc/govoice/config"
)

// InvoiceTemplate the template to genreate the pdf
type InvoiceTemplate struct {
	Page     Page               `toml:"page"`
	Sections map[string]Section `toml:"sections"`
}

type Page struct {
	Orientation     string  `toml:"orientation"`
	Size            string  `toml:"size"`
	BackgroundColor []int   `toml:"background_color"`
	FontColor       []int   `toml:"font_color"`
	Margins         Margins `toml:"margins"`
	Font            Font    `toml:"font"`
	Table           Table   `toml:"table"`
}

type Margins struct {
	Bottom float64 `toml:"bottom"`
	Left   float64 `toml:"left"`
	Right  float64 `toml:"right"`
	Top    float64 `toml:"top"`
}

type Font struct {
	Family           string  `toml:"family"`
	SizeNormal       float64 `toml:"sizeNormal"`
	SizeH1           float64 `toml:"size_h1"`
	SizeH2           float64 `toml:"size_h2"`
	SizeSmall        float64 `toml:"size_small"`
	LineHeightNormal float64 `toml:"line_height_normal"`
	LineHeightH1     float64 `toml:"line_height_h1"`
	LineHeightH2     float64 `toml:"line_height_h2"`
	LineHeightSmall  float64 `toml:"line_height_small"`
}

type Table struct {
	Col1W                 float64  `toml:"col1w"`
	Col2W                 float64  `toml:"col2w"`
	Col3W                 float64  `toml:"col3w"`
	Col4W                 float64  `toml:"col4w"`
	HeadHeight            float64  `toml:"head_height"`
	RowHeight             float64  `toml:"row_height"`
	Header                []string `toml:"header"`
	LabelTotal            string   `toml:"label_total"`
	LabelSubtotal         string   `toml:"label_subtotal"`
	LabelTax              string   `toml:"label_tax"`
	HeaderFontColor       []int    `toml:"header_font_color"`
	HeaderBackgroundColor []int    `toml:"header_background_color"`
}

// Section represents an pdf block
type Section struct {
	X        float64 `toml:"x"`
	Y        float64 `toml:"y"`
	Title    string  `toml:"title"`
	Template string  `toml:"tpl"`
	Content  string  `toml:"-"`
}

// Setup setup the applications,
// create the workspace and the master invoice
// create the configuration with default values
// returns the configration file path and the master template path
func Setup(workspace string) (configPath string, masterPath string, err error) {
	// create configuration with defaults
	config.Govoice = config.MainConfig{
		Workspace:         workspace,
		MasterDescriptor:  "_master",
		DateInputFormat:   "%d.%m.%y",
		SearchResultLimit: 50,
		DefaultInvoiceNet: 30,
	}
	// first create directories
	if err = os.MkdirAll(config.GetConfigHome(), 0770); err != nil {
		return
	}

	if err = os.MkdirAll(config.GetTemplatesHome(), 0770); err != nil {
		return
	}

	// write default configuration file
	configPath = config.GetConfigFilePath()
	err = writeTomlToFile(configPath, config.Govoice)
	if err != nil {
		return
	}

	var exists bool
	// write default template
	if tplPath, exists := config.GetTemplatePath(config.DefaultTemplateName); !exists {
		err = writeTomlToFile(tplPath, defaultTemplate())
		if err != nil {
			return
		}
	}
	// write master.json
	// create the config directory if not exists
	_ = os.Mkdir(workspace, os.FileMode(0770))
	if masterPath, exists = config.GetMasterPath(); !exists {
		err = writeJsonToFile(masterPath, masterInvoice())
		if err != nil {
			return
		}
	}

	// create bleve index
	err = CreateSearchIndex()

	return
}

func masterInvoice() (master Invoice) {
	master = Invoice{
		From:           Recipient{"My Name", "My Address", "My City", "My Post Code", "My Country", "My Tax ID", "My VAT Number", "My Email"},
		To:             Recipient{"Customer Name", "Customer Address", "Customer City", "Customer Post Code", "Customer Country", "Customre Tax ID", "Customer VAT number", "Customer Email"},
		PaymentDetails: BankCoordinates{"My Name", "My Bank Name", "My IBAN", "My BIC/SWIFT"},
		Invoice:        InvoiceData{"0000000", "23.01.2017", "23.02.2017"},
		Settings:       InvoiceSettings{45, "", 19, "€", "en", "", false},
		Dailytime:      Daily{Enabled: false},
		Items:          &[]Item{Item{"item 1 description", 10, 0, ""}, Item{"item 2 description", 5, 60, ""}},
		Notes:          []string{"first note", "second note"},
	}
	return
}

func defaultTemplate() (tpl InvoiceTemplate) {
	tpl = InvoiceTemplate{
		Page: Page{
			BackgroundColor: []int{255, 255, 255},
			FontColor:       []int{0, 0, 0},
			Orientation:     "P",
			Size:            "A4",
			Font: Font{
				Family:           "helvetica",
				LineHeightH1:     8.0,
				LineHeightH2:     7.0,
				LineHeightNormal: 3.7,
				LineHeightSmall:  3.0,
				SizeNormal:       8.0,
				SizeH1:           14.0,
				SizeH2:           16.0,
				SizeSmall:        6.0,
			},
			Margins: Margins{
				Bottom: 0,
				Left:   20,
				Right:  20,
				Top:    10,
			},

			Table: Table{
				Col1W:                 60.0,
				Col2W:                 13.0,
				Col3W:                 13.0,
				Col4W:                 13.0,
				HeadHeight:            8.0,
				RowHeight:             6.0,
				HeaderBackgroundColor: []int{0, 0, 0},
				HeaderFontColor:       []int{255, 255, 255},
				Header:                []string{"Description", "Quantity", "Rate", "Cost"},
				LabelTotal:            "total",
				LabelSubtotal:         "sbutotal",
				LabelTax:              "tax",
			},
		},
	}

	tpl.Sections = make(map[string]Section)
	tpl.Sections["from"] = Section{
		Title:    "FROM",
		Template: `{{.Name}}\n{{.Address}}\n\t\t{{.AreaCode}}, {{.City}}\n\t\t{{.Country}}\n\t\t{{if .TaxId }}Tax Number: {{.TaxId}} {{end}}\n\t\t{{if .VatNumber }}VAT: {{.VatNumber}} {{end}}\n\t\t`,
		X:        -1.0,
		Y:        28.0,
	}

	tpl.Sections["invoice"] = Section{
		Title:    "INVOICE",
		Template: "N.      {{.Number}}\nDate: {{.Date}}\nDue:  {{.Due}}\n\t\t",
		X:        140.0,
		Y:        28.0,
	}

	tpl.Sections["notes"] = Section{
		Title:    "NOTES",
		Template: "{{ range . }}{{ . }}\n{{ end }}",
		X:        -1.0,
		Y:        240.0,
	}

	tpl.Sections["payments"] = Section{
		Title:    "PAYMENTS DETAILS",
		Template: "{{.AccountHolder}}\n\nBank: {{.Bank}}\nIBAN: {{.Iban}}\nBIC:  {{.Bic}}",
		X:        -1.0,
		Y:        210.0,
	}

	tpl.Sections["to"] = Section{
		Title:    "TO",
		Template: "{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\n{{if .TaxId }}Tax Number: {{.TaxId}} {{end}}\n{{if .VatNumber }}VAT: {{.VatNumber}} {{end}}\n\t\t",
		X:        -1.0,
		Y:        65.0,
	}

	return
}
