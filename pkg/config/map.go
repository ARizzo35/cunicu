package config

import (
	"fmt"
	"reflect"
	"strings"
)

var configPkgPath = reflect.TypeOf(Settings{}).PkgPath()

func Map(v any, tagName string) map[string]any {
	rv := reflect.ValueOf(v)

	return _map(rv, tagName).(map[string]any)
}

func _map(v reflect.Value, tagName string) any {
	t := v.Type()

	// Stringable types will be stringed
	if s, ok := v.Interface().(fmt.Stringer); ok {
		return s.String()
	}

	if v.CanAddr() {
		if s, ok := v.Addr().Interface().(fmt.Stringer); ok {
			return s.String()
		}
	}

	// Types outside the config package will be taken as an interface
	if t.PkgPath() != configPkgPath && t.PkgPath() != "" {
		return v.Interface()
	}

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		t = v.Type()
	}

	switch v.Kind() {
	case reflect.Struct:
		d := map[string]any{}

		for i := 0; i < v.NumField(); i++ {
			fv := v.Field(i)
			sf := t.Field(i)

			if fv.IsValid() && !fv.IsZero() {
				if tag, ok := sf.Tag.Lookup(tagName); ok {
					name := strings.Split(tag, ",")[0]
					n := _map(fv, tagName)
					if name != "" {
						d[name] = n
					} else if m, ok := n.(map[string]any); ok {
						for k, v := range m {
							d[k] = v
						}
					}
				}
			}
		}

		return d

	case reflect.Map:
		if t.Key().Kind() == reflect.String {
			d := map[string]any{}

			for _, e := range v.MapKeys() {
				d[e.String()] = _map(v.MapIndex(e), tagName)
			}

			return d
		} else {
			return v.Interface()
		}

	case reflect.Slice:
		d := []any{}

		for i := 0; i < v.Len(); i++ {
			d = append(d, _map(v.Index(i), tagName))
		}

		return d

	default:
		return v.Interface()
	}
}
