package mustache

import (
	"reflect"
	"strings"
)

// The lookup function searches for a property that matches name within the
// context chain.
func lookup(name string, context ...interface{}) (interface{}, bool) {
	// If the dot notation was used we split the word in two and perform two
	// consecutive lookups. If the first one fails we return no value and a
	// negative truth.
	if name != "." && strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		if value, ok := lookup(parts[0], context...); ok {
			return lookup(parts[1], value)
		}
		return nil, false
	}
	// Iterate over the context chain and try to match the name to a value.
	for _, c := range context {
		reflectValue := reflect.ValueOf(c)
		if name == "." {
			return c, truth(reflectValue)
		}
		switch reflectValue.Kind() {
		case reflect.Map:
			mapValue := reflectValue.MapIndex(reflect.ValueOf(name))
			if mapValue.IsValid() {
				return mapValue.Interface(), truth(mapValue)
			}
		case reflect.Struct:
			fieldValue := reflectValue.FieldByName(name)
			if fieldValue.IsValid() {
				return fieldValue.Interface(), truth(fieldValue)
			} else {
				method := reflectValue.MethodByName(name)
				if method.IsValid() && method.Type().NumIn() == 1 {
					out := method.Call(nil)[0]
					return out.Interface(), truth(out)
				}
			}
		}
	}
	return nil, false
}

func truth(r reflect.Value) bool {
	switch r.Kind() {
	case reflect.Array, reflect.Slice:
		return r.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return r.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return r.Uint() > 0
	case reflect.String:
		return r.String() != ""
	case reflect.Bool:
		return r.Bool()
	default:
		return r.Interface() != nil
	}
}
