// Copyright © 2020 NAME HERE <EMAIL ADDRESS>
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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/koderizer/arc/model"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var extraAPIs []string
var outFile string

//ManifestData represent key extracts of the k8s manifest
type ManifestData struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// helmmapCmd represents the hemlmap command
var helmmapCmd = &cobra.Command{
	Use:   "helmmap",
	Short: "Map up all helm chart in the target directory into an .arc file",
	Long: `This command assist to initialize the structure of a microservice system
that have been built and deployed using helm-chart. 

Given no argument, the command will default to maping current directory and write to an arc.yaml. 
The default behavior is recursively look at all sub directories to profile
`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		var path string
		if len(args) > 0 {
			path = args[0]
		} else {
			var err error
			path, err = os.Getwd()
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("Mapping application at root: ", path)
		files := getHelms(path)
		arcData := model.ArcType{}
		arcData.App = filepath.Base(path)
		arcData.Desc = ToFillNotice
		arcData.InternalSystems = make([]model.InternalSystem, 0)
		arcData.ExternalSystems = make([]model.ExternalSystem, 0)

		actionConfig := &action.Configuration{
			Releases:     storage.Init(driver.NewMemory()),
			KubeClient:   &kubefake.PrintingKubeClient{Out: ioutil.Discard},
			Capabilities: chartutil.DefaultCapabilities,
			Log:          func(format string, v ...interface{}) {},
		}
		client := action.NewInstall(actionConfig)
		client.DryRun = true
		client.Replace = true // Skip the name check
		client.ClientOnly = true
		client.APIVersions = chartutil.VersionSet(extraAPIs)

		for _, f := range files {
			name := filepath.Base(f)
			client.ReleaseName = name
			settings := cli.New()
			cp, err := client.ChartPathOptions.LocateChart(f, settings)
			if err != nil {
				fmt.Println(err)
				return
			}
			p := getter.All(settings)
			valueOpts := &values.Options{}
			vals, err := valueOpts.MergeValues(p)
			if err != nil {
				fmt.Println(err)
				return
			}

			// Check chart dependencies to make sure all are present in /charts
			chartRequested, err := loader.Load(cp)
			if err != nil {
				fmt.Println(err)
				return
			}

			if req := chartRequested.Metadata.Dependencies; req != nil {
				// If CheckDependencies returns an error, we have unfulfilled dependencies.
				// As of Helm 2.4.0, this is treated as a stopping condition:
				// https://github.com/helm/helm/issues/2209
				if err = action.CheckDependencies(chartRequested, req); err != nil {
					if client.DependencyUpdate {
						man := &downloader.Manager{
							Out:              os.Stdout,
							ChartPath:        cp,
							Keyring:          client.ChartPathOptions.Keyring,
							SkipUpdate:       false,
							Getters:          p,
							RepositoryConfig: settings.RepositoryConfig,
							RepositoryCache:  settings.RepositoryCache,
							Debug:            settings.Debug,
						}
						if err = man.Update(); err != nil {
							fmt.Println(err)
							return
						}
						// Reload the chart with the updated Chart.lock file.
						if chartRequested, err = loader.Load(cp); err != nil {
							fmt.Println(err)
							return
						}
					} else {
						fmt.Println(err)
						return
					}
				}
			}

			client.Namespace = settings.Namespace()
			rel, err := client.Run(chartRequested, vals)
			if err != nil {
				fmt.Println(err)
				return
			}
			decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
			kubeObj, _, err := decode([]byte(rel.Manifest), nil, nil)
			if err != nil {
				fmt.Println(err)
				return
			}
			manifest := &ManifestData{}
			if err := yaml.Unmarshal([]byte(rel.Manifest), manifest); err != nil {
				fmt.Println(err)
				return
			}
			// fmt.Println(rel.Manifest)
			switch manifest.APIVersion + "." + manifest.Kind {
			case "extensions/v1beta1.Deployment":
				deployment := kubeObj.(*v1beta1.Deployment)
				system := model.InternalSystem{
					Name:       deployment.GetName(),
					Desc:       cp + deployment.GetSelfLink(),
					Containers: make([]model.Container, 0),
				}

				for _, c := range deployment.Spec.Template.Spec.Containers {
					system.Containers = append(system.Containers, model.Container{
						Name:       c.Name,
						Desc:       ToFillNotice,
						Technology: c.Image,
						Runtime:    manifest.Kind,
					})
				}
				arcData.InternalSystems = append(arcData.InternalSystems, system)
			case "v1.Service":
				service := kubeObj.(*v1.Service)
				extSystem := model.ExternalSystem{
					Name: service.GetName(),
					Desc: string(service.GetUID()),
				}
				arcData.ExternalSystems = append(arcData.ExternalSystems, extSystem)
			default:
				fmt.Println("Not supported")
			}
		}
		arcTpl, err := template.New("arcTemplate").Parse(arcTemplate)
		if err != nil {
			fmt.Println(err)
			return
		}

		arcYaml := []byte{}
		wr := bytes.NewBuffer(arcYaml)
		if err := arcTpl.ExecuteTemplate(wr, "arcTemplate", arcData); err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(wr.String())
		if err := ioutil.WriteFile(outFile, wr.Bytes(), os.ModeAppend); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	},
}

func getHelms(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil
	}
	results := make([]string, 0)
	if chart, err := chartutil.IsChartDir(dir); chart && err == nil {
		return append(results, dir)
	}
	for _, f := range files {
		absName := fmt.Sprintf("%s/%s", absPath, f.Name())
		if f.IsDir() {
			if chart, err := chartutil.IsChartDir(absName); chart && err == nil {
				results = append(results, absName)
			} else {
				results = append(results, getHelms(fmt.Sprintf("%s/%s", absPath, f.Name()))...)
				continue
			}
		}
	}
	return results
}

const (
	//ToFillNotice is the default description text
	ToFillNotice = "Autogen please fill this"
)

var arcTemplate = `
app: {{.App}}
desc: {{.Desc}}

internal-systems:
{{range .InternalSystems}}
- name: {{.Name}}
  desc: {{.Desc}}
  containers:
{{range .Containers}}
  - name: {{.Name}}
    desc: {{.Desc}}
    technology: {{.Technology}}
{{end}}
{{end}}
external-systems:
{{range .ExternalSystems}}
- name: {{.Name}}
  desc: {{.Desc}}
{{end}}
`
var sch *runtime.Scheme

func init() {
	sch = runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(sch)
	_ = apiextv1beta1.AddToScheme(sch)

	rootCmd.AddCommand(helmmapCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hemlmapCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hemlmapCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	extraAPIs = *helmmapCmd.PersistentFlags().StringArrayP("api-versions", "a", []string{}, "Kubernetes api versions used for Capabilities.APIVersions")
	outFile = *helmmapCmd.PersistentFlags().StringP("out", "o", "arc.yaml", "Output file to write")
}
