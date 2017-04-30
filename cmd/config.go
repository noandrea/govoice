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

	"github.com/spf13/cobra"
	gv "gitlab.com/almost_cc/govoice/invoice"
	"os"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "configure govoice",
	Long: `the config command should be used once for installation
`,
	Run: config,
}

func init() {
	RootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	defaultWorkspace := fmt.Sprintf("%s/govoice", os.Getenv("HOME"))

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	configCmd.PersistentFlags().String("workspace", defaultWorkspace, "set the workspace of govoice, default to "+defaultWorkspace)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func config(cmd *cobra.Command, args []string) {

	workspace, _ := cmd.LocalFlags().GetString("workspace")
	// create configuration with defaults
	c := gv.Config{
		Workspace:      workspace,
		Encrypt:        true,
		MasterTemplate: "_master",
		Layout: gv.Layout{
			Style:    gv.Style{gv.Margins{0, 20, 20, 10}, "helvetica", 8, 14, 16, 6, 3.7, 6, 4, 3, 60, 13, 13, 13, 8, 6},
			Items:    gv.Block{gv.Coords{-1, 100}},
			From:     gv.Block{gv.Coords{-1, 28}},
			To:       gv.Block{gv.Coords{-1, 60}},
			Invoice:  gv.Block{gv.Coords{140, 28}},
			Payments: gv.Block{gv.Coords{-1, 210}},
			Notes:    gv.Block{gv.Coords{-1, 240}},
		},
	}

	// write default configuration file
	configPath := gv.GetConfigFilePath()
	gv.WriteTomlToFile(configPath, c)
	fmt.Println("config created at ", configPath)

	// write internationalization file
	en := gv.Translation{
		From:               gv.I18NOther{"FROM"},
		Sender:             gv.I18NOther{"{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\nTax Number: {{.TaxId}}\nVAT: {{.VatNumber}}"},
		To:                 gv.I18NOther{"TO"},
		Recipient:          gv.I18NOther{"{{.Name}}\n{{.Address}}\n{{.AreaCode}}, {{.City}}\n{{.Country}}\nTax Number: {{.TaxId}}\nVAT: {{.VatNumber}}"},
		Invoice:            gv.I18NOther{"INVOICE"},
		InvoiceData:        gv.I18NOther{"N. {{.Number}}\nDate: {{.Date}}\nDue:{{.Due}}"},
		PaymentDetails:     gv.I18NOther{"PAYMENTS DETAILS"},
		PaymentDetailsData: gv.I18NOther{"{{.AccountHolder}}\n\nBank: {{.Bank}}\nIBAN: {{.Iban}}\nBIC: {{.Bic}}"},
		Notes:              gv.I18NOther{"NOTES"},
		Desc:               gv.I18NOther{"Description"},
		Quantity:           gv.I18NOther{"Quantity"},
		Rate:               gv.I18NOther{"Rate"},
		Cost:               gv.I18NOther{"Cost"},
		Subtotal:           gv.I18NOther{"Subtotal"},
		Total:              gv.I18NOther{"Total"},
		Tax:                gv.I18NOther{"VAT"},
	}
	enPath := gv.GetI18nTranslationPath("en")
	gv.WriteTomlToFile(enPath, en)

	// write master.json
	// create the config directory if not exists
	_ = os.Mkdir(workspace, os.FileMode(0770))
	masterPath, exists := c.GetMasterPath()
	// don't overwrite the master if already exists
	if !exists {
		master := gv.Invoice{
			From:           gv.Recipient{"Mathis Hecht", "880 Whispering Half", "Hamburg", "67059", "Deutsheland", "9999999", "DE99999999999", "mh@ex.com"},
			To:             gv.Recipient{"Encarnicion Tellez Nino", "Calle Burch No. 139", "Valencia", "19490", "España", "55555555", "ES55555555", "etn@bs.com"},
			PaymentDetails: gv.BankCoordinates{"Mathis Hecht", "B Bank", "DE 1111 1111 1111 1111 11", "XXXXXXXX"},
			Invoice:        gv.InvoiceData{"0000000", "01.01.2017", "01.02.2017"},
			Settings:       gv.InvoiceSettings{45, 19, "€", "", "en"},
			Dailytime:      gv.Daily{Enabled: false},
			Items:          &[]gv.Item{gv.Item{"web dev", 10, 0}, gv.Item{"training", 5, 60}},
			Notes:          []string{"first note", "second note"},
		}
		gv.WriteJsonToFile(masterPath, master)
		fmt.Println("master invoice created", masterPath)
	}

	fmt.Println("master invoice is ", masterPath)
}
