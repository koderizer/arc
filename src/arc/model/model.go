package model

//ArchUser represent a person who use some software
type ArchUser struct {
	Name string `yaml:"name"`
	Desc string `yaml:"desc"`
}

//ArchInternalSystem represent a software system in the application
type ArchInternalSystem struct {
	Name       string `yaml:"name"`
	Desc       string `yaml:"desc"`
	Containers []struct {
		Name       string `yaml:"name"`
		Desc       string `yaml:"desc"`
		Runtime    string `yaml:"runtime"`
		Technology string `yaml:"technology"`
	} `yaml:"containers"`
}

//ArchExternalSystem represent an external software system
type ArchExternalSystem struct {
	Name string `yaml:"name"`
	Desc string `yaml:"desc"`
}

//ArchType is the core data structure of a software architecture
type ArchType struct {
	App             string               `yaml:"app"`
	Desc            string               `yaml:"desc"`
	Users           []ArchUser           `yaml:"users"`
	InternalSystems []ArchInternalSystem `yaml:"internal-systems"`
	ExternalSystems []ArchExternalSystem `yaml:"external-systems"`
	Relations       []string             `yaml:"relations"`
}
