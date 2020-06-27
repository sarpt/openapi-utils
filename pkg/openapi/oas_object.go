package openapi

import (
	"fmt"
	"reflect"
)

var (
	// ErrIncorrectParent occurs when instance of oasObject has parent that cannot be used to find the object.
	ErrIncorrectParent = fmt.Errorf("parent has incorrect type")
	// ErrIncorrectObjectType occurs when underlying instance of object has incorrect type and cannot be parsed.
	ErrIncorrectObjectType = fmt.Errorf("object has incorrect type")
)

// OasObject respresent the object of the OpenAPI schema.
// For a parent that is a pointer or a map, the name is used to hold the field name for which object is accessible in the parent.
// For a parent that is a list, the index should be used to obtain the object from it.
type OasObject struct {
	parent   interface{}
	instance interface{}
	name     string
	idx      int
}

// NewOasObjectByName takes parent and a name under which object can be found
func NewOasObjectByName(parent interface{}, name string) (OasObject, error) {
	obj := OasObject{
		parent: parent,
		name:   name,
	}

	instance, err := obj.parse()
	if err != nil {
		return obj, err
	}

	obj.instance = instance

	return obj, nil
}

// NewOasObjectByIdx takes parent and an index under which object can be found
func NewOasObjectByIdx(parent interface{}, idx int) (OasObject, error) {
	obj := OasObject{
		parent: parent,
		idx:    idx,
	}

	instance, err := obj.parse()
	if err != nil {
		return obj, err
	}

	obj.instance = instance

	return obj, nil
}

// parse returns the interface of the OpenAPI object that is pointed by the instance.
func (o OasObject) parse() (interface{}, error) {
	parentVal := reflect.ValueOf(o.parent)
	switch parentVal.Kind() {
	case reflect.Ptr:
		return parentVal.Elem().FieldByName(o.name).Interface(), nil
	case reflect.Map:
		_, val, err := itemFromMapByName(parentVal, o.name)
		return val.Interface(), err
	case reflect.Slice:
		return parentVal.Index(o.idx).Interface(), nil
	default:
		return nil, ErrIncorrectParent
	}
}

// Set unsets the underlying object and sets a provided value.
func (o OasObject) Set(val interface{}) error {
	parentVal := reflect.ValueOf(o.parent)
	refVal := reflect.ValueOf(val)

	switch parentVal.Kind() {
	case reflect.Slice:
		parentVal.Index(o.idx).Set(refVal)
	case reflect.Ptr:
		field := parentVal.Elem().FieldByName(o.name)
		field.Set(refVal)
	case reflect.Map:
		key, _, err := itemFromMapByName(parentVal, o.name)
		if err != nil {
			return err
		}

		parentVal.SetMapIndex(key, refVal)
	default:
		return ErrIncorrectParent
	}

	return nil
}

// references returns list of all references that need to be resolved for object to be independent from its references.
// That list includes children references along with object's own references since parsing is done recursively until refs in all possible descendants are found.
func (o OasObject) references() ([]reference, error) {
	var allRefs []reference

	value := reflect.ValueOf(o.instance)

	switch value.Kind() {
	case reflect.Ptr:
		refPath, fields, err := parsePtrValue(value)
		if err != nil {
			return allRefs, err
		}

		if refPath != "" {
			ref := reference{
				object: o,
				path:   refPath,
			}
			allRefs = append(allRefs, ref)
		} else { // when refPath is in an oas object that is not a slice or map, other tahn $ref fields can be ignored per specification
			for _, field := range fields {
				obj, err := NewOasObjectByName(value.Interface(), field)
				if err != nil {
					return allRefs, err
				}

				objRefs, err := obj.references()
				if err != nil {
					return allRefs, err
				}

				allRefs = append(allRefs, objRefs...)
			}
		}
	case reflect.Map:
		refPaths, keys, err := parseMapValue(value)
		if err != nil {
			return allRefs, err
		}

		for _, refPath := range refPaths {
			ref := reference{
				object: o,
				path:   refPath,
			}
			allRefs = append(allRefs, ref)
		}

		for _, key := range keys {
			obj, err := NewOasObjectByName(value.Interface(), key)
			if err != nil {
				return allRefs, err
			}

			newRefs, err := obj.references()
			if err != nil {
				return allRefs, err
			}
			allRefs = append(allRefs, newRefs...)
		}
	case reflect.Slice:
		refPaths, indexes, err := parseSliceValue(value)
		if err != nil {
			return allRefs, err
		}

		for _, refPath := range refPaths {
			ref := reference{
				object: o,
				path:   refPath,
			}
			allRefs = append(allRefs, ref)
		}

		for _, idx := range indexes {
			obj, err := NewOasObjectByIdx(value.Interface(), idx)
			if err != nil {
				return allRefs, err
			}

			newRefs, err := obj.references()
			if err != nil {
				return allRefs, err
			}
			allRefs = append(allRefs, newRefs...)
		}
	default:
		return allRefs, ErrIncorrectObjectType
	}

	return allRefs, nil
}

func parsePtrValue(value reflect.Value) (string, []string, error) {
	var fieldsToParse []string

	itemElem := value.Elem()
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

func parseMapValue(value reflect.Value) ([]string, []string, error) {
	var keysToParse []string
	var refs []string

	mapIter := value.MapRange()
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

func parseSliceValue(value reflect.Value) ([]string, []int, error) {
	var refs []string
	var indexesToParse []int

	for i := 0; i < value.Len(); i++ {
		childItem := value.Index(i)

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
