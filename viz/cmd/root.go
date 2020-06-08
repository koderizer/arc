/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/koderizer/arc/model"
	"github.com/koderizer/arc/viz/server"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/spf13/viper"
)

type rootConfig struct {
	Port         string `mapstructure:"port"`
	PlantUmlAddr string `mapstructure:"PUML_ADDR"`
}

var cfgFile string
var config = &rootConfig{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "viz",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Implement ArcVizServer
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("%#v", config)
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Port))
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		grpcServer := grpc.NewServer()
		model.RegisterArcVizServer(grpcServer, server.NewArcViz(config.PlantUmlAddr))
		log.Printf("Start server listening on port: %s\n depending on Plantuml at %s", config.Port, config.PlantUmlAddr)
		grpcServer.Serve(lis)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.viz.yaml)")

	rootCmd.PersistentFlags().StringVar(&config.Port, "port", "10000", "listening port(default is 10000)")
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	rootCmd.PersistentFlags().StringVar(&config.PlantUmlAddr, "pumladdr", "http://localhost:8080", "Address of the Plant UML server(default is on localhost)")
	viper.BindPFlag("PUML_ADDR", rootCmd.PersistentFlags().Lookup("pumladdr"))

	viper.AutomaticEnv() // read in environment variables that match
	err := viper.Unmarshal(&config)
	log.Println(viper.GetString("PUML_ADDR"))
	if err != nil {
		log.Println(errors.Wrap(err, "unmarshal config file"))
	}
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".viz" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".viz")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
