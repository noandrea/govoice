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
	"github.com/skratchdot/open-golang/open"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit the master descriptor using the system editor",
	Long: `launch the default system editor opening the master descriptor for editing.

To open the descriptor with a specific application the --app (-a) is available.

Examples:
govoice edit  // uses the default json editor to open the master descriptor
govoice edit -a "Sublime Text"  // uses Sublime Text to open the master descriptor
`,
	Run: edit,
}

func init() {
	RootCmd.AddCommand(editCmd)


	editCmd.Flags().StringP("app", "a", "", "open master using the specifica application")

}


func edit(cmd *cobra.Command, args []string) {

	var c gv.Config
	var err error
	// load configuration
	viper.Unmarshal(&c)
	// get the master path
	mp,e := c.GetMasterPath()
	if !e {
		fmt.Println("master path at ", mp, "does not exists! run govoice config to restore it")
	}


	app,_ := cmd.Flags().GetString("app")
	if app == ""{
		err = open.Run(mp)
	} else {
		err = open.RunWith(mp, app)
	}

	if err != nil{
		fmt.Println("error opening master descriptor", err)
		return
	}
	fmt.Println("run 'govoice render' when done editing to create the invoice pdf")
}