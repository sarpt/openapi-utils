package main

import (
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"

	"github.com/sarpt/openapi-yaml-combine/pkg/openapi"
)

const filepath string = "../../examples/test.yaml"

func main() {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatalf("Error while reading source file: %v", err)
	}

	apiContent := openapi.OpenApi{}

	err = yaml.Unmarshal([]byte(data), &apiContent)
	if err != nil {
		log.Fatalf("Error while parsing YAML: %v", err)
	}
	fmt.Printf("%+v\n", apiContent)
}
