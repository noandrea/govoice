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
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/almost_cc/govoice/config"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "govoice",
	Short: "invoice generation an indexing for devz",
	Long: `Govoice is a tool to generate pdf invoices from a descriptor file.
It offers:
- encrypted invoice descriptor generator (to be stored on a VCS)
- indexing and search of invoices
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().BoolVar(&config.DebugEnabled, "debug", false, "enable debug log")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	// configuration folder
	configHome := config.GetConfigHome()
	// create the config directory if not exists
	_ = os.Mkdir(configHome, os.FileMode(0770))
	// search for config file
	viper.SetConfigName("config")   // name of config file (without extension)
	viper.AddConfigPath(configHome) // adding home directory as first search path
	viper.AutomaticEnv()
	// If a config file is found, read it in.
	viper.ReadInConfig()
	if err := viper.ReadInConfig(); err == nil {
		log.Println("d: Using config file:", viper.ConfigFileUsed())
		// load configurations (overwrited defautls)
		viper.Unmarshal(&config.Main)
		//log.Println("t: config", spew.Sdump(config.Db, config.Authority, config.RestAPI, config.Chats))
	} else {
		log.Fatalln("a: configuration file not found", err)
	}
}
