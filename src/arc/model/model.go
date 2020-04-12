package model

//ArchType is the core data structure of a software architecture
type ArchType struct {
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
