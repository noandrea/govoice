// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"github.com/spf13/viper"
	"gitlab.com/almost_cc/govoice/config"
	govoice "gitlab.com/almost_cc/govoice/invoice"
)

// previewCmd represents the preview command
var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "render a preview of the pdf",
	Long:  `works like render but doesn't asks for passwords and owerwrite existing files`,
	Run:   preview,
}

func init() {
	RootCmd.AddCommand(previewCmd)
	tp, _ := config.GetTemplatePath(config.DefaultTemplateName)
	help := fmt.Sprintln("template name or path, defaults to:", tp)
	previewCmd.PersistentFlags().StringVarP(&config.TemplateName, fname, "t", config.DefaultTemplateName, help)
	viper.BindPFlag(fname, previewCmd.PersistentFlags().Lookup(fname))
}

func preview(cmd *cobra.Command, args []string) {

	templatePath, te := config.GetTemplatePath(config.TemplateName)
	// if template path is a string, load the template from the default folder
	if !te {
		fmt.Println("template file", templatePath, "does not exists")
		return
	}
	fmt.Println("template is ", templatePath)
	// render invoice
	if invoiceNumber, err := govoice.PreviewInvoice(templatePath); err == govoice.InvoiceDescriptorExists {
		fmt.Println("ok, nothing to do")
	} else if err != nil {
		fmt.Println("error rendering invoice:", err)
	} else {
		path, _ := config.GetInvoicePdfPath(config.PreviewFileName)
		fmt.Println("preview invoice number", invoiceNumber, "at", path)
		open.Run(path)
	}
}
