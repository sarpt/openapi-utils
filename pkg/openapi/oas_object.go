package openapi

import (
	"errors"
	"reflect"
)

var (
	// ErrIncorrectParent occurs when instance of oasObject has parent that cannot be used to find the object.
	ErrIncorrectParent = errors.New("parent has incorrect type")
	// ErrIncorrectObjectType occurs when underlying instance of object has incorrect type and cannot be parsed.
	ErrIncorrectObjectType = errors.New("object has incorrect type")
	// ErrNoValueWithKey occurs when for specified map and key, the value could not be retrieved due to key missing in the map
	ErrNoValueWithKey = errors.New("no map value matches specified key")
	// ErrFieldWithNameUnusable occurs when for specified field name, the value is unusable either by no field being present or field being zero pointer
	ErrFieldWithNameUnusable = errors.New("field with specified name is unusable")
	// ErrFieldWithNameNotInType occurs when for specified field name, the parent type has no child field matching i
	ErrFieldWithNameNotInType = errors.New("field with specified is not specified by type")
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
	obj := &OasObject{
		parent: parent,
		name:   name,
	}

	err := obj.parse()
	if err != nil {
		return *obj, err
	}

	return *obj, nil
}

// NewOasObjectByIdx takes parent and an index under which object can be found
func NewOasObjectByIdx(parent interface{}, idx int) (OasObject, error) {
	obj := &OasObject{
		parent: parent,
		idx:    idx,
	}

	err := obj.parse()
	if err != nil {
		return *obj, err
	}

	return *obj, nil
}

// parse retrieves and fills the OpenAPI instance.
func (o *OasObject) parse() error {
	parentVal := reflect.ValueOf(o.parent)
	switch parentVal.Kind() {
	case reflect.Ptr:
		field := parentVal.Elem().FieldByName(o.name)
		if field.IsZero() {
			return ErrFieldWithNameUnusable
		}

		o.instance = field.Interface()
		return nil
	case reflect.Map:
		_, val, err := itemFromMapByName(parentVal, o.name)
		if err != nil {
			return err
		}

		o.instance = val.Interface()
		return nil
	case reflect.Slice:
		o.instance = parentVal.Index(o.idx).Interface()
		return nil
	default:
		return ErrIncorrectParent
	}
}

// Init force for the underlying OasObject to be replaced with zero value
// For OasObject which is of map type it means making a map that can be used
// For OasObject which is an entry inside a map it means creating entry in a map
// Slice to be implemented (does $ref permit indexes inside references? And if it does, are there any referencable elements in a slice?)
func (o *OasObject) Init() error {
	parentType := reflect.TypeOf(o.parent)
	objectType := parentType.Elem()

	switch parentType.Kind() {
	case reflect.Ptr:
		structField, ok := objectType.FieldByName(o.name)
		if !ok {
			return ErrFieldWithNameNotInType
		}

		// Slices to be implemented
		if structField.Type.Kind() != reflect.Map {
			return nil
		}

		newMap := reflect.MakeMap(structField.Type).Interface()
		o.Set(newMap)
	case reflect.Map:
		childVal := reflect.New(objectType).Elem().Interface()
		o.Set(childVal)
	default:
		return ErrIncorrectParent
	}

	return nil
}

// Set replaces the underlying object with a provided value.
func (o *OasObject) Set(val interface{}) error {
	parentVal := reflect.ValueOf(o.parent)
	refVal := reflect.ValueOf(val)

	switch parentVal.Kind() {
	case reflect.Slice:
		parentVal.Index(o.idx).Set(refVal)
	case reflect.Ptr:
		field := parentVal.Elem().FieldByName(o.name)
		field.Set(refVal)
	case reflect.Map:
		childKey := reflect.New(reflect.TypeOf(o.parent).Key()).Elem()
		childKey.Set(reflect.ValueOf(o.name))
		parentVal.SetMapIndex(childKey, refVal)
	default:
		return ErrIncorrectParent
	}

	o.instance = val
	return nil
}

// Unset removes value, removing instance of OpenAPi object from a document.
// Note that depending on the options of the parent document, the copy of this object can be present someplace else in the document after unsetting.
// After calling Unset, calling Set has unspecified behavior and should be considered invalid as it can either crash at runtime or replace other objects under the same parent.
func (o OasObject) Unset() error {
	return o.Set(nil)
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

func itemFromMapByName(mapVal reflect.Value, key string) (reflect.Value, reflect.Value, error) {
	mapIter := mapVal.MapRange()
	for mapIter.Next() {
		if mapIter.Key().String() == key {
			return mapIter.Key(), mapIter.Value(), nil
		}
	}

	return reflect.Value{}, reflect.Value{}, ErrNoValueWithKey
}
