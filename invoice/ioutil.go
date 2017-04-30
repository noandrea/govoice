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
	EXT_CFB = ".cfb"
)









// ReadInvoice parse the json file for an invoice
func readInvoiceDescriptor(path *string, i *Invoice) {
	rawJsonDescriptor, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(rawJsonDescriptor, &i)
}

func readInvoiceDescriptorEncrypted(path *string, i *Invoice, password *string)(error) {
	rawJsonDescriptor, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	rawJsonDescriptor = decryptCFB(*password, &rawJsonDescriptor)
	return json.Unmarshal(rawJsonDescriptor, &i)

}

// WriteInvoice write the invoice as descriptor
func writeInvoiceDescriptor(i *Invoice, workspace *string) {
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil {
		ioutil.WriteFile(fmt.Sprintf("%s/%s.json", *workspace, i.Invoice.Number), content, os.FileMode(0660))
	}
}

func writeInvoiceDescriptorEncrypted(i *Invoice, jsonPath, password *string) {
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil {
		encContent := encryptCFB(*password, &content)
		ioutil.WriteFile(*jsonPath, encContent, os.FileMode(0660))
	}
}

func writeTomlToFile(path string, v interface{}) {
	content, _ := toml.Marshal(v)
	if !strings.HasSuffix(path, ".toml"){
		path += ".toml"
	}
	ioutil.WriteFile(path, content, os.FileMode(0660))
}

func writeJsonToFile(path string, v interface{}) {
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
