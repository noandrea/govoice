package invoice

import (
	"math/rand"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"gitlab.com/almost_cc/govoice/config"
)

//makeTmpHome create a random tmp home to execute tests to prevent race conditions.
func makeTmpHome() (string, string) {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	u1 := string(b)
	tmpHome := path.Join(os.TempDir(), "govoice", u1)
	tmpWorkspace := path.Join(tmpHome, "workspace")
	os.Setenv("HOME", tmpHome)
	return tmpHome, tmpWorkspace
}

func TestWorkspacePaths(t *testing.T) {

	tmpHome, tmpWorkspace := makeTmpHome()
	defer os.RemoveAll(tmpHome)
	t.Log("home is ", os.Getenv("HOME"))

	config.Main = config.MainConfig{
		Workspace:         tmpWorkspace,
		MasterDescriptor:  "_master",
		SearchResultLimit: 10,
	}
	// pdf
	filePath, exists := config.GetInvoicePdfPath("0000")

	if exists == true {
		t.Error("path ", filePath, "should not exists") // dumb test
	}

	if fp := path.Join(tmpWorkspace, "0000.pdf"); filePath != fp {
		t.Error("Expected\n", fp, "\nfound\n", filePath)
	}
	// json
	filePath, exists = config.GetInvoiceJsonPath("0000")
	if exists == true {
		t.Error("path ", filePath, "should not exists") // dumb test
	}

	if fp := path.Join(tmpWorkspace, "0000.json.cfb"); filePath != fp {
		t.Error("Expected\n", fp, "\nfound\n", filePath)
	}
}

func TestRenderPDF(t *testing.T) {
	tmpHome, tmpWorkspace := makeTmpHome()
	defer os.RemoveAll(tmpHome)
	t.Log("home is ", os.Getenv("HOME"))

	config.Main = config.MainConfig{
		Workspace:         tmpWorkspace,
		MasterDescriptor:  "_master",
		SearchResultLimit: 10,
	}
	// pdf
	filePath, exists := config.GetInvoicePdfPath("0000")
	if exists == true {
		t.Error("path ", filePath, "should not exists") // dumb test
	}

	if fp := path.Join(tmpWorkspace, "0000.pdf"); filePath != fp {
		t.Error("Expected\n", fp, "\nfound\n", filePath)
	}
	// json
	filePath, exists = config.GetInvoiceJsonPath("0000")
	if exists == true {
		t.Error("path ", filePath, "should not exists") // dumb test
	}

	if fp := path.Join(tmpWorkspace, "0000.json.cfb"); filePath != fp {
		t.Error("Expected\n", fp, "\nfound\n", filePath)
	}

	it := defaultTemplate()

	invoice := Invoice{}

	RenderPDF(&invoice, "", &it)
}

func TestSetup(t *testing.T) {
	tmpHome, tmpWorkspace := makeTmpHome()
	defer os.RemoveAll(tmpHome)

	t.Log("home is ", os.Getenv("HOME"))
	// check if config path and master path are set
	configPath, masterPath, err := Setup(tmpWorkspace)

	if err != nil {
		t.Error("error Setup", err)
	}

	if p := path.Join(tmpHome, ".govoice", "config.toml"); configPath != p {
		t.Error("Expected", p, "found", configPath)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Error("file", p, "does not exists")
		}
	}

	if p := path.Join(tmpWorkspace, "_master.json"); masterPath != p {
		t.Error("Expected", p, "found", masterPath)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Error("file", p, "does not exists")
		}
	}
}

func TestLoadTemplate(t *testing.T) {
	th, tmpWorkspace := makeTmpHome()
	defer os.RemoveAll(th)
	t.Log("home is ", os.Getenv("HOME"))
	config.Main = config.MainConfig{
		Workspace:         tmpWorkspace,
		MasterDescriptor:  "_master",
		SearchResultLimit: 10,
	}
	// load a mock template
	tpl := defaultTemplate()
	// get the path to the thing
	tp, te := config.GetTemplatePath(config.DefaultTemplateName)
	t.Log("template is ", tp)
	if te {
		t.Error(tp, "already exists and should not")
	}
	err := writeTomlToFile("/tmp/default.toml", tpl)
	err = writeTomlToFile(tp, tpl)
	if err != nil {
		t.Error(err)
	}

	tp, te = config.GetTemplatePath(config.DefaultTemplateName)
	if !te {
		t.Error(tp, "doesnt exists and should")
	}

	tpl2, err := readInvoiceTemplate(tp)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(tpl, tpl2) {
		t.Log("original", spew.Sprint(tpl))
		t.Log("cloned", spew.Sprint(tpl2))
		t.Error("templates are different")
	}

}
