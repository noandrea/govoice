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
	"github.com/spf13/viper"
	gv "gitlab.com/almost_cc/govoice/invoice"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore INVOICE_NUMBER",
	Short: "restore a generated (and ecrypted) invoice descriptor to the master descriptor for editing",
	Long: `this command should be used when it's necessary to edit an already rendered invoice
the command will overwrite the master descriptor with the content of the invoice found`,
	Run: restore,
}

func init() {
	RootCmd.AddCommand(restoreCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// restoreCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// restoreCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func restore(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		fmt.Println(cmd.Name(), "requires parameter INVOICE_NUMBER")
		cmd.Help()
		return
	}

	var c gv.Config
	viper.Unmarshal(&c)

	invoiceNumber := args[0]

	// if the invoice does not exists stop it
	_, e := c.GetInvoiceJsonPath(invoiceNumber)
	if !e {
		fmt.Println("invoice ", invoiceNumber, "does not exist in workspace")
		return
	}

	// read user password for decrypt
	password, err := gv.ReadUserPassword("Enter password:")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = gv.RestoreInvoice(&c, invoiceNumber, password)
	if err != nil {
		fmt.Println(err)
	}

}
