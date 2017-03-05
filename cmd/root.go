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
	log "github.com/Sirupsen/logrus"
	//DB driver
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/dougEfresh/dbr.v2"
	"os"
	"strings"
)

var config struct {
	BindAddr string
	Debug bool
	Syslog string
	Dsn string
}

var cfgFile string

// RootCmd for ssh pot
var RootCmd = &cobra.Command{
	Use:   "ssh-audit",
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
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ssh-audit-geo.yaml)")
	RootCmd.PersistentFlags().StringVar(&config.Dsn,"dsn", "", "DSN database url")
	RootCmd.PersistentFlags().StringVar(&config.BindAddr, "bind", "127.0.0.1:8080", "bind to this address:port")
	RootCmd.PersistentFlags().StringVar(&config.Syslog, "syslog", "localhost", "use syslog server")
	RootCmd.PersistentFlags().BoolVar(&config.Debug, "debug", false, "Enable Debug")

	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".ssh-audit-geo") // name of config file (without extension)
	viper.AddConfigPath("$HOME")          // adding home directory as first search path
	viper.AutomaticEnv()                  // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func loadDSN(dsn string) *dbr.Connection {
	var db *dbr.Connection
	var err error
	if strings.Contains(dsn, "mysql") {
		log.Debug("Using mysql driver")
		db, err = dbr.Open("mysql", dsn, dbEventLogger)
	} else {
		log.Debug("Using pq driver")
		db, err = dbr.Open("postgres", dsn, dbEventLogger)
	}

	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}
