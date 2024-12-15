package main

// #cgo CFLAGS: -g -Wall -Iinclude
import (
	"C"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sarpt/openapi-utils/pkg/openapi"
)

const (
	NoError           = 0
	InputFilepathErr  = 11
	InputStdinErr     = 12
	RefDirCwdErr      = 13
	OutputFilepathErr = 21
	OutputWriteErr    = 22
	OutputStdoutErr   = 23
	RootDocumentErr   = 31
	RefResolveErr     = 32
)

//export oasYamlCombine
func oasYamlCombine(inputFilePath *C.char, outputFilePath *C.char, refDirPath *C.char, inlineLocalRefs C.int, inlineRemoteRefs C.int, keepLocalRefs C.int) C.int {
	inputFile := C.GoString(inputFilePath)
	outputFile := C.GoString(outputFilePath)
	refDirectory := C.GoString(refDirPath)

	rootCfg := openapi.Config{
		InlineLocalRefs:  inlineLocalRefs == 1,
		InlineRemoteRefs: inlineRemoteRefs == 1,
		KeepLocalRefs:    keepLocalRefs == 1,
	}

	rootDocument := openapi.NewDocument(rootCfg)

	if inputFile != "" {
		inputFilePath, err := filepath.Abs(inputFile)
		if err != nil {
			log.Printf("Could not parse input file path: %v", err)
			return InputFilepathErr
		}

		err = rootDocument.ReadFile(inputFilePath)
		if err != nil {
			log.Printf("Error while parsing the root document: %v", err)
			return RootDocumentErr
		}
	} else {
		err := rootDocument.Read(os.Stdin)
		if err != nil {
			log.Printf("Error while reading from standard input: %v", err)
			return InputStdinErr
		}

		if refDirectory != "" {
			rootDocument.SetRefDirectory(refDirectory)
		} else {
			pwdRefDir, err := os.Getwd()
			if err != nil {
				log.Printf("Could not set reference directory to current working directory: %v", err)
				return RefDirCwdErr
			}

			rootDocument.SetRefDirectory(pwdRefDir)
		}
	}

	err := rootDocument.ResolveReferences()
	if err != nil {
		log.Printf("Error while resolving references in root document: %v", err)
		return RefResolveErr
	}

	if outputFile != "" {
		outputFilePath, err := filepath.Abs(outputFile)
		if err != nil {
			log.Printf("Could not parse output file path: %v", err)
			return OutputFilepathErr
		}

		err = rootDocument.WriteFile(outputFilePath)
		if err != nil {
			log.Printf("Error while writing output to path %s: %v", outputFilePath, err)
			return OutputWriteErr
		}

		fmt.Printf("Wrote output YAML file to %s", outputFilePath)
	} else {
		err := rootDocument.Write(os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write yaml to standard output: %v", err)
			return OutputStdoutErr
		}
	}

	return NoError
}

func main() {}
