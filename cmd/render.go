// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/spf13/cobra"
	gv "gitlab.com/almost_cc/govoice/invoice"
	gvext "gitlab.com/almost_cc/govoice/ext"
	"github.com/spf13/viper"
	"os"
	"log"
	"strings"
)

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "Render the master invoice to the target directory",
	Long: `
Render the invoice master in pdf in the workspace directory. It also create
a encrypted version of the invoice data`,
	Run: render,
}

func init() {
	RootCmd.AddCommand(renderCmd)

}


func render(cmd *cobra.Command, args []string) {

	// parse configuration
	var c gv.Config
	viper.Unmarshal(&c)

	// check if master exists
	descriptorPath, exists := c.GetMasterPath()
	if !exists {
		// file not exists, search for the encrypted version
		log.Fatal("master descriptor not found: ", descriptorPath)
		os.Exit(1)
	}

	// read the master descriptor
	var i gv.Invoice
	gv.ReadInvoiceDescriptor(&descriptorPath, &i)

	fmt.Println("rendering invoice number", i.Invoice.Number)

	// load translations
	i18n.MustLoadTranslationFile(gv.GetI18nTranslationPath(i.Settings.Language))
	T, _ := i18n.Tfunc(i.Settings.Language)

	// if Daylitime is enabled retrieve the content
	if i.Dailytime.Enabled {
		gvext.ScanItemsFromDaily(&i)
	}
	// render pdf
	pdfPath,pdfExists := c.GetInvoicePdfPath(i.Invoice.Number)
	jsonEncPath,jsonExists := c.GetInvoiceJsonPath(i.Invoice.Number)

	if pdfExists || jsonExists {
		reply := gv.ReadUserInput(fmt.Sprint("invoice ", i.Invoice.Number, " already exists, overwrite? [yes/no] yes"))
		if reply != "" && reply != "yes"{
			return
		}
	}

	if strings.TrimSpace(i.Invoice.Number)  == "" {
		fmt.Println("invoice number field must be filled")
		return
	}

	gv.RenderPDF(&i, &c.Layout, &pdfPath, T)
	// disable extensions in invoice
	i.DisableExtensions()
	// read the password
	password := gv.ReadUserPassword()
	gv.WriteInvoiceDescriptorEncrypted(&i, &jsonEncPath, &password)

	fmt.Println("encrypted descriptor created at", jsonEncPath)
	fmt.Println("pdf created at", pdfPath)
}
