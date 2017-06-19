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
	//DB driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var config struct {
	BindAddr string
	Debug    bool
	Syslog   string
	Dsn      string
	Pprof    string
	NewRelic string
	NoCache  bool
	Trace    bool
	Logz     string
}

var cfgFile string

// RootCmd for pot
var RootCmd = &cobra.Command{
	Use:   "passwd-pot",
	Short: "",
	Long:  "",
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
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.passwd-pot.yaml)")
	RootCmd.PersistentFlags().BoolVar(&config.Debug, "debug", false, "debug mode")
	RootCmd.PersistentFlags().StringVar(&config.Syslog, "syslog", "", "use syslog server")
	RootCmd.PersistentFlags().StringVar(&config.Logz, "logz", "", "key for logz.io")
	RootCmd.PersistentFlags().StringVar(&config.Pprof, "pprof", "", "pprof endpoint (localhost:6060)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".passwd-pot") // name of config file (without extension)
	viper.AddConfigPath("$HOME")       // adding home directory as first search path
	viper.AutomaticEnv()               // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
