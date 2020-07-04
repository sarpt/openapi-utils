package openapi

import (
	"errors"
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
	// RefTag is a tag which specifies symbol as a OpenAPI $ref reference valua
	RefTag = "$ref"
)

var (
	// ErrNoFieldWithTag informs that struct has no field/element (direct descendant) with specified tag
	ErrNoFieldWithTag = errors.New("could not find field with specified tag")
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

func (doc Document) replaceReference(ref reference) error { // method on reference instead on document? 'isLocal' could be calculated at creation time, or reference could be an interface that 'local' and 'remote' satisfy by implementing "replace". To be considered
	if !isLocalReference(ref.path) {
		return doc.replaceRemoteReference(ref)
	}

	if !doc.Cfg.InlineLocalRefs {
		return nil
	}

	return doc.replaceLocalReference(ref)
}

func (doc Document) replaceLocalReference(ref reference) error {
	referencedDocument, err := doc.getReferencedDocument(ref.path)
	if err != nil {
		return fmt.Errorf("could not get reference document: %w", err)
	}

	referencedObject, err := referencedDocument.getReferencedObjectByPath(ref.path)
	if err != nil {
		return err
	}

	err = ref.object.Set(referencedObject.instance)
	if err != nil {
		return err
	}

	if doc.Cfg.KeepLocalRefs || !doc.Cfg.InlineLocalRefs {
		return nil
	}

	return referencedObject.Unset()
}

func (doc Document) replaceRemoteReference(ref reference) error {
	referencedDocument, err := doc.getReferencedDocument(ref.path)
	if err != nil {
		return fmt.Errorf("could not get reference document: %w", err)
	}

	refObject, err := referencedDocument.getReferencedObjectByPath(ref.path)
	if err != nil {
		return err
	}

	localComponentsPath := convertRemoteToLocalPath(ref.path)
	componentsObject, err := doc.getOrCreatePath(localComponentsPath)
	if err != nil {
		return err
	}

	componentsObject.Set(refObject.instance)
	if err != nil {
		return err
	}

	return changeRefPath(ref.object, localComponentsPath)
}

func changeRefPath(o OasObject, newRefPath string) error {
	oasObjectStruct := reflect.ValueOf(o.instance).Elem()
	newRefPathValue := reflect.ValueOf(newRefPath)

	refFieldName, err := getFieldNameByTag(RefTag, oasObjectStruct)
	if err != nil {
		return err
	}

	refField := oasObjectStruct.FieldByName(refFieldName)
	refField.Set(newRefPathValue)

	return nil
}

// getReferencedValueByPath walks the provided reference path, trying obtain the oas object
func (doc Document) getReferencedObjectByPath(refPath string) (OasObject, error) {
	var object OasObject
	var err error

	itemNames := referencePathToItems(refPath)
	var parentValue reflect.Value = reflect.ValueOf(&doc.Root).Elem() // since parentValue is reused further with addressable values, the initializer has to addressable too (meaning, not a copy of a pointer to root)

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

// getOrCreatePath walks the provided reference path, trying obtain the oas object and creating it if it does not exists.
// TODO: this function is nearly identical to the getReferencedObjectByPath. Common function could be specified "walkPath" that takes callback.
// Or just paremtrize the function to force init of objects when necessary (blerh).
// I cannot be bothered to do this in this commit.
func (doc Document) getOrCreatePath(refPath string) (OasObject, error) {
	var object OasObject
	var err error

	itemNames := referencePathToItems(refPath)
	var parentValue reflect.Value = reflect.ValueOf(&doc.Root).Elem() // since parentValue is reused further with addressable values, the initializer has to addressable too (meaning, not a copy of a pointer to root)

	for _, itemName := range itemNames {
		switch parentValue.Kind() {
		case reflect.Ptr:
			childItemName, err := getFieldNameByTag(itemName, parentValue.Elem())
			if err != nil {
				return object, fmt.Errorf("could not find item %s in path %s: %w", itemName, refPath, err)
			}

			object, err = NewOasObjectByName(parentValue.Interface(), childItemName)
			if err != nil && !errors.Is(err, ErrFieldWithNameUnusable) {
				return object, err
			}

			if errors.Is(err, ErrFieldWithNameUnusable) {
				err = object.Init()
				if err != nil {
					return object, err
				}
			}

			parentValue = reflect.ValueOf(object.instance)
		case reflect.Map:
			object, err = NewOasObjectByName(parentValue.Interface(), itemName)
			if errors.Is(err, ErrNoValueWithKey) {
				err = object.Init()
				if err != nil {
					return object, err
				}
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

	return "", ErrNoFieldWithTag
}

func getYamlKeyFromField(field reflect.StructField) string {
	yamlTag := field.Tag.Get(YamlTag)

	return strings.Split(yamlTag, YamlTagSeparator)[0]
}
