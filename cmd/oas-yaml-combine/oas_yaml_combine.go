package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/sarpt/openapi-utils/pkg/openapi"
)

var inputFile *string
var outputFile *string

func init() {
	inputFile = flag.String("input-file", "", "path to the input yaml file to be processed")
	outputFile = flag.String("output-file", "", "path to the output yaml file")
	flag.Parse()
}

func main() {
	if *inputFile == "" || *outputFile == "" {
		flag.PrintDefaults()
		return
	}

	inputFilePath, err := filepath.Abs(*inputFile)
	if err != nil {
		log.Fatalf("Could not parse input file path: %v", err)
	}

	outputFilePath, err := filepath.Abs(*outputFile)
	if err != nil {
		log.Fatalf("Could not parse output file path: %v", err)
	}

	rootDocument := openapi.NewDocument(inputFilePath)

	err = rootDocument.ParseFile()
	if err != nil {
		log.Fatalf("Error while parsing the root document: %v", err)
	}

	err = rootDocument.ResolveReferences()
	if err != nil {
		log.Fatalf("Error while resolving references in root document: %v", err)
	}

	err = rootDocument.WriteFile(outputFilePath)
	if err != nil {
		log.Fatalf("Error while writing output to path %s: %v", outputFilePath, err)
	}

	log.Printf("Wrote output YAML file to %s", outputFilePath)
}
