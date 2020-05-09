package openapi

import (
	"fmt"
	"io/ioutil"
	"reflect"

	yaml "gopkg.in/yaml.v2"
)

// Document represents single OpenAPI source file and it's content
// A Document can be dependent on other Documents by using OpenAPI references
type Document struct {
	FilePath            string
	Root                *OpenAPI
	ReferencedDocuments map[string]*Document
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

// ResolveReferences takes a document and tries to find and resolve all references
// After execution all elements that had not empty Ref properties have their contents replaced with referenced content
func (doc Document) ResolveReferences() error {
	return doc.parseItem(&doc, "Root")
}

func (doc Document) parseItem(parent interface{}, fieldName string) error {
	var refPath string
	item, err := getFieldFromParent(parent, fieldName)
	if err != nil {
		return err
	}

	if !(item.Kind() == reflect.Ptr || item.Kind() == reflect.Map) {
		return fmt.Errorf("%s is neither a pointer nor a map", item.Type().Name())
	}

	childrenItemsToParse := []string{}

	switch item.Kind() {
	case reflect.Ptr:
		itemElem := item.Elem()
		itemType := itemElem.Type()
		for i := 0; i < itemElem.NumField(); i++ {
			childItem := itemElem.Field(i)

			if childItem.IsZero() {
				continue
			}

			switch childItem.Kind() {
			case reflect.String:
				if itemType.Field(i).Name == "Ref" {
					refPath = childItem.String()
				}
			case reflect.Array, reflect.Slice:
			case reflect.Struct, reflect.Ptr:
				childrenItemsToParse = append(childrenItemsToParse, itemType.Field(i).Name)
			case reflect.Map:
				childrenItemsToParse = append(childrenItemsToParse, itemType.Field(i).Name)
			}
		}
	case reflect.Map:
		mapIter := item.MapRange()
		for mapIter.Next() {
			childItem := mapIter.Value()

			if childItem.IsZero() {
				continue
			}

			switch childItem.Kind() {
			case reflect.String:
				if mapIter.Key().String() == "Ref" {
					refPath = childItem.String()
				}
			case reflect.Array, reflect.Slice:
			case reflect.Struct, reflect.Ptr:
				childrenItemsToParse = append(childrenItemsToParse, mapIter.Key().String())
			case reflect.Map:
				childrenItemsToParse = append(childrenItemsToParse, mapIter.Key().String())
			}
		}
	}

	if refPath != "" {
		return doc.assignReference(refPath, parent, fieldName)
	}

	for _, childItem := range childrenItemsToParse {
		err := doc.parseItem(item.Interface(), childItem)
		if err != nil {
			return err
		}
	}
	return nil
}

func getFieldFromParent(parent interface{}, fieldName string) (reflect.Value, error) {
	parentVal := reflect.ValueOf(parent)
	switch parentVal.Kind() {
	case reflect.Ptr:
		return parentVal.Elem().FieldByName(fieldName), nil
	case reflect.Map:
		mapIter := parentVal.MapRange()
		for mapIter.Next() {
			if mapIter.Key().String() == fieldName {
				return mapIter.Value(), nil
			}
		}
	default:
		return reflect.Value{}, fmt.Errorf("provided parent is neither pointer to struct nor a map")
	}

	return reflect.Value{}, fmt.Errorf("provided name %s could not be found on parent", fieldName)
}

func (doc Document) assignReference(refPath string, parent interface{}, fieldName string) error {
	referencedDocument, err := doc.getReferencedDocument(refPath)
	if err != nil {
		return fmt.Errorf("Could not get reference document: %w", err)
	}

	refItem, err := referencedDocument.getItemByPath(refPath)
	if err != nil {
		return err
	}

	return replaceReference(parent, fieldName, refItem)
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
			mapIter := parentValue.MapRange()
			for mapIter.Next() {
				if mapIter.Key().String() == itemName {
					parentValue = mapIter.Value()
				}
			}
		}
	}

	return parentValue.Interface(), nil
}

func getFieldByTag(tag string, structItem reflect.Value) (reflect.Value, error) {
	structItemType := structItem.Type()
	for i := 0; i < structItemType.NumField(); i++ {
		childItem := structItemType.Field(i)
		yamlTag := childItem.Tag.Get("yaml")
		if yamlTag == tag {
			return structItem.Field(i), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("the field with tag %s could not be found in type %s", tag, structItemType.Name())
}

func replaceReference(parent interface{}, fieldName string, ref interface{}) error {
	parentVal := reflect.ValueOf(parent)
	refVal := reflect.ValueOf(ref)

	switch parentVal.Kind() {
	case reflect.Ptr:
	case reflect.Map:
		mapIter := parentVal.MapRange()
		for mapIter.Next() {
			if mapIter.Key().String() == fieldName {
				parentVal.SetMapIndex(mapIter.Key(), refVal)
			}
		}
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
