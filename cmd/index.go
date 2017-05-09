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

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "(re)generate the searchable index of invoices",
	Long:  ``,
	Run:   index,
}

func init() {
	RootCmd.AddCommand(indexCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// indexCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// indexCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func index(cmd *cobra.Command, args []string) {

	var c gv.Config

	// parse configuration
	viper.Unmarshal(&c)

	// retrieve password
	password, err := gv.ReadUserPassword("Enter password:")
	if err != nil {
		fmt.Println(err)
	}
	// create the index
	count, elapsed, err := gv.RebuildSearchIndex(&c, &password)
	// if index creation is
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("indexed created with", count, "invoices in", elapsed)
}
