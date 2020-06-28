package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sarpt/openapi-utils/pkg/openapi"
)

var (
	inputFile       *string
	outputFile      *string
	refDirectory    *string
	inlineLocalRefs *bool
	keepLocalRefs   *bool
)

func init() {
	inputFile = flag.String("input-file", "", "path to the input yaml file to be processed. Providing input-file sets the ref directory to the parent directory of provided input-file path")
	outputFile = flag.String("output-file", "", "path to the output yaml file")
	refDirectory = flag.String("ref-dir", "", "directory used as a root for ref relative paths resolution. By default current working directory is used, unless the input-file is provided")
	inlineLocalRefs = flag.Bool("inline-local", false, "should local refs be inlined in place when resolved. When set to false, local references are left in place since they are skipped from resolving. False by default")
	keepLocalRefs = flag.Bool("keep-local", false, "keep local refs after inlining. Makes sense only when inline-local is specified as true, otherwise has no effect in order to prevent outputting incorrect yaml file with missing references")
	flag.Parse()
}

func main() {
	rootCfg := openapi.Config{
		InlineLocalRefs: *inlineLocalRefs,
		KeepLocalRefs:   *keepLocalRefs,
	}

	rootDocument := openapi.NewDocument(rootCfg)
	if *inputFile != "" {
		inputFilePath, err := filepath.Abs(*inputFile)
		if err != nil {
			log.Fatalf("Could not parse input file path: %v", err)
		}

		err = rootDocument.ReadFile(inputFilePath)
		if err != nil {
			log.Fatalf("Error while parsing the root document: %v", err)
		}
	} else {
		err := rootDocument.Read(os.Stdin)
		if err != nil {
			log.Fatalf("Error while reading from standard input: %v", err)
		}

		if *refDirectory != "" {
			rootDocument.SetRefDirectory(*refDirectory)
		} else {
			pwdRefDir, err := os.Getwd()
			if err != nil {
				log.Fatalf("Could not set reference directory to current working directory: %v", err)
			}

			rootDocument.SetRefDirectory(pwdRefDir)
		}
	}

	err := rootDocument.ResolveReferences()
	if err != nil {
		log.Fatalf("Error while resolving references in root document: %v", err)
	}

	if *outputFile != "" {
		outputFilePath, err := filepath.Abs(*outputFile)
		if err != nil {
			log.Fatalf("Could not parse output file path: %v", err)
		}

		err = rootDocument.WriteFile(outputFilePath)
		if err != nil {
			log.Fatalf("Error while writing output to path %s: %v", outputFilePath, err)
		}

		fmt.Printf("Wrote output YAML file to %s", outputFilePath)
	} else {
		err := rootDocument.Write(os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write yaml to standard output: %v", err)
		}
	}
}
