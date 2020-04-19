package model

import (
	"bytes"
	"encoding/gob"
	"log"
)

//go:generate protoc -I . --go_out=plugins=grpc:./ ./model.proto

//ArchUser represent a person who use some software
type ArchUser struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Desc string `yaml:"desc"`
}

//ArchInternalSystem represent a software system in the application
type ArchInternalSystem struct {
	Name       string          `yaml:"name"`
	Role       string          `yaml:"role"`
	Desc       string          `yaml:"desc"`
	Containers []ArchContainer `yaml:"containers"`
}

//ArchContainer represent a Container software runtime
type ArchContainer struct {
	Name       string          `yaml:"name"`
	Role       string          `yaml:"role"`
	Desc       string          `yaml:"desc"`
	Runtime    string          `yaml:"runtime"`
	Technology string          `yaml:"technology"`
	Components []ArchComponent `yaml:"components"`
}

//ArchComponent represent a Component that make up the implementation of a software running in a Container
type ArchComponent struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Desc string `yaml:"desc"`
	Code string `yaml:"code"`
}

//ArchExternalSystem represent an external software system
type ArchExternalSystem struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Desc string `yaml:"desc"`
}

//ArcType is the core data structure of a software architecture
type ArcType struct {
	App             string               `yaml:"app"`
	Desc            string               `yaml:"desc"`
	Users           []ArchUser           `yaml:"users"`
	InternalSystems []ArchInternalSystem `yaml:"internal-systems"`
	ExternalSystems []ArchExternalSystem `yaml:"external-systems"`
	Relations       []string             `yaml:"relations"`
}

//Decode struct to byte
func (a *ArcType) Decode(inData []byte) error {
	dec := gob.NewDecoder(bytes.NewBuffer(inData))
	if err := dec.Decode(a); err != nil {
		log.Printf("Fail to decode data: %v", err)
		return err
	}
	return nil
}

//Encode struct to byte
func (a *ArcType) Encode() ([]byte, error) {
	var out bytes.Buffer
	enc := gob.NewEncoder(&out)
	err := enc.Encode(a)
	if err != nil {
		log.Printf("Fail to encode data: %v", err)
		return nil, err
	}
	return out.Bytes(), nil
}
