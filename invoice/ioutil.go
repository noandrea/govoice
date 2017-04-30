package invoice

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"path"
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
	"bufio"
)

const (
	EXT_PDF = "pdf"
	EXT_JSON = "json"
	EXT_TOML = "toml"
	EXT_JSONE = "json.cfb"
)


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

type Daily struct {
	Enabled  bool           `json:"enabled"`
	DateFrom string         `json:"date_from,omitempty"`
	DateTo   string         `json:"date_to",omitempty`
	Projects []DailyProject `json:"projects",omitempty`
}

type DailyProject struct {
	Name            string `json:"name"`
	ItemDescription string `json:"item_description"`
}

type InvoiceSettings struct {
	BaseItemPrice  float64 `json:"base_item_price"`
	VatRate        float64  `json:"vat_rate"`
	CurrencySymbol string  `json:"currency_symbol"`
	QuantitySymbol string  `json:"quantity_symbol"`
	Language       string  `json:"lang"`
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
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	ItemPrice   float64 `json:"item_price,omitempty"`
}

func (i *Invoice) PushItem(description string, quantity float64) {
	*i.Items = append(*i.Items, Item{description, quantity, -1})
}

func (i *Invoice) DisableExtensions() {
	i.Dailytime.Enabled = false;
}

// ============== CONFIGURATION FILE =================
type Config struct {
	Workspace      string `toml:"workspace"`
	Encrypt        bool   `toml:"encrypt"`
	MasterTemplate string `toml:"masterTemplate"`
	DefaultLang    string `toml:"defaultLang"`
	Layout         Layout `toml:"layout"`
}

type Layout struct {
	Style    Style  `toml:"style"`

	Items    Block  `toml:"items"`
	From     Block  `toml:"from"`
	To       Block  `toml:"to"`
	Invoice  Block  `toml:"invoice"`
	Payments Block  `toml:"payments"`
	Notes    Block  `toml:"notes"`
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
	TableCol1W      float64  `toml:"tableCol1w"`
	TableCol2W      float64  `toml:"tableCol2w"`
	TableCol3W      float64  `toml:"tableCol3w"`
	TableCol4W      float64  `toml:"tableCol4w"`
	TableHeadHeight float64  `toml:"tableHeadHeight"`
	TableRowHeight  float64  `toml:"tableRowHeight"`
}

func (c *Config) GetMasterPath()(string, bool){

	dp := path.Join(c.Workspace, fmt.Sprintf("%s.json",c.MasterTemplate))
	if _, err := os.Stat(dp); os.IsNotExist(err) {
		return dp, false
	}
	return dp, true
}

func (c *Config) GetInvoiceJsonPath(name string)(string, bool){
	return getPath(c.Workspace, name, EXT_JSONE)
}

func (c *Config) GetInvoicePdfPath(name string)(string, bool){
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



func GetConfigHome()(string){
	return path.Join(os.Getenv("HOME"), ".govoice")
}

func GetConfigFilePath()(string){
	return path.Join(GetConfigHome(), "config.toml")
}

func GetI18nHome()(string){
	return path.Join(GetConfigHome(), "i18n")
}

func GetI18nTranslationPath(lang string)(string){
	if lang == ""{
		lang = "en"
	}

	l,_ := getPath(GetI18nHome(), lang, EXT_TOML )
	return l
}

// ReadInvoice parse the json file for an invoice
func ReadInvoiceDescriptor(path *string, i *Invoice) {
	rawJsonDescriptor, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(rawJsonDescriptor, &i)
}

func ReadInvoiceDescriptorEncrypted(path *string, i *Invoice, password *string)(error) {
	rawJsonDescriptor, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	rawJsonDescriptor = decryptCFB(*password, &rawJsonDescriptor)
	return json.Unmarshal(rawJsonDescriptor, &i)

}

// WriteInvoice write the invoice as descriptor
func WriteInvoiceDescriptor(i *Invoice, workspace *string) {
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil {
		ioutil.WriteFile(fmt.Sprintf("%s/%s.json", *workspace, i.Invoice.Number), content, os.FileMode(0660))
	}
}

func WriteInvoiceDescriptorEncrypted(i *Invoice, jsonPath, password *string) {
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil {
		encContent := encryptCFB(*password, &content)
		ioutil.WriteFile(*jsonPath, encContent, os.FileMode(0660))
	}
}

func WriteTomlToFile(path string, v interface{}) {
	content, _ := toml.Marshal(v)
	if !strings.HasSuffix(path, ".toml"){
		path += ".toml"
	}
	ioutil.WriteFile(path, content, os.FileMode(0660))
}

func WriteJsonToFile(path string, v interface{}) {
	content, _ := json.MarshalIndent(v, "", "  ")
	if !strings.HasSuffix(path, ".json"){
		path += ".json"
	}
	ioutil.WriteFile(path, content, os.FileMode(0660))
}

func ReadUserPassword()(string){
	// password
	fmt.Print("Enter Password: ")
	bytePassword,_ := terminal.ReadPassword(int(syscall.Stdin))
	// pad the key for aes encryption
	password := fmt.Sprintf("%32s",strings.TrimSpace(string(bytePassword)))
	if len(password) > 32 {
		println("password is too long (max 32 characters)")
		os.Exit(1)
	}
	fmt.Println()
	return password
}

func ReadUserInput(message string)(string){
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(message)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func getPath(basePath, fileName, ext string )(string, bool) {
	dp := path.Join(basePath, fmt.Sprintf("%s.%s",fileName, ext))
	if _, err := os.Stat(dp); os.IsNotExist(err) {
		return dp, false
	}
	return dp, true
}

func decryptCFB(k string, data *[]byte) []byte {
	key := []byte(k)
	ciphertext, _ := hex.DecodeString(string(*data))

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)
	// Output: some plaintext
	return ciphertext
}

func encryptCFB(k string, plaintext *[]byte) []byte {
	key := []byte(k)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(*plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], *plaintext)

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	cypertexthex := make([]byte, hex.EncodedLen(len(ciphertext)))
	hex.Encode(cypertexthex, ciphertext)
	return cypertexthex
}
