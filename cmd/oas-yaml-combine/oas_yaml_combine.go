package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sarpt/openapi-utils/pkg/openapi"
)

const filepath string = "../../examples/test.yaml"
const outfile string = "out.yaml"

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error while trying to open user home directory")
	}

	outPath := fmt.Sprintf("%s/%s", homeDir, outfile)
	rootDocument := openapi.NewDocument(filepath)

	err = rootDocument.ParseFile()
	if err != nil {
		log.Fatalf("Error while parsing the root document: %v", err)
	}

	err = rootDocument.ResolveReferences()
	if err != nil {
		log.Fatalf("Error while resolving references in root document: %v", err)
	}

	err = rootDocument.WriteFile(outPath)
	if err != nil {
		log.Fatalf("Error while writing output to path %s: %v", outPath, err)
	}

	log.Printf("Wrote output YAML file to %s", outPath)
}
