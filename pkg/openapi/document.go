package openapi

import (
	"io/ioutil"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Document struct {
	FilePath string
	References
	Root                OpenApi
	ReferencedDocuments map[string]Document
}

func (doc Document) FindReferences() {
	for path, pathSchema := range doc.Root.Paths {
		if pathSchema.Ref != "" {
			doc.PathReferences[pathSchema.Ref] = doc.Root.Paths[path]
			continue
		}

		for response, responseSchema := range pathSchema.Get.Responses {
			if responseSchema.Ref != "" {
				doc.ResponseReferences[responseSchema.Ref] = pathSchema.Get.Responses[response]
				continue
			}
		}
	}
}

func (doc *Document) ParseFile() error {
	data, err := ioutil.ReadFile(doc.FilePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &doc.Root)
	return err
}

func (doc *Document) ResolveReferences() error {
	for refPath, response := range doc.ResponseReferences {
		if isLocalReference(refPath) {
			continue
		}

		var referencedDocument Document
		pathToDocument := strings.Split(refPath, "#")[0]

		if document, ok := doc.ReferencedDocuments[pathToDocument]; ok {
			referencedDocument = document
		} else {
			referencedDocument = NewDocument(pathToDocument)

			err := referencedDocument.ParseFile()
			if err != nil {
				return err
			}

			referencedDocument.FindReferences()

			err = referencedDocument.ResolveReferences()
			if err != nil {
				return err
			}
		}

		responseComponentName := path.Base(refPath)
		*response = *referencedDocument.Root.Components.Responses[responseComponentName]
	}

	return nil
}

func isLocalReference(path string) bool {
	return strings.IndexRune(path, '#') == 0
}

func NewDocument(filePath string) Document {
	return Document{
		FilePath: filePath,
		References: References{
			PathReferences:     make(map[string]*Path),
			ResponseReferences: make(map[string]*Response),
		},
		Root:                OpenApi{},
		ReferencedDocuments: make(map[string]Document),
	}
}
