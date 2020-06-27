package openapi

import (
	"fmt"
	"reflect"
)

// oasObject respresent the object of the OpenAPI schema.
// For a parent that is a pointer or a map, the name is used to hold the field name for which object is accessible in the parent.
// For a parent that is a list, the index should be used to obtain the object from it.
type oasObject struct {
	parent interface{}
	name   string
	idx    int
}

// Get returns the reflect.Value of the OpenAPI object that is pointed by the instance.
func (o oasObject) Get() (reflect.Value, error) {
	parentVal := reflect.ValueOf(o.parent)
	switch parentVal.Kind() {
	case reflect.Ptr:
		return parentVal.Elem().FieldByName(o.name), nil
	case reflect.Map:
		_, val, err := itemFromMapByName(parentVal, o.name)
		return val, err
	case reflect.Slice:
		return parentVal.Index(o.idx), nil
	default:
		return reflect.Value{}, fmt.Errorf("provided parent for %s field is not a correct type", o.name)
	}
}
