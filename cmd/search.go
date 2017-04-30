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
	"strings"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"github.com/leekchan/accounting"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search QUERY",
	Short: "query the index to search for invoices",
	Long: `the indexing is made on the fields:
- Customer: customer name
- Date: invoice date
- Number: invoice number
- Amount: invoice subtotal

examples of queries are

govoice search "Amount:>1000" // search for invoices with amount greather than 1000
govoice search wolskwagen  // full text search on all field for wolkswagen

the full text search is provided by bleve, visit the bleve documentation for query examples

`,
	Run: search,
}

func init() {
	RootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func search(cmd *cobra.Command, args []string) {

	var c gv.Config

	// parse configuration
	viper.Unmarshal(&c)

	if len(args) == 0{
		fmt.Println("missing query string")
		return
	}

	entries, total, elapsed, err := gv.SearchInvoice(strings.Join(args, " "))
	if err != nil{
		fmt.Println(err)
		return
	}


	// output results to console as a table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetCenterSeparator("|")
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"File","Number","Customer","Date","Amount"})
	// for amount formatting
	ac := accounting.Accounting{Symbol:"€", Precision: 2}

	fmt.Println("found",total,"results in",elapsed)

	for _,e := range entries{
		path,_ := c.GetInvoicePdfPath(e.Number)
		table.Append([]string{
			path,
			e.Number,
			e.Customer,
			e.Date,
			ac.FormatMoney(e.Amount),
		})
	}
	// render the output
	table.Render()

}
