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
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gitlab.com/almost_cc/govoice/config"
)

// infoCmd represents the info command
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "print information about paths (when you forget where they are)",
	Long:  ``,
	Run:   info,
}

func init() {
	RootCmd.AddCommand(infoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// infoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// infoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func info(cmd *cobra.Command, args []string) {
	// print table in console
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	mp, _ := config.GetMasterPath()
	println()
	table.SetHeader([]string{"path", "desc"})
	table.Append([]string{gv.GetConfigHome(), "application home"})
	table.Append([]string{gv.GetConfigFilePath(), "config file"})
	table.Append([]string{gv.GetI18nHome(), "translation files"})
	table.Append([]string{c.Workspace, "workspace"})
	table.Append([]string{mp, "master descriptor"})
	table.Render()
	println()

}
