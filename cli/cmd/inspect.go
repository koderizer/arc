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
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/koderizer/arc/model"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var vizAddress string

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "inspect an architecture from an arc yaml config file",
	Long: `given the arc yaml config file, this command will ingest the content
and parse into an arc datastructure and trigger the gui to display the arch view 
This command require an access to a arc-viz server, default is set to on localhost at port 10000`,

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
		arc := &model.ArcType{}
		err = yaml.Unmarshal(arcFile, arc)
		if err != nil {
			fmt.Println("fail to parse yaml content")
			return
		}
		data, err := arc.Encode()
		if err != nil {
			fmt.Println("Fail to encode data")
			return
		}
		opts := grpc.WithInsecure()
		conn, err := grpc.Dial(vizAddress, opts)
		if err != nil {
			log.Printf("Unable to contact arc-viz server at %s with error: %+v", vizAddress, err)
			return
		}
		defer conn.Close()

		client := model.NewArcVizClient(conn)

		pngViz, err := client.Render(context.Background(), &model.RenderRequest{
			VisualFormat: model.ArcVisualFormat_PNG,
			DataFormat:   model.ArcDataFormat_ARC,
			Data:         data,
		})
		if err != nil {
			log.Printf("Fail to render with error: %+v", err)
			return
		}
		png := base64.StdEncoding.EncodeToString(pngViz.GetData())

		log.Printf("<img src=\"data:image/png;base64,%s\" />", png)
		return
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	inspectCmd.PersistentFlags().StringVar(&vizAddress, "viz", "localhost:10000", "URI of an acr-viz service (default to localhost:10000)")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// inspectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
