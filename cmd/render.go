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

	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"gitlab.com/almost_cc/govoice/config"
	govoice "gitlab.com/almost_cc/govoice/invoice"
)

// renderCmd represents the render command
var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "render the master invoice in the workspace",
	Long: `
Render the invoice master in pdf in the workspace directory. It also create
a encrypted version of the invoice data`,
	Run: render,
}

func init() {
	RootCmd.AddCommand(renderCmd)
	tp, _ := config.GetTemplatePath(config.DefaultTemplateName)
	help := fmt.Sprintln("template name or path, defaults to:", tp)
	renderCmd.Flags().StringVarP(&templatePath, "template", "t", config.DefaultTemplateName, help)
}

var templatePath string

func render(cmd *cobra.Command, args []string) {
	// read the password
	password, err := govoice.ReadUserPassword("Enter password:")
	if err != nil {
		fmt.Println(err)
		return
	}

	templatePath, te := config.GetTemplatePath(templatePath)
	// if template path is a string, load the template from the default folder
	if !te {
		fmt.Println("template file", templatePath, "does not exists")
		return
	}

	// render invoice
	if invoiceNumber, err := govoice.RenderInvoice(password, templatePath); err == govoice.InvoiceDescriptorExists {
		fmt.Println("ok, nothing to do")
	} else if err != nil {
		fmt.Println("error rendering invoice:", err)
	} else {
		path, _ := config.GetInvoicePdfPath(invoiceNumber)
		fmt.Println("rendered invoice number", invoiceNumber, "at", path)
		open.Run(path)
	}
}
