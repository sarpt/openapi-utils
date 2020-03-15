package main

import (
	"fmt"
	"log"

	"github.com/sarpt/openapi-yaml-combine/pkg/openapi"
)

const filepath string = "../../examples/test.yaml"

func main() {
	rootDocument := openapi.NewDocument(filepath)

	err := rootDocument.ParseFile()
	if err != nil {
		log.Fatalf("Error while parsing the root document: %v", err)
	}

	err = rootDocument.ResolveReferences()
	if err != nil {
		log.Fatalf("Error while resolving references in root document: %v", err)
	}

	fmt.Printf("%+v\n", rootDocument.Root.Paths["/companies"].Get.Responses["200"])
}
