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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	gv "gitlab.com/almost_cc/govoice/invoice"

	"fmt"
	"github.com/skratchdot/open-golang/open"
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

}

func render(cmd *cobra.Command, args []string) {

	// parse configuration
	var c gv.Config
	viper.Unmarshal(&c)
	// read the password
	password, err := gv.ReadUserPassword("Enter password:")
	if err != nil {
		fmt.Println(err)
		return
	}

	// render invoice
	if invoiceNumber, err := gv.RenderInvoice(&c, password); err == gv.InvoiceDescriptorExists {
		fmt.Println("ok, nothing to do")
	} else if err != nil {
		fmt.Println("error rendering invoice:", err)
	} else {
		path, _ := c.GetInvoicePdfPath(invoiceNumber)
		fmt.Println("rendered invoice number", invoiceNumber, "at", path)
		open.Run(path)
	}

}
