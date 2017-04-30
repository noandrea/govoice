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
	Use:   "restore",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
		fmt.Println(cmd.Name(), "requires one argument")
		cmd.Help()
		return
	}

	var c gv.Config
	var i gv.Invoice

	invoiceName := args[0]

	// parse configuration
	viper.Unmarshal(&c)

	// check if the invoice descriptor exists
	descriptorPath, exists := c.GetInvoiceJsonPath(invoiceName)
	if !exists {
		fmt.Println("Invoice ", invoiceName, " not found in ", c.Workspace)
		return
	}
	// read user password for decrypt
	password := gv.ReadUserPassword()
	// parse de invoice
	err := gv.ReadInvoiceDescriptorEncrypted(&descriptorPath, &i, &password)
	if err != nil{
		fmt.Println("invalid password, try again")
		return
	}
	// dump it on master descriptor
	masterDescriptorPath,_ := c.GetMasterPath()
	gv.WriteJsonToFile(masterDescriptorPath, i)
}