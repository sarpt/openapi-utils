package openapi

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

const (
	// RootItem is an OpenAPI Root
	RootItem = "Root"
	// Ref is a reference field
	Ref = "Ref"
	// YamlTag is a tag key which is used to parse YAML into internal representation
	YamlTag = "yaml"
	// YamlTagSeparator is a symbol which separates YAML key in tag from flags
	YamlTagSeparator = ","
)

// Document represents single OpenAPI source file and it's content.
// A Document can be dependent on other Documents by using OpenAPI references.
type Document struct {
	Cfg                 Config
	RefDirectory        string
	FileName            string
	Root                *OpenAPI
	ReferencedDocuments map[string]*Document
}

// reference contains information about OpenAPI object that contains reference and path of reference
type reference struct {
	object OasObject
	path   string
}

// Config specifies document handling
type Config struct {
	InlineLocalRefs bool
	KeepLocalRefs   bool
}

// NewDocument constructs new Document instance
func NewDocument(cfg Config) Document {
	return Document{
		Cfg:                 cfg,
		Root:                &OpenAPI{},
		ReferencedDocuments: make(map[string]*Document),
	}
}

// ParseDocument takes path to the file that should be parsed and have it's references resolved
func ParseDocument(cfg Config, path string) (Document, error) {
	referencedDocument := NewDocument(cfg)

	err := referencedDocument.ReadFile(path)
	if err != nil {
		return Document{}, err
	}

	err = referencedDocument.ResolveReferences()
	return referencedDocument, err
}

// Parse unmarshalls the yaml content
func (doc Document) Parse(data []byte) error {
	return yaml.Unmarshal(data, doc.Root)
}

// Read takes a Reader and parses the content after encountering EOF
func (doc Document) Read(r io.Reader) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return doc.Parse(data)
}

// ReadFile attempts to read & parse content of file Document points to
func (doc *Document) ReadFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	doc.RefDirectory = filepath.Dir(path)
	doc.FileName = filepath.Base(path)

	err = yaml.Unmarshal(data, doc.Root)
	return err
}

// WriteFile writes content of a document to a YAML file pointed by path
func (doc Document) WriteFile(path string) error {
	yaml, err := doc.YAML()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, yaml, os.FileMode(0777))
}

// Write writes content of a document to a writer
func (doc Document) Write(w io.Writer) error {
	yaml, err := doc.YAML()
	if err != nil {
		return err
	}

	_, err = w.Write(yaml)
	return err
}

// YAML converts contents of a document to YAML
func (doc Document) YAML() ([]byte, error) {
	return yaml.Marshal(doc.Root)
}

// SetRefDirectory sets the directory which is used as root for refs relative paths resolution
func (doc *Document) SetRefDirectory(dir string) {
	doc.RefDirectory = dir
}

// ResolveReferences takes a document and tries to find and resolve all references.
// After execution all elements that had not empty Ref properties have their contents replaced with referenced content.
// References are first sorted before resolution/assignment due to use-case where local reference aliases remote one.
// TODO:
// Improve handling of inlineLocal when set to false - local references that were unresolved in referenced documents should
// be modified (path in $ref changed) in the document being resolved to point to the referenced document.
// Currently when resolving references to the remote documents, their local references end up in the document being resolved.
// This behavior ends up creating incorrect output file.
// Current workaround is that referenced documents are parsed with different inlineLocal config from the root document.
// It's fine for now, but ideally referenced documents should resue config of parent.
func (doc Document) ResolveReferences() error {
	rootObject, err := NewOasObjectByName(&doc, RootItem)
	if err != nil {
		return err
	}

	refs, err := rootObject.references()
	if err != nil {
		return err
	}

	sort.Slice(refs, func(i, j int) bool {
		return sortReferences(refs[i], refs[j])
	})

	for _, ref := range refs {
		err := doc.replaceReference(ref)
		if err != nil {
			return err
		}
	}

	return nil
}

// replaceReference replaces ref either with inline object or a local reference
func (doc Document) replaceReference(ref reference) error {
	isRefLocal := isLocalReference(ref.path)
	if !doc.Cfg.InlineLocalRefs && isRefLocal {
		return nil
	}

	referencedDocument, err := doc.getReferencedDocument(ref.path)
	if err != nil {
		return fmt.Errorf("could not get reference document: %w", err)
	}

	refObject, err := referencedDocument.getReferencedObjectByPath(ref.path)
	if err != nil {
		return err
	}

	err = ref.object.Set(refObject.instance)
	if err != nil {
		return err
	}

	if doc.Cfg.KeepLocalRefs || !doc.Cfg.InlineLocalRefs || !isRefLocal {
		return nil
	}

	return refObject.Set(nil)
}

// getReferencedValueByPath walks the provided reference path, trying obtain the oas object
func (doc Document) getReferencedObjectByPath(refPath string) (OasObject, error) {
	var object OasObject
	var err error

	itemNames := referencePathToItems(refPath)
	var parentValue reflect.Value = reflect.ValueOf(&doc.Root).Elem()

	for _, itemName := range itemNames {
		switch parentValue.Kind() {
		case reflect.Ptr:
			childItemName, err := getFieldNameByTag(itemName, parentValue.Elem())
			if err != nil {
				return object, fmt.Errorf("could not find item %s in path %s: %w", itemName, refPath, err)
			}

			object, err = NewOasObjectByName(parentValue.Interface(), childItemName)
			if err != nil {
				return object, err
			}

			parentValue = reflect.ValueOf(object.instance)
		case reflect.Map:
			object, err = NewOasObjectByName(parentValue.Interface(), itemName)
			if err != nil {
				return object, err
			}

			parentValue = reflect.ValueOf(object.instance)
		default:
			return object, fmt.Errorf("could not resolve path %s due to path including incorrect items", refPath)
		}
	}

	return object, nil
}

func (doc Document) getReferencedDocument(refPath string) (*Document, error) {
	var referencedDocument *Document

	if isLocalReference(refPath) {
		return &doc, nil
	}

	documentPath := getDocumentPath(refPath)
	documentFilePath := filepath.Join(doc.RefDirectory, documentPath)

	if document, ok := doc.ReferencedDocuments[documentFilePath]; ok {
		referencedDocument = document
	} else {
		cfg := Config{
			InlineLocalRefs: true,
		}
		parsedDocument, err := ParseDocument(cfg, documentFilePath)
		if err != nil {
			return nil, err
		}

		referencedDocument = &parsedDocument
		doc.ReferencedDocuments[documentFilePath] = referencedDocument
	}

	return referencedDocument, nil
}

func getFieldNameByTag(tag string, structItem reflect.Value) (string, error) {
	structItemType := structItem.Type()

	for i := 0; i < structItemType.NumField(); i++ {
		childField := structItemType.Field(i)
		yamlKey := getYamlKeyFromField(childField)
		if yamlKey == tag {
			return childField.Name, nil
		}
	}

	return "", fmt.Errorf("the field name with tag %s could not be found in type %s", tag, structItemType.Name())
}

func itemFromMapByName(mapVal reflect.Value, key string) (reflect.Value, reflect.Value, error) {
	mapIter := mapVal.MapRange()
	for mapIter.Next() {
		if mapIter.Key().String() == key {
			return mapIter.Key(), mapIter.Value(), nil
		}
	}

	return reflect.Value{}, reflect.Value{}, fmt.Errorf("could not find %s key in map", key)
}

func getYamlKeyFromField(field reflect.StructField) string {
	yamlTag := field.Tag.Get(YamlTag)

	return strings.Split(yamlTag, YamlTagSeparator)[0]
}
