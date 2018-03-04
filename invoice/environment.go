package invoice

import (
	"bytes"
	"os"
	"text/template"

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
	Col1W         float64  `toml:"col1w"`
	Col2W         float64  `toml:"col2w"`
	Col3W         float64  `toml:"col3w"`
	Col4W         float64  `toml:"col4w"`
	HeadHeight    float64  `toml:"head_height"`
	RowHeight     float64  `toml:"row_height"`
	Header        []string `toml:"header"`
	LabelTotal    string   `toml:"label_total"`
	LabelSubtotal string   `toml:"label_subtotal"`
	LabelTax      string   `toml:"label_tax"`
}

// Section represents an pdf block
type Section struct {
	X        float64 `toml:"x"`
	Y        float64 `toml:"y"`
	Title    string  `toml:"title"`
	Template string  `toml:"tpl"`
	Content  string  `toml:"-"`
}

// render a section template with the data
func (s Section) applyTemplate(data interface{}) (content string) {
	tmpl, err := template.New("__").Parse(s.Template)
	if err != nil {
		panic(err)
	}
	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		panic(err)
	}
	content = rendered.String()
	return
}

// Setup setup the applications,
// create the workspace and the master invoice
// create the configuration with default values
// returns the configration file path and the master template path
func Setup(workspace string) (string, string, error) {
	var configPath, masterPath string
	// create configuration with defaults
	c := config.MainConfig{
		Workspace:         workspace,
		MasterDescriptor:  "_master",
		DateInputFormat:   "%d.%m.%y",
		SearchResultLimit: 50,
		// Layout: Layout{
		// 	Style:    Style{Margins{0, 20, 20, 10}, "helvetica", 8, 14, 16, 6, 3.7, 6, 4, 3, 60, 13, 13, 13, 8, 6},
		// 	Items:    Block{Coords{-1, 100}},
		// 	From:     Block{Coords{-1, 28}},
		// 	To:       Block{Coords{-1, 60}},
		// 	Invoice:  Block{Coords{140, 28}},
		// 	Payments: Block{Coords{-1, 210}},
		// 	Notes:    Block{Coords{-1, 240}},
		// },
	}
	// first create directories
	if err := os.MkdirAll(config.GetConfigHome(), 0770); err != nil {
		return configPath, masterPath, err
	}

	if err := os.MkdirAll(config.GetTemplatesHome(), 0770); err != nil {
		return configPath, masterPath, err
	}

	// write default configuration file
	configPath = config.GetConfigFilePath()
	err := writeTomlToFile(configPath, c)
	if err != nil {
		return configPath, masterPath, err
	}

	// write internationalization file

	// en := Translation{
	// 	From:               I18NOther{"FROM"},
	// 	Sender:             I18NOther{"{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\nTax Number: {{.TaxId}}\nVAT: {{.VatNumber}}"},
	// 	To:                 I18NOther{"TO"},
	// 	Recipient:          I18NOther{"{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\nTax Number: {{.TaxId}}\nVAT: {{.VatNumber}}"},
	// 	Invoice:            I18NOther{"INVOICE"},
	// 	InvoiceData:        I18NOther{"N. {{.Number}}\nDate: {{.Date}}\nDue:{{.Due}}"},
	// 	PaymentDetails:     I18NOther{"PAYMENTS DETAILS"},
	// 	PaymentDetailsData: I18NOther{"{{.AccountHolder}}\n\nBank: {{.Bank}}\nIBAN: {{.Iban}}\nBIC: {{.Bic}}"},
	// 	Notes:              I18NOther{"NOTES"},
	// 	Desc:               I18NOther{"Description"},
	// 	Quantity:           I18NOther{"Quantity"},
	// 	Rate:               I18NOther{"Rate"},
	// 	Cost:               I18NOther{"Cost"},
	// 	Subtotal:           I18NOther{"Subtotal"},
	// 	Total:              I18NOther{"Total"},
	// 	Tax:                I18NOther{"VAT"},
	// }
	enPath, exists := config.GetTemplatePath(config.DefaultTemplateName)
	//TODO not good
	err = writeTomlToFile(enPath, "")
	if err != nil {
		return configPath, masterPath, err
	}

	// write master.json
	// create the config directory if not exists
	_ = os.Mkdir(workspace, os.FileMode(0770))
	masterPath, exists = config.GetMasterPath()
	// don't overwrite the master if already exists
	if !exists {
		master := Invoice{
			From:           Recipient{"My Name", "My Address", "My City", "My Post Code", "My Country", "My Tax ID", "My VAT Number", "My Email"},
			To:             Recipient{"Customer Name", "Customer Address", "Customer City", "Customer Post Code", "Customer Country", "Customre Tax ID", "Customer VAT number", "Customer Email"},
			PaymentDetails: BankCoordinates{"My Name", "My Bank Name", "My IBAN", "My BIC/SWIFT"},
			Invoice:        InvoiceData{"0000000", "23.01.2017", "23.02.2017"},
			Settings:       InvoiceSettings{45, "", 19, "€", "en", "", false},
			Dailytime:      Daily{Enabled: false},
			Items:          &[]Item{Item{"item 1 description", 10, 0, ""}, Item{"item 2 description", 5, 60, ""}},
			Notes:          []string{"first note", "second note"},
		}
		err = writeJsonToFile(masterPath, master)
		if err != nil {
			return configPath, masterPath, err
		}

	}

	// create bleve index
	err = CreateSearchIndex()

	return configPath, masterPath, err
}
