package magicsql

import (
	"reflect"
	"strings"
)

type boundField struct {
	Name  string
	Field reflect.StructField
}

// magicTable represents a named database table for reading data from a single
// table into a tagged structure
type magicTable struct {
	generator func() interface{}
	name      string
	RType     reflect.Type
	sqlFields []*boundField
}

// NewMagicTable creates a table structure with some pre-computed reflection
// data for the given generator.  The generator must be a zero-argument
// function which simply returns the type to be used with mapping sql to data.
// It must be safe to run the generator immediately in order to read its
// structure.
func NewMagicTable(tableName string, generator func() interface{}) *magicTable {
	var t = &magicTable{generator: generator, name: tableName}
	t.reflect()
	return t
}

// reflect traverses the wrapped structure to figure out which fields map to
// database table fields and how
func (t *magicTable) reflect() {
	var obj = t.generator()
	t.RType = reflect.TypeOf(obj).Elem()
	var rVal = reflect.ValueOf(obj).Elem()

	for i := 0; i < t.RType.NumField(); i++ {
		var sf = t.RType.Field(i)

		if !rVal.Field(i).CanSet() {
			continue
		}

		var sqlf = sf.Tag.Get("sql")
		if sqlf == "-" {
			continue
		}
		if sqlf == "" {
			sqlf = strings.ToLower(sf.Name)
		}

		t.sqlFields = append(t.sqlFields, &boundField{sqlf, sf})
	}
}

// FieldNames returns all known table field names based on the tag parsing done
// in NewMagicTable
func (t *magicTable) FieldNames() []string {
	var names []string
	for _, bf := range t.sqlFields {
		names = append(names, bf.Name)
	}
	return names
}

// ScanStruct sets up a structure suitable for calling Scan to populate dest
func (t *magicTable) ScanStruct(dest interface{}) []interface{} {
	var fields = make([]interface{}, len(t.sqlFields))
	var rVal = reflect.ValueOf(dest).Elem()
	for i, bf := range t.sqlFields {
		var vf = rVal.FieldByName(bf.Field.Name)
		fields[i] = &NullableField{Value: vf.Addr().Interface()}
	}

	return fields
}
