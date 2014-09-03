package structs

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errNotExported = errors.New("field is not exported")
	errNotSettable = errors.New("field is not settable")
)

// Field represents a single struct field that encapsulates high level
// functions around the field.
type Field struct {
	value reflect.Value
	field reflect.StructField
}

// Tag returns the value associated with key in the tag string. If there is no
// such key in the tag, Tag returns the empty string.
func (f *Field) Tag(key string) string {
	return f.field.Tag.Get(key)
}

// Value returns the underlying value of of the field. It panics if the field
// is not exported.
func (f *Field) Value() interface{} {
	return f.value.Interface()
}

// IsEmbedded returns true if the given field is an anonymous field (embedded)
func (f *Field) IsEmbedded() bool {
	return f.field.Anonymous
}

// IsExported returns true if the given field is exported.
func (f *Field) IsExported() bool {
	return f.field.PkgPath == ""
}

// IsZero returns true if the given field is not initalized (has a zero value).
// It panics if the field is not exported.
func (f *Field) IsZero() bool {
	zero := reflect.Zero(f.value.Type()).Interface()
	current := f.Value()

	return reflect.DeepEqual(current, zero)
}

// Name returns the name of the given field
func (f *Field) Name() string {
	return f.field.Name
}

// Kind returns the fields kind, such as "string", "map", "bool", etc ..
func (f *Field) Kind() reflect.Kind {
	return f.value.Kind()
}

func (f *Field) checkForSet(v reflect.Value, val interface{}) (reflect.Value, error) {
	var given reflect.Value
	if !f.IsExported() {
		return given, errNotExported
	}
	// do we get here? not sure...
	if !v.CanSet() {
		return given, errNotSettable
	}
	given = reflect.ValueOf(val)
	if given.Kind() == reflect.Ptr {
		given = given.Elem()
	}
	return given, nil
}

// Set sets the field to given value v. It retuns an error if the field is not
// settable (not addresable or not exported) or if the given value's type
// doesn't match the fields type.
func (f *Field) Set(val interface{}) error {
	// needed to make the field settable
	v := reflect.Indirect(f.value)
	given, err := f.checkForSet(v, val)
	if err != nil {
		return err
	}
	if v.Kind() != given.Kind() {
		return fmt.Errorf("wrong kind: %s want: %s", given.Kind(), v.Kind())
	}
	v.Set(given)
	return nil
}

// SetInt sets the field underlying value to val. It panics if the field is not a Kind is not Int, Int8, Int16, Int32, or Int64, or if CanSet() is false.
func (f *Field) SetInt(val int64) error {
	v := reflect.Indirect(f.value)
	_, err := f.checkForSet(v, val)
	if err != nil {
		return err
	}
	v.SetInt(val)
	return nil
}

// SetFloat sets the field underlying value to val. It panics if the field is not a Kind is not Float32 or Float64, or if CanSet() is false.
func (f *Field) SetFloat(val float64) error {
	v := reflect.Indirect(f.value)
	_, err := f.checkForSet(v, val)
	if err != nil {
		return err
	}
	v.SetFloat(val)
	return nil
}

//SetUint sets the field underlying value to val. It panics if the field is not a Kind is not Uint, Uintptr, Uint8, Uint16, Uint32, or Uint64, or if CanSet() is false.
func (f *Field) SetUint(val uint64) error {
	v := reflect.Indirect(f.value)
	_, err := f.checkForSet(v, val)
	if err != nil {
		return err
	}
	v.SetUint(val)
	return nil
}
// Fields returns a slice of Fields. This is particular handy to get the fields
// of a nested struct . A struct tag with the content of "-" ignores the
// checking of that particular field. Example:
//
//   // Field is ignored by this package.
//   Field *http.Request `structs:"-"`
//
// It panics if field is not exported or if field's kind is not struct
func (f *Field) Fields() []*Field {
	return getFields(f.value)
}

// Field returns the field from a nested struct. It panics if the nested struct
// is not exported or if the field was not found.
func (f *Field) Field(name string) *Field {
	field, ok := f.FieldOk(name)
	if !ok {
		panic("field not found")
	}

	return field
}

// Field returns the field from a nested struct. The boolean returns true if
// the field was found. It panics if the nested struct is not exported or if
// the field was not found.
func (f *Field) FieldOk(name string) (*Field, bool) {
	v := strctVal(f.value.Interface())
	t := v.Type()

	field, ok := t.FieldByName(name)
	if !ok {
		return nil, false
	}

	return &Field{
		field: field,
		value: v.FieldByName(name),
	}, true
}
