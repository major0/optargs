package optargs

import (
	"fmt"
	"reflect"
	"strings"
)

// mapValue is the generic implementation for map typed values.
// First Set() replaces the default map; subsequent calls merge.
type mapValue struct {
	p        any // pointer to destination map
	valType  reflect.Type
	typeName string
	firstSet bool // tracks whether first Set() has been called
}

func (v *mapValue) Set(s string) error {
	dest := reflect.ValueOf(v.p).Elem()

	// First Set() replaces the default map.
	if !v.firstSet {
		dest.Set(reflect.MakeMap(dest.Type()))
		v.firstSet = true
	}

	pairs := strings.Split(s, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		before, after, ok := strings.Cut(pair, "=")
		if !ok {
			return fmt.Errorf("invalid map entry %q: expected key=value", pair)
		}
		key := before
		val := after

		converted, err := Convert(val, v.valType)
		if err != nil {
			return err
		}
		dest.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(converted))
	}
	return nil
}

func (v *mapValue) String() string {
	dest := reflect.ValueOf(v.p).Elem()
	if dest.IsNil() || dest.Len() == 0 {
		return "map[]"
	}
	parts := make([]string, 0, dest.Len())
	iter := dest.MapRange()
	for iter.Next() {
		parts = append(parts, fmt.Sprintf("%v=%v", iter.Key().Interface(), iter.Value().Interface()))
	}
	return "map[" + strings.Join(parts, ",") + "]"
}

func (v *mapValue) Type() string { return v.typeName }

// Reset clears the map to its zero value (empty map).
func (v *mapValue) Reset() {
	dest := reflect.ValueOf(v.p).Elem()
	dest.Set(reflect.MakeMap(dest.Type()))
	v.firstSet = false
}

// --- Map constructors ---

func NewStringToStringValue(val map[string]string, p *map[string]string) TypedValue {
	if p == nil {
		p = new(map[string]string)
	}
	*p = val
	return &mapValue{p: p, valType: stringType, typeName: "stringToString"}
}

func NewStringToIntValue(val map[string]int, p *map[string]int) TypedValue {
	if p == nil {
		p = new(map[string]int)
	}
	*p = val
	return &mapValue{p: p, valType: intType, typeName: "stringToInt"}
}

func NewStringToInt64Value(val map[string]int64, p *map[string]int64) TypedValue {
	if p == nil {
		p = new(map[string]int64)
	}
	*p = val
	return &mapValue{p: p, valType: int64Type, typeName: "stringToInt64"}
}
