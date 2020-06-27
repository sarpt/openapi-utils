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
	object oasObject
	path   string
}

// Config specifies document handling
type Config struct {
	InlineLocalRefs bool
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
func (doc Document) ResolveReferences() error {
	rootObject := oasObject{
		parent: &doc,
		name:   RootItem,
	}

	refs, err := doc.parseOASObject(rootObject)
	if err != nil {
		return err
	}

	sort.Slice(refs, func(i, j int) bool {
		return sortReferences(refs[i], refs[j])
	})

	for _, ref := range refs {
		isLocal := isLocalReference(ref.path)
		if doc.Cfg.InlineLocalRefs && isLocal {
			continue
		}

		err := doc.assignReference(ref)
		if err != nil {
			return err
		}
	}

	return nil
}

// parseOASObject takes an object and returns list of all references that need to be resolved for object to be independent from references objects
func (doc Document) parseOASObject(object oasObject) ([]reference, error) {
	var allRefs []reference

	item, err := object.Get()
	if err != nil {
		return allRefs, err
	}

	switch item.Kind() {
	case reflect.Ptr:
		refPath, fields, err := parsePtrItem(item)
		if err != nil {
			return allRefs, err
		}

		if refPath != "" {
			ref := reference{
				object: object,
				path:   refPath,
			}
			allRefs = append(allRefs, ref)
		} else { // when refPath is in an oas object that is not a slice or map, other tahn $ref fields can be ignored per specification
			for _, field := range fields {
				obj := oasObject{
					parent: item.Interface(),
					name:   field,
				}
				objRefs, err := doc.parseOASObject(obj)
				if err != nil {
					return allRefs, err
				}

				allRefs = append(allRefs, objRefs...)
			}
		}
	case reflect.Map:
		refPaths, keys, err := parseMapItem(item)
		if err != nil {
			return allRefs, err
		}

		for _, refPath := range refPaths {
			ref := reference{
				object: object,
				path:   refPath,
			}
			allRefs = append(allRefs, ref)
		}

		for _, key := range keys {
			obj := oasObject{
				parent: item.Interface(),
				name:   key,
			}
			newRefs, err := doc.parseOASObject(obj)
			if err != nil {
				return allRefs, err
			}
			allRefs = append(allRefs, newRefs...)
		}
	case reflect.Slice:
		refPaths, indexes, err := parseSliceItem(item)
		if err != nil {
			return allRefs, err
		}

		for _, refPath := range refPaths {
			ref := reference{
				object: object,
				path:   refPath,
			}
			allRefs = append(allRefs, ref)
		}

		for _, idx := range indexes {
			obj := oasObject{
				parent: item.Interface(),
				idx:    idx,
			}
			newRefs, err := doc.parseOASObject(obj)
			if err != nil {
				return allRefs, err
			}
			allRefs = append(allRefs, newRefs...)
		}
	default:
		return allRefs, fmt.Errorf("could not parse field %s due to incorrect type", object.name)
	}

	return allRefs, nil
}

func parsePtrItem(item reflect.Value) (string, []string, error) {
	var fieldsToParse []string

	itemElem := item.Elem()
	itemType := itemElem.Type()
	for i := 0; i < itemElem.NumField(); i++ {
		childItem := itemElem.Field(i)

		if childItem.IsZero() {
			continue
		}

		switch childItem.Kind() {
		case reflect.String:
			if itemType.Field(i).Name == Ref {
				return childItem.String(), []string{}, nil // when ref found in an object then no need to parse other fields
			}
		case reflect.Struct, reflect.Ptr, reflect.Map, reflect.Slice, reflect.Array:
			fieldsToParse = append(fieldsToParse, itemType.Field(i).Name)
		}
	}

	return "", fieldsToParse, nil
}

func parseMapItem(item reflect.Value) ([]string, []string, error) {
	var keysToParse []string
	var refs []string

	mapIter := item.MapRange()
	for mapIter.Next() {
		childItem := mapIter.Value()

		if childItem.IsZero() {
			continue
		}

		switch childItem.Kind() {
		case reflect.String:
			if mapIter.Key().String() == Ref {
				refs = append(refs, childItem.String())
			}
		case reflect.Struct, reflect.Ptr, reflect.Map, reflect.Slice, reflect.Array:
			keysToParse = append(keysToParse, mapIter.Key().String())
		}
	}

	return refs, keysToParse, nil
}

func parseSliceItem(item reflect.Value) ([]string, []int, error) {
	var refs []string
	var indexesToParse []int

	for i := 0; i < item.Len(); i++ {
		childItem := item.Index(i)

		if childItem.IsZero() {
			continue
		}

		switch childItem.Kind() {
		case reflect.String:
			if childItem.String() == Ref {
				refs = append(refs, childItem.String())
			}
		case reflect.Struct, reflect.Ptr, reflect.Map, reflect.Slice, reflect.Array:
			indexesToParse = append(indexesToParse, i)
		}
	}

	return refs, indexesToParse, nil
}

// assignReference replaces ref either with inline object or a local reference
func (doc Document) assignReference(ref reference) error {
	referencedDocument, err := doc.getReferencedDocument(ref.path)
	if err != nil {
		return fmt.Errorf("could not get reference document: %w", err)
	}

	refItem, err := referencedDocument.getItemByPath(ref.path)
	if err != nil {
		return err
	}

	return replaceReference(ref.object, refItem)
}

func (doc Document) getItemByPath(refPath string) (interface{}, error) {
	itemNames := referencePathToItems(refPath)
	var parentValue reflect.Value = reflect.ValueOf(doc.Root)

	for _, itemName := range itemNames {
		switch parentValue.Kind() {
		case reflect.Ptr:
			parentElem := parentValue.Elem()
			childItem, err := getFieldByTag(itemName, parentElem)
			if err != nil {
				return nil, fmt.Errorf("could not find item %s in path %s: %w", itemName, refPath, err)
			}

			parentValue = childItem
		case reflect.Map:
			_, val, err := itemFromMapByName(parentValue, itemName)
			if err != nil {
				return nil, fmt.Errorf("could not resolve path %s due to error: %w", refPath, err)
			}

			parentValue = val
		default:
			return nil, fmt.Errorf("could not resolve path %s due to path including incorrect items", refPath)
		}
	}

	return parentValue.Interface(), nil
}

func getFieldByTag(tag string, structItem reflect.Value) (reflect.Value, error) {
	structItemType := structItem.Type()

	for i := 0; i < structItemType.NumField(); i++ {
		childField := structItemType.Field(i)
		yamlKey := getYamlKeyFromField(childField)
		if yamlKey == tag {
			return structItem.Field(i), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("the field with tag %s could not be found in type %s", tag, structItemType.Name())
}

// replaceReference unsets the reference and inlines refered object
func replaceReference(object oasObject, ref interface{}) error {
	parentVal := reflect.ValueOf(object.parent)
	refVal := reflect.ValueOf(ref)

	switch parentVal.Kind() {
	case reflect.Slice:
		parentVal.Index(object.idx).Set(refVal)
	case reflect.Ptr:
		field := parentVal.Elem().FieldByName(object.name)
		field.Set(refVal)
	case reflect.Map:
		key, _, err := itemFromMapByName(parentVal, object.name)
		if err != nil {
			return err
		}

		parentVal.SetMapIndex(key, refVal)
	}
	return nil
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
		parsedDocument, err := ParseDocument(doc.Cfg, documentFilePath)
		if err != nil {
			return nil, err
		}

		referencedDocument = &parsedDocument
		doc.ReferencedDocuments[documentFilePath] = referencedDocument
	}

	return referencedDocument, nil
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
