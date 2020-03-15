package openapi

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Document represents single OpenAPI source file and it's content
// A Document can be dependent on other Documents by using OpenAPI references
type Document struct {
	FilePath            string
	Root                *OpenAPI
	ReferencedDocuments map[string]*Document
}

// ResolveReferences takes a document and tries to find and resolve all references
// After execution all elements that had not empty Ref properties have their contents replaced with referenced content
func (doc Document) ResolveReferences() error {
	for _, pathSchema := range doc.Root.Paths {
		if pathSchema.Ref != "" {
			doc.replaceReference(pathSchema.Ref, pathSetter)
			continue
		}

		for _, responseSchema := range pathSchema.Get.Responses {
			if responseSchema.Ref != "" {
				doc.replaceReference(responseSchema.Ref, responseSetter)
				continue
			}
		}
	}

	return nil
}

func (doc Document) replaceReference(refPath string, setter func(elementName string, targetDocument Document, referenceDocument Document)) error {
	referencedDocument, err := doc.getReferencedDocument(refPath)
	if err != nil {
		return fmt.Errorf("Could not get reference document: %w", err)
	}

	elementName := path.Base(refPath)
	setter(elementName, doc, *referencedDocument)
	return nil
}

// ParseFile attempts to read & parse content of file Document points to
func (doc Document) ParseFile() error {
	data, err := ioutil.ReadFile(doc.FilePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), doc.Root)
	return err
}

func (doc Document) getReferencedDocument(refPath string) (*Document, error) {
	var referencedDocument *Document
	var err error

	if isLocalReference(refPath) {
		return &doc, nil
	}

	pathToDocument := getPathToRemoteDocument(refPath)

	if document, ok := doc.ReferencedDocuments[pathToDocument]; ok {
		referencedDocument = document
	} else {
		referencedDocument, err = ParseDocument(pathToDocument)
		if err != nil {
			return nil, err
		}
	}

	return referencedDocument, nil
}

func isLocalReference(path string) bool {
	return strings.IndexRune(path, '#') == 0
}

func getPathToRemoteDocument(path string) string {
	return strings.Split(path, "#")[0]
}

// ParseDocument takes path to the file that should be parsed and have it's references resolved
func ParseDocument(path string) (*Document, error) {
	referencedDocument := NewDocument(path)

	err := referencedDocument.ParseFile()
	if err != nil {
		return nil, err
	}

	err = referencedDocument.ResolveReferences()
	return &referencedDocument, err
}

// NewDocument constructs new Document instance
func NewDocument(filePath string) Document {
	return Document{
		FilePath:            filePath,
		Root:                &OpenAPI{},
		ReferencedDocuments: make(map[string]*Document),
	}
}
