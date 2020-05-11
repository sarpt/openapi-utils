package openapi

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"

	yaml "gopkg.in/yaml.v2"
)

const (
	// RootItem is an OpenAPI Root
	RootItem = "Root"
	// Ref is a reference field
	Ref = "Ref"
	// YamlTag is a tag key which is used to parse YAML into internal representation
	YamlTag = "yaml"
)

// Document represents single OpenAPI source file and it's content
// A Document can be dependent on other Documents by using OpenAPI references
type Document struct {
	FilePath            string
	Root                *OpenAPI
	ReferencedDocuments map[string]*Document
}

// OASObject respresent the object of the OpenAPI schema
type OASObject struct {
	parent interface{}
	name   string
	idx    int
}

// Reference contains information about OpenAPI object that contains reference and path of reference
type Reference struct {
	object OASObject
	path   string
}

// NewDocument constructs new Document instance
func NewDocument(filePath string) Document {
	return Document{
		FilePath:            filePath,
		Root:                &OpenAPI{},
		ReferencedDocuments: make(map[string]*Document),
	}
}

// ParseDocument takes path to the file that should be parsed and have it's references resolved
func ParseDocument(path string) (Document, error) {
	referencedDocument := NewDocument(path)

	err := referencedDocument.ParseFile()
	if err != nil {
		return Document{}, err
	}

	err = referencedDocument.ResolveReferences()
	return referencedDocument, err
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

// WriteFile writes content of a document to a YAML file pointed by path
func (doc Document) WriteFile(path string) error {
	yaml, err := doc.YAML()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, yaml, os.FileMode(0777))
}

// YAML converts contents of a document to YAML
func (doc Document) YAML() ([]byte, error) {
	return yaml.Marshal(doc.Root)
}

// ResolveReferences takes a document and tries to find and resolve all references
// After execution all elements that had not empty Ref properties have their contents replaced with referenced content
func (doc Document) ResolveReferences() error {
	object := OASObject{
		parent: &doc,
		name:   RootItem,
	}

	refs, err := doc.parseOASObject(object)
	if err != nil {
		return err
	}

	sort.Slice(refs, func(i, j int) bool {
		return sortReferences(refs[i], refs[j])
	})

	for _, ref := range refs {
		err := doc.assignReference(ref)
		if err != nil {
			return err
		}
	}

	return nil
}

func (doc Document) parseOASObject(object OASObject) ([]Reference, error) {
	var allRefs []Reference

	item, err := getFieldFromParent(object)
	if err != nil {
		return allRefs, err
	}

	switch item.Kind() {
	case reflect.Ptr:
		ref, fields, err := parsePtrItem(item)
		if err != nil {
			return allRefs, err
		}

		if ref != "" {
			reference := Reference{
				object: object,
				path:   ref,
			}
			allRefs = append(allRefs, reference)
		}

		for _, field := range fields {
			obj := OASObject{
				parent: item.Interface(),
				name:   field,
			}
			objRefs, err := doc.parseOASObject(obj)
			if err != nil {
				return allRefs, err
			}

			allRefs = append(allRefs, objRefs...)
		}
	case reflect.Map:
		refs, keys, err := parseMapItem(item)
		if err != nil {
			return allRefs, err
		}

		for _, ref := range refs {
			reference := Reference{
				object: object,
				path:   ref,
			}
			allRefs = append(allRefs, reference)
		}

		for _, key := range keys {
			obj := OASObject{
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
		refs, indexes, err := parseSliceItem(item)
		if err != nil {
			return allRefs, err
		}

		for _, ref := range refs {
			reference := Reference{
				object: object,
				path:   ref,
			}
			allRefs = append(allRefs, reference)
		}

		for _, idx := range indexes {
			obj := OASObject{
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
				return childItem.String(), fieldsToParse, nil
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

func getFieldFromParent(object OASObject) (reflect.Value, error) {
	parentVal := reflect.ValueOf(object.parent)
	switch parentVal.Kind() {
	case reflect.Ptr:
		return parentVal.Elem().FieldByName(object.name), nil
	case reflect.Map:
		_, val, err := itemFromMapByName(parentVal, object.name)
		return val, err
	case reflect.Slice:
		return parentVal.Index(object.idx), nil
	default:
		return reflect.Value{}, fmt.Errorf("provided parent for %s field is not a correct type", object.name)
	}
}

func (doc Document) assignReference(ref Reference) error {
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
				return nil, fmt.Errorf("could not find item %s in path %s", itemName, refPath)
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
		childItem := structItemType.Field(i)
		yamlTag := childItem.Tag.Get(YamlTag)
		if yamlTag == tag {
			return structItem.Field(i), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("the field with tag %s could not be found in type %s", tag, structItemType.Name())
}

func replaceReference(object OASObject, ref interface{}) error {
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

	pathToDocument := getPathToRemoteDocument(refPath)

	if document, ok := doc.ReferencedDocuments[pathToDocument]; ok {
		referencedDocument = document
	} else {
		parsedDocument, err := ParseDocument(pathToDocument)
		if err != nil {
			return nil, err
		}

		referencedDocument = &parsedDocument
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
