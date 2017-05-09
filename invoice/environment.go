package invoice

import (
	"fmt"
	"os"
	"path"
)

// ============== CONFIGURATION FILE =================
type Config struct {
	Workspace         string `toml:"workspace"`
	SearchResultLimit int    `toml:"searchResultLimit"`
	MasterTemplate    string `toml:"masterTemplate"`
	DateInputFormat   string `toml:"dateInputFormat"`
	Layout            Layout `toml:"layout"`
}

type Layout struct {
	Style Style `toml:"style"`

	Items    Block `toml:"items"`
	From     Block `toml:"from"`
	To       Block `toml:"to"`
	Invoice  Block `toml:"invoice"`
	Payments Block `toml:"payments"`
	Notes    Block `toml:"notes"`
}

type Margins struct {
	Bottom float64 `toml:"bottom"`
	Left   float64 `toml:"left"`
	Right  float64 `toml:"right"`
	Top    float64 `toml:"top"`
}

type Style struct {
	Margins          Margins `toml:"margins"`
	FontFamily       string  `toml:"fontFamily"`
	FontSizeNormal   float64 `toml:"fontSizeNormal"`
	FontSizeH1       float64 `toml:"fontSizeH1"`
	FontSizeH2       float64 `toml:"fontSizeH2"`
	FontSizeSmall    float64 `toml:"fontsizeSmall"`
	LineHeightNormal float64 `toml:"lineHeightNormal"`
	LineHeightH1     float64 `toml:"lineHeightH1"`
	LineHeightH2     float64 `toml:"lineHeightH2"`
	LineHeightSmall  float64 `toml:"lineHeightSmall"`
	TableCol1W       float64 `toml:"tableCol1w"`
	TableCol2W       float64 `toml:"tableCol2w"`
	TableCol3W       float64 `toml:"tableCol3w"`
	TableCol4W       float64 `toml:"tableCol4w"`
	TableHeadHeight  float64 `toml:"tableHeadHeight"`
	TableRowHeight   float64 `toml:"tableRowHeight"`
}

func (c *Config) GetMasterPath() (string, bool) {

	dp := path.Join(c.Workspace, fmt.Sprintf("%s.json", c.MasterTemplate))
	if _, err := os.Stat(dp); os.IsNotExist(err) {
		return dp, false
	}
	return dp, true
}

//GetInvoiceJsonPath get the path of the encrypted version of an invoice in the workspace.
// returns the path of the invoice, and a boolean if the invoice already exists (true) or not (false)
func (c *Config) GetInvoiceJsonPath(name string) (string, bool) {
	return getPath(c.Workspace, name, EXT_JSONE)
}

func (c *Config) GetInvoicePdfPath(name string) (string, bool) {
	return getPath(c.Workspace, name, EXT_PDF)
}

// Block represents an pdf block
type Block struct {
	Position Coords
}

type Coords struct {
	X float64 `toml:"x"`
	Y float64 `toml:"y"`
}

//============== TRANSLATION ================

type Translation struct {
	From               I18NOther `toml:"from"`
	Sender             I18NOther `toml:"sender"`
	To                 I18NOther `toml:"to"`
	Recipient          I18NOther `toml:"recipient"`
	Invoice            I18NOther `toml:"invoice"`
	InvoiceData        I18NOther `toml:"invoice_data"`
	PaymentDetails     I18NOther `toml:"payment_details"`
	PaymentDetailsData I18NOther `toml:"payment_details_data"`
	Notes              I18NOther `toml:"notes"`
	Desc               I18NOther `toml:"desc"`
	Quantity           I18NOther `toml:"quantity"`
	Rate               I18NOther `toml:"rate"`
	Cost               I18NOther `toml:"cost"`
	Subtotal           I18NOther `toml:"subtotal"`
	Total              I18NOther `toml:"total"`
	Tax                I18NOther `toml:"tax"`
}

type I18NOther struct {
	Other string `toml:"other"`
}

// ======== functions =========

func GetConfigHome() string {
	return path.Join(os.Getenv("HOME"), ".govoice")
}

func GetConfigFilePath() string {
	return path.Join(GetConfigHome(), "config.toml")
}

func GetI18nHome() string {
	return path.Join(GetConfigHome(), "i18n")
}

func GetSearchIndexFilePath() (string, bool) {
	ifp := path.Join(GetConfigHome(), "index.bleve")
	if _, err := os.Stat(ifp); os.IsNotExist(err) {
		return ifp, false
	}
	return ifp, true
}

func GetI18nTranslationPath(lang string) string {
	if lang == "" {
		lang = "en"
	}

	l, _ := getPath(GetI18nHome(), lang, EXT_TOML)
	return l
}

// Setup setup the applications,
// create the workspace and the master invoice
// create the configuration with default values
// returns the configration file path and the master template path
func Setup(workspace string) (string, string, error) {
	var configPath, masterPath string
	// create configuration with defaults
	c := Config{
		Workspace:       workspace,
		MasterTemplate:  "_master",
		DateInputFormat: "%d/%m/%y",
		Layout: Layout{
			Style:    Style{Margins{0, 20, 20, 10}, "helvetica", 8, 14, 16, 6, 3.7, 6, 4, 3, 60, 13, 13, 13, 8, 6},
			Items:    Block{Coords{-1, 100}},
			From:     Block{Coords{-1, 28}},
			To:       Block{Coords{-1, 60}},
			Invoice:  Block{Coords{140, 28}},
			Payments: Block{Coords{-1, 210}},
			Notes:    Block{Coords{-1, 240}},
		},
	}
	// first create directories
	if err := os.MkdirAll(GetConfigHome(),0770); err != nil{
		return configPath, masterPath, err
	}

	if err := os.MkdirAll(GetI18nHome(),0770); err != nil{
		return configPath, masterPath, err
	}



	// write default configuration file
	configPath = GetConfigFilePath()
	err := writeTomlToFile(configPath, c)
	if err != nil {
		return configPath, masterPath, err
	}

	// write internationalization file
	en := Translation{
		From:               I18NOther{"FROM"},
		Sender:             I18NOther{"{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\nTax Number: {{.TaxId}}\nVAT: {{.VatNumber}}"},
		To:                 I18NOther{"TO"},
		Recipient:          I18NOther{"{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\nTax Number: {{.TaxId}}\nVAT: {{.VatNumber}}"},
		Invoice:            I18NOther{"INVOICE"},
		InvoiceData:        I18NOther{"N. {{.Number}}\nDate: {{.Date}}\nDue:{{.Due}}"},
		PaymentDetails:     I18NOther{"PAYMENTS DETAILS"},
		PaymentDetailsData: I18NOther{"{{.AccountHolder}}\n\nBank: {{.Bank}}\nIBAN: {{.Iban}}\nBIC: {{.Bic}}"},
		Notes:              I18NOther{"NOTES"},
		Desc:               I18NOther{"Description"},
		Quantity:           I18NOther{"Quantity"},
		Rate:               I18NOther{"Rate"},
		Cost:               I18NOther{"Cost"},
		Subtotal:           I18NOther{"Subtotal"},
		Total:              I18NOther{"Total"},
		Tax:                I18NOther{"VAT"},
	}
	enPath := GetI18nTranslationPath("en")
	err = writeTomlToFile(enPath, en)
	if err != nil {
		return configPath, masterPath, err
	}

	// write master.json
	// create the config directory if not exists
	_ = os.Mkdir(workspace, os.FileMode(0770))
	masterPath, exists := c.GetMasterPath()
	// don't overwrite the master if already exists
	if !exists {
		master := Invoice{
			From:           Recipient{"Mathis Hecht", "880 Whispering Half", "Hamburg", "67059", "Deutsheland", "9999999", "DE99999999999", "mh@ex.com"},
			To:             Recipient{"Encarnicion Tellez Nino", "Calle Burch No. 139", "Valencia", "19490", "España", "55555555", "ES55555555", "etn@bs.com"},
			PaymentDetails: BankCoordinates{"Mathis Hecht", "B Bank", "DE 1111 1111 1111 1111 11", "XXXXXXXX"},
			Invoice:        InvoiceData{"0000000", "01.01.2017", "01.02.2017"},
			Settings:       InvoiceSettings{45, 19, "€", "", "en", ""},
			Dailytime:      Daily{Enabled: false},
			Items:          &[]Item{Item{"web dev", 10, 0}, Item{"training", 5, 60}},
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
