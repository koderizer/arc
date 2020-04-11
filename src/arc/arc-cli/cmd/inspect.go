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
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

//archModel is the core data structure of a software architecture
type archModel struct {
	App   string `yaml:"app"`
	Desc  string `yaml:"desc"`
	Users []struct {
		Name string `yaml:"name"`
		Desc string `yaml:"desc"`
	} `yaml:"users"`
	InternalSystems []struct {
		Name       string `yaml:"name"`
		Desc       string `yaml:"desc"`
		Containers []struct {
			Name       string `yaml:"name"`
			Desc       string `yaml:"desc"`
			Runtime    string `yaml:"runtime"`
			Technology string `yaml:"technology"`
		} `yaml:"containers"`
	} `yaml:"internal-systems"`
	ExternalSystems []struct {
		Name string `yaml:"name"`
		Desc string `yaml:"desc"`
	}
	Relations []string `yaml:"relations"`
}

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "inspect an architecture from an arc yaml config file",
	Long: `given the arc yaml config file, this command will ingest the content
and parse into an arc datastructure and trigger the arc-gui to display the arch view `,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("inspect called")
		if len(args) == 0 {
			fmt.Println("missing input arc yaml file")
			return
		}
		arcFile, err := ioutil.ReadFile(args[0])
		if err != nil {
			fmt.Println("fail to read arc yaml file")
			return
		}
		arc := new(archModel)
		err = yaml.Unmarshal(arcFile, arc)
		if err != nil {
			fmt.Println("fail to parse yaml content")
			return
		}
		fmt.Printf("%+v\n", arc)
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// inspectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inspectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
