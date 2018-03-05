package invoice

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pelletier/go-toml"
	"gitlab.com/almost_cc/govoice/config"
	"golang.org/x/crypto/ssh/terminal"
)

// write a file and creates intermediate directories if
// they not exists
func writeFile(path string, content []byte) (err error) {
	pp := filepath.Dir(path)
	if !config.FileExists(pp) {
		if err = os.MkdirAll(pp, os.FileMode(0770)); err != nil {
			return
		}
	}
	return ioutil.WriteFile(path, content, os.FileMode(0660))
}

// ReadInvoice parse the json file for an invoice
func readInvoiceDescriptor(path string) (i Invoice, err error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = json.Unmarshal(rawData, &i)
	return
}

func readInvoiceTemplate(path string) (tpl InvoiceTemplate, err error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = toml.Unmarshal(rawData, &tpl)
	return
}

func readInvoiceDescriptorEncrypted(path, password string) (i Invoice, err error) {
	rawData, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	rawData = decryptCFB(password, &rawData)
	err = json.Unmarshal(rawData, &i)
	return
}

// WriteInvoice write the invoice as descriptor
func writeInvoiceDescriptor(i *Invoice) {
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil {
		writeFile(fmt.Sprintf("%s/%s.json", config.Govoice.Workspace, i.Invoice.Number), content)
	}
}

func writeInvoiceDescriptorEncrypted(i *Invoice, jsonPath, password string) {
	content, err := json.MarshalIndent(*i, "", "  ")
	if err == nil {
		encContent := encryptCFB(password, &content)
		writeFile(jsonPath, encContent)
	}
}

func writeTomlToFile(path string, v interface{}) error {
	content, _ := toml.Marshal(v)
	fileExt := fmt.Sprintf(".%s", config.ExtToml)
	if !strings.HasSuffix(path, fileExt) {
		path += fileExt
	}
	return writeFile(path, content)
}

func writeJsonToFile(path string, v interface{}) error {
	content, _ := json.MarshalIndent(v, "", "  ")
	fileExt := fmt.Sprintf(".%s", config.ExtJson)
	if !strings.HasSuffix(path, fileExt) {
		path += fileExt
	}
	return writeFile(path, content)
}

func ReadMasterDescriptor() (invoice Invoice, err error) {
	// check if master exists
	descriptorPath, exists := config.GetMasterPath()
	if !exists {
		// file not exists, search for the encrypted version
		err = errors.New("master descriptor not found")
		return
	}
	invoice, err = readInvoiceDescriptor(descriptorPath)
	return
}

func ReadUserPassword(message string) (string, error) {
	// password
	fmt.Print(message)
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	// pad the key for aes encryption
	password := fmt.Sprintf("%32s", strings.TrimSpace(string(bytePassword)))
	if len(password) > 32 {
		return "", errors.New("password is too long (max 32 characters)")
	}
	fmt.Println()
	return password, nil
}

func ReadUserInput(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message + ": ")
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
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
