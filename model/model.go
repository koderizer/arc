package model

import (
	"bytes"
	"encoding/gob"
	"log"
)

//go:generate protoc -I . --go_out=plugins=grpc:./ ./model.proto

//User represent a person who use some software
type User struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Desc string `yaml:"desc"`
}

//InternalSystem represent a software system in the application
type InternalSystem struct {
	Name       string      `yaml:"name"`
	Role       string      `yaml:"role"`
	Desc       string      `yaml:"desc"`
	Containers []Container `yaml:"containers"`
}

//Container represent a Container software runtime
type Container struct {
	Name       string      `yaml:"name"`
	Role       string      `yaml:"role"`
	Desc       string      `yaml:"desc"`
	Runtime    string      `yaml:"runtime"`
	Technology string      `yaml:"technology"`
	Components []Component `yaml:"components"`
}

//Component represent a Component that make up the implementation of a software running in a Container
type Component struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Desc string `yaml:"desc"`
	Code string `yaml:"code"`
}

//ExternalSystem represent an external software system
type ExternalSystem struct {
	Name string `yaml:"name"`
	Role string `yaml:"role"`
	Desc string `yaml:"desc"`
}

//ArcType is the core data structure of a software architecture
type ArcType struct {
	App             string           `yaml:"app"`
	Desc            string           `yaml:"desc"`
	Users           []User           `yaml:"users"`
	InternalSystems []InternalSystem `yaml:"internal-systems"`
	ExternalSystems []ExternalSystem `yaml:"external-systems"`
	Relations       []Relation       `yaml:"relations"`
}

//Relation represent a relationship path between different elements
type Relation struct {
	Subject string `yaml:"s"`
	Pointer string `yaml:"p"`
	Object  string `yaml:"o"`
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
