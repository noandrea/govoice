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
	"path"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "configure govoice",
	Long:  `the config command should be used once for installation`,
	Run:   config,
}

func init() {
	RootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	defaultWorkspace := gv.GetConfigHome()

	// Cobra supports Persistent Flags which will work for this command
	// and all sub-commands, e.g.:
	configCmd.PersistentFlags().String("workspace", defaultWorkspace, "set the workspace of govoice, default to "+defaultWorkspace)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func config(cmd *cobra.Command, args []string) {

	workspace, _ := cmd.LocalFlags().GetString("workspace")
	if workspace == "" {
		workspace = path.Join(os.Getenv("HOME"), "GOVOICE")
	}

	cp, mp, err := gv.Setup(workspace)
	if err != nil {
		fmt.Println("configuration failed", err)
	}

	fmt.Println("configuration file created at    ", cp)
	fmt.Println("master descriptor file created at", mp)
}
