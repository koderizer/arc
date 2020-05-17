/*
Copyright © 2020 Koderizer

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
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"runtime"

	"github.com/koderizer/arc/model"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var vizAddress string
var arcFilename string

//defaultArcFile point to the arc.yaml in the current directory arcli run
const defaultArcFile = "./arc.yaml"

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect <perspective> <targets>",
	Short: "Inspect an architecture from an arc yaml config file in the current source",
	Long: `

Given the arc yaml config file explicitly with -f option, this command will ingest the content
and parse into an arc data structure and trigger your browser to display the architecture view requested

inspect take in 0 to n arguments, as follow:
 - without any argument and option, inspect try to render a Landscape perspective for the software application specified in arc.yaml in the current directory
 - with one argument, inspect try to render all elements that can be viewed at any one of the following perspective: landscape,context,container,component
 - with 2 or more arguments, inspect try to render the perspective specified by first argument for all elements given from the second arguments onward.

Eg: 
To render the Container perspective of your amazingSystem1 and amazingSystem2 as specified in an arc.yaml file in the current directory

	arcli inspect container amazingSystem1 amazingSystem2`,

	Run: func(cmd *cobra.Command, args []string) {
		arcFile, err := ioutil.ReadFile(arcFilename)
		if err != nil {
			fmt.Println("fail to read arc yaml file")
			return
		}
		arc := &model.ArcType{}
		err = yaml.Unmarshal(arcFile, arc)
		if err != nil {
			fmt.Println("fail to parse yaml content", err)
			return
		}
		targets := make([]string, 0)
		var pers model.PresentationPerspective = model.PresentationPerspective_LANDSCAPE
		if len(args) > 0 {
			switch args[0] {
			case "context":
				pers = model.PresentationPerspective_CONTEXT
			case "container":
				pers = model.PresentationPerspective_CONTAINER
			case "component":
				pers = model.PresentationPerspective_COMPONENT
			case "code":
				log.Println("Not supported for now. Make simple readable code")
				return
			case "landscape":
				pers = model.PresentationPerspective_LANDSCAPE
			default:
				log.Printf("Perspective %s not supported, please indicate one of: landscape, context, container, component", args[0])
				return
			}
		}
		if len(args) > 1 {
			targets = args[1:]
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
			Target:       targets,
			Perspective:  pers,
		})
		if err != nil {
			log.Printf("Fail to render with error: %+v", err)
			return
		}
		// png := base64.StdEncoding.EncodeToString(pngViz.GetData())

		// htm := fmt.Sprintf("<img src=\"data:image/png;base64,%s\" />", png)
		// err = ioutil.WriteFile("./index.html", []byte(htm), 0644)
		// if err != nil {
		// 	log.Printf("Fail to write output %s", err)
		// 	return
		// }
		// http.Handle("/", http.FileServer(http.Dir("./")))
		uri := string(pngViz.GetData())
		log.Println(uri)
		open(uri)
		// panic(http.ListenAndServe(":10001", nil))
	},
}

// open opens the specified URL in the default browser of the user.
func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func init() {
	rootCmd.AddCommand(inspectCmd)

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	inspectCmd.PersistentFlags().StringVar(&vizAddress, "viz", "localhost:10000", "URI of an acrviz app")
	inspectCmd.PersistentFlags().StringVarP(&arcFilename, "file", "f", defaultArcFile, "Path to the arc.yaml file to inspect")

}
