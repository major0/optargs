package goarg

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/major0/optargs"
)

// FlagBuilder converts goarg FieldMetadata into core parser flag maps.
type FlagBuilder struct {
	metadata  *StructMetadata
	config    Config
	setFields map[int]bool // tracks field indices explicitly set during parsing
}

// SetFields returns the set-fields tracker, populated during parsing
// via handler callbacks. The PostProcessor uses this to skip fields
// that were explicitly set.
func (fb *FlagBuilder) SetFields() map[int]bool {
	return fb.setFields
}

// Cached reflect.Type for time.Duration and TextUnmarshaler interface.
var (
	durationType         = reflect.TypeFor[time.Duration]()
	durationSliceType    = reflect.TypeFor[[]time.Duration]()
	textUnmarshalerIface = reflect.TypeFor[encoding.TextUnmarshaler]()
)

// typedValueForField creates an optargs.TypedValue backed by a pointer to
// the struct field's storage. Type dispatch happens once here at setup time;
// the returned TypedValue handles all subsequent Set() calls.
//
//nolint:gocyclo,cyclop,funlen // type switch over all supported Go types is inherently branchy
func typedValueForField(fieldValue reflect.Value, field *FieldMetadata) (optargs.TypedValue, error) {
	ft := field.Type

	// Pointer types: wrap in a ptrValue that allocates on first Set().
	if ft.Kind() == reflect.Ptr {
		return &ptrValue{fieldValue: fieldValue, elemType: ft.Elem(), field: field}, nil
	}

	// TextUnmarshaler takes priority over kind-based dispatch — user-defined
	// types (e.g., net.IP which is []byte) must be handled here before the
	// slice/scalar switch below.
	ptrType := reflect.PointerTo(ft)
	if ptrType.Implements(textUnmarshalerIface) {
		dest := fieldValue.Addr().Interface().(encoding.TextUnmarshaler) //nolint:errcheck // type verified by Implements check above
		var val encoding.TextMarshaler
		if m, ok := dest.(encoding.TextMarshaler); ok {
			val = m
		}
		return optargs.NewTextValue(val, dest), nil
	}

	// time.Duration must be checked before int64 (same Kind).
	if ft == durationType {
		p := fieldValue.Addr().Interface().(*time.Duration) //nolint:errcheck // type verified by ft == durationType check
		return optargs.NewDurationValue(*p, p), nil
	}

	// Scalar types.
	switch ft.Kind() {
	case reflect.String:
		p := fieldValue.Addr().Interface().(*string) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewStringValue(*p, p), nil
	case reflect.Bool:
		p := fieldValue.Addr().Interface().(*bool) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewBoolValue(*p, p), nil
	case reflect.Int:
		p := fieldValue.Addr().Interface().(*int) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewIntValue(*p, p), nil
	case reflect.Int8:
		p := fieldValue.Addr().Interface().(*int8) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewInt8Value(*p, p), nil
	case reflect.Int16:
		p := fieldValue.Addr().Interface().(*int16) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewInt16Value(*p, p), nil
	case reflect.Int32:
		p := fieldValue.Addr().Interface().(*int32) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewInt32Value(*p, p), nil
	case reflect.Int64:
		p := fieldValue.Addr().Interface().(*int64) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewInt64Value(*p, p), nil
	case reflect.Uint:
		p := fieldValue.Addr().Interface().(*uint) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewUintValue(*p, p), nil
	case reflect.Uint8:
		p := fieldValue.Addr().Interface().(*uint8) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewUint8Value(*p, p), nil
	case reflect.Uint16:
		p := fieldValue.Addr().Interface().(*uint16) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewUint16Value(*p, p), nil
	case reflect.Uint32:
		p := fieldValue.Addr().Interface().(*uint32) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewUint32Value(*p, p), nil
	case reflect.Uint64:
		p := fieldValue.Addr().Interface().(*uint64) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewUint64Value(*p, p), nil
	case reflect.Float32:
		p := fieldValue.Addr().Interface().(*float32) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewFloat32Value(*p, p), nil
	case reflect.Float64:
		p := fieldValue.Addr().Interface().(*float64) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewFloat64Value(*p, p), nil

	case reflect.Slice:
		return typedValueForSlice(fieldValue, ft)

	case reflect.Map:
		return typedValueForMap(fieldValue, ft)
	}

	return nil, fmt.Errorf("unsupported type %s for field %s", ft, field.Name)
}

// typedValueForSlice handles slice field types.
func typedValueForSlice(fieldValue reflect.Value, ft reflect.Type) (optargs.TypedValue, error) {
	// []time.Duration must be checked before []int64.
	if ft == durationSliceType {
		p := fieldValue.Addr().Interface().(*[]time.Duration) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewDurationSliceValue(*p, p), nil
	}

	switch ft.Elem().Kind() {
	case reflect.String:
		p := fieldValue.Addr().Interface().(*[]string) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewStringSliceValue(*p, p), nil
	case reflect.Bool:
		p := fieldValue.Addr().Interface().(*[]bool) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewBoolSliceValue(*p, p), nil
	case reflect.Int:
		p := fieldValue.Addr().Interface().(*[]int) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewIntSliceValue(*p, p), nil
	case reflect.Int32:
		p := fieldValue.Addr().Interface().(*[]int32) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewInt32SliceValue(*p, p), nil
	case reflect.Int64:
		p := fieldValue.Addr().Interface().(*[]int64) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewInt64SliceValue(*p, p), nil
	case reflect.Uint:
		p := fieldValue.Addr().Interface().(*[]uint) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewUintSliceValue(*p, p), nil
	case reflect.Float32:
		p := fieldValue.Addr().Interface().(*[]float32) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewFloat32SliceValue(*p, p), nil
	case reflect.Float64:
		p := fieldValue.Addr().Interface().(*[]float64) //nolint:errcheck // type verified by ft.Kind()+ft.Elem() switch
		return optargs.NewFloat64SliceValue(*p, p), nil
	}

	return nil, fmt.Errorf("unsupported slice element type: %s", ft.Elem())
}

// typedValueForMap handles map field types.
func typedValueForMap(fieldValue reflect.Value, ft reflect.Type) (optargs.TypedValue, error) {
	if ft.Key().Kind() != reflect.String {
		return nil, fmt.Errorf("unsupported map key type: %s", ft.Key())
	}

	switch ft.Elem().Kind() {
	case reflect.String:
		p := fieldValue.Addr().Interface().(*map[string]string) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewStringToStringValue(*p, p), nil
	case reflect.Int:
		p := fieldValue.Addr().Interface().(*map[string]int) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewStringToIntValue(*p, p), nil
	case reflect.Int64:
		p := fieldValue.Addr().Interface().(*map[string]int64) //nolint:errcheck // type verified by ft.Kind() switch
		return optargs.NewStringToInt64Value(*p, p), nil
	}

	return nil, fmt.Errorf("unsupported map value type: %s", ft.Elem())
}

// ptrValue wraps a pointer field. Allocates the pointed-to value on first
// Set() so that unset pointer fields remain nil.
type ptrValue struct {
	fieldValue reflect.Value
	elemType   reflect.Type
	field      *FieldMetadata
	inner      optargs.TypedValue // created lazily on first Set()
}

func (v *ptrValue) Set(s string) error {
	if v.inner == nil {
		// Allocate the pointer and create the inner TypedValue.
		v.fieldValue.Set(reflect.New(v.elemType))
		elemField := &FieldMetadata{
			Name:       v.field.Name,
			FieldIndex: v.field.FieldIndex,
			Type:       v.elemType,
		}
		var err error
		v.inner, err = typedValueForField(v.fieldValue.Elem(), elemField)
		if err != nil {
			return err
		}
	}
	return v.inner.Set(s)
}

func (v *ptrValue) String() string {
	if v.fieldValue.IsNil() {
		return ""
	}
	if v.inner != nil {
		return v.inner.String()
	}
	return ""
}

func (v *ptrValue) Type() string {
	return v.elemType.String()
}

func (v *ptrValue) IsBoolFlag() bool {
	return v.elemType.Kind() == reflect.Bool
}

// makeHandler returns a Handle callback that sets the struct field value when
// the option is parsed.
func (fb *FlagBuilder) makeHandler(field *FieldMetadata, destValue reflect.Value) (func(string, string) error, error) {
	fieldValue := fieldByMeta(destValue, field)
	if !fieldValue.CanSet() {
		return nil, fmt.Errorf("cannot set field %s", field.Name)
	}
	tv, err := typedValueForField(fieldValue, field)
	if err != nil {
		return nil, err
	}
	idx := field.FieldIndex
	return func(_, arg string) error {
		if arg == "" {
			if _, ok := tv.(optargs.BoolValuer); ok {
				if err := tv.Set("true"); err != nil {
					return err
				}
				fb.setFields[idx] = true
				return nil
			}
		}
		if err := tv.Set(arg); err != nil {
			return err
		}
		fb.setFields[idx] = true
		return nil
	}, nil
}

// makeBoolPrefixHandler returns a handler for a prefixed boolean option.
func (fb *FlagBuilder) makeBoolPrefixHandler(field *FieldMetadata, destValue reflect.Value, val bool) func(string, string) error {
	return func(_, _ string) error {
		fv := fieldByMeta(destValue, field)
		fv.SetBool(val)
		fb.setFields[field.FieldIndex] = true
		return nil
	}
}

// makeNegatableHandler returns a handler for --no-<name> on a non-boolean field.
func (fb *FlagBuilder) makeNegatableHandler(field *FieldMetadata, destValue reflect.Value) func(string, string) error {
	return func(_, _ string) error {
		fv := fieldByMeta(destValue, field)
		fv.Set(reflect.Zero(fv.Type()))
		fb.setFields[field.FieldIndex] = true
		return nil
	}
}

// Build produces the short and long option maps for optargs.NewParser.
func (fb *FlagBuilder) Build(destValue reflect.Value) (map[byte]*optargs.Flag, map[string]*optargs.Flag, error) {
	fb.setFields = make(map[int]bool)
	nOpts := len(fb.metadata.Options)
	shortOpts := make(map[byte]*optargs.Flag, nOpts)
	longOpts := make(map[string]*optargs.Flag, nOpts)

	for i := range fb.metadata.Options {
		field := &fb.metadata.Options[i]
		handler, err := fb.makeHandler(field, destValue)
		if err != nil {
			return nil, nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
		argName := strings.ToUpper(field.Name)
		defVal := formatDefault(field)

		hasShort := field.Short != ""
		hasLong := field.Long != ""

		switch {
		case hasShort && hasLong:
			flag := &optargs.Flag{
				Name:         field.Short,
				HasArg:       field.ArgType,
				Help:         field.Help,
				ArgName:      argName,
				DefaultValue: defVal,
				Handle:       handler,
			}
			shortOpts[field.Short[0]] = flag
			longOpts[field.Long] = flag
		case hasShort:
			shortOpts[field.Short[0]] = &optargs.Flag{
				Name:         field.Short,
				HasArg:       field.ArgType,
				Help:         field.Help,
				ArgName:      argName,
				DefaultValue: defVal,
				Handle:       handler,
			}
		case hasLong:
			longOpts[field.Long] = &optargs.Flag{
				Name:         field.Long,
				HasArg:       field.ArgType,
				Help:         field.Help,
				ArgName:      argName,
				DefaultValue: defVal,
				Handle:       handler,
			}
		}

		// Register prefix pair options for boolean fields (always NoArgument)
		if hasLong {
			for _, pp := range field.Prefixes {
				trueName := pp.True + "-" + field.Long
				falseName := pp.False + "-" + field.Long
				longOpts[trueName] = &optargs.Flag{
					Name:   trueName,
					HasArg: optargs.NoArgument,
					Handle: fb.makeBoolPrefixHandler(field, destValue, true),
				}
				longOpts[falseName] = &optargs.Flag{
					Name:   falseName,
					HasArg: optargs.NoArgument,
					Handle: fb.makeBoolPrefixHandler(field, destValue, false),
				}
			}

			// Register --no-<name> for negatable non-boolean fields
			if field.Negatable && field.Type.Kind() != reflect.Bool {
				negName := "no-" + field.Long
				longOpts[negName] = &optargs.Flag{
					Name:   negName,
					HasArg: optargs.NoArgument,
					Handle: fb.makeNegatableHandler(field, destValue),
				}
			}
		}
	}

	return shortOpts, longOpts, nil
}
