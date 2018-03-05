package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ============== CONFIGURATION FILE =================

// GetConfigHome returns govoice configuration home
// that is USER_HOME/.govoice
func GetConfigHome() string {
	return path.Join(os.Getenv("HOME"), ".govoice")
}

// GetConfigFilePath returns the config file path
// default is CONFIG_HOME/config.toml
func GetConfigFilePath() string {
	return path.Join(GetConfigHome(), "config.toml")
}

// GetTemplatesHome returns the templates home
// default is CONFIG_HOME/templates/
func GetTemplatesHome() string {
	return path.Join(GetConfigHome(), "templates")
}

// GetSearchIndexFilePath returns the bleve index folder
// default is CONFIG_HOME/index.bleve/
func GetSearchIndexFilePath() (string, bool) {
	ifp := path.Join(GetConfigHome(), "index.bleve")
	if _, err := os.Stat(ifp); os.IsNotExist(err) {
		return ifp, false
	}
	return ifp, true
}

// GetTemplatePath returns the path of a template within the
// templates home folder (@see GetTemplatesHome)
func GetTemplatePath(name string) (templatePath string, templateExists bool) {
	// if the name is an absolute path just control if exists
	if filepath.IsAbs(name) {
		return name, FileExists(name)
	}

	// if name is empty use the default template
	if len(strings.TrimSpace(name)) == 0 {
		name = DefaultTemplateName
	}
	// and get the template from a folder
	return getPath(GetTemplatesHome(), name, ExtToml)
}

// MainConfig the main govoice configuration
type MainConfig struct {
	Workspace         string `toml:"workspace"`
	SearchResultLimit int    `toml:"searchResultLimit"`
	MasterDescriptor  string `toml:"masterDescriptor"`
	DateInputFormat   string `toml:"dateInputFormat"`
	DefaultInvoiceNet int    `toml:"defaultInvoiceNet"`
}

//GetMasterPath returns the path to the master invoice
func GetMasterPath() (masterFilePath string, masterFileExists bool) {
	return getPath(Main.Workspace, Main.MasterDescriptor, ExtJson)
}

// GetInvoiceJsonPath get the path of the encrypted version of an invoice in the workspace.
// returns the path of the invoice, and a boolean if the invoice already exists (true) or not (false)
func GetInvoiceJsonPath(name string) (string, bool) {
	return getPath(Main.Workspace, name, ExtJsonEncripted)
}

// GetInvoicePdfPath get the pdf path
func GetInvoicePdfPath(name string) (string, bool) {
	return getPath(Main.Workspace, name, ExtPdf)
}

//getPath build a path composed of baseFolder, fileName, extension
//
// returns the composed path and a bool to tell if the file exists (true) or not (false)
func getPath(basePath, fileName, ext string) (filePath string, fileExists bool) {
	filePath = path.Join(basePath, fmt.Sprintf("%s.%s", fileName, ext))
	fileExists = FileExists(filePath)
	return
}

// FileExists utility to check if a file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// Main main configuration loaded at startup
var Main MainConfig

// DebugEnabled if debugging is enbled or not
var DebugEnabled = false

// TemplateName loaded at startup via -t paramter
var TemplateName = DefaultTemplateName
