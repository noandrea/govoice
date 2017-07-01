// Copyright © 2017 NAME HERE <is@almost.cc>
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
)

var GovoiceVersion string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print the version number of Govoice",
	Long:    `All software has versions. This is Govoice's`,
	Run:     printVersion,
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func printVersion(cmd *cobra.Command, args []string) {
	// print the current version
	fmt.Println("v", GovoiceVersion)
}
