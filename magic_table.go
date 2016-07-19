package magicsql

import (
	"fmt"
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
	generator  func() interface{}
	name       string
	RType      reflect.Type
	sqlFields  []*boundField
	primaryKey *boundField
}

// newMagicTable creates a table structure with some pre-computed reflection
// data for the given generator.  The generator must be a zero-argument
// function which simply returns the type to be used with mapping sql to data.
// It must be safe to run the generator immediately in order to read its
// structure.  If conf is nil, the generator's object's tags are used to
// determine field mappings, otherwise the conf data is used.
func newMagicTable(tableName string, generator func() interface{}, conf ConfigTags) *magicTable {
	var t = &magicTable{generator: generator, name: tableName}
	t.reflect(conf)
	return t
}

// reflect traverses the wrapped structure to figure out which fields map to
// database table fields and how.  If conf is non-nil, that is used in place
// of struct tags.
func (t *magicTable) reflect(conf ConfigTags) {
	var obj = t.generator()
	t.RType = reflect.TypeOf(obj).Elem()
	var rVal = reflect.ValueOf(obj).Elem()

	for i := 0; i < t.RType.NumField(); i++ {
		var sf = t.RType.Field(i)

		if !rVal.Field(i).CanSet() {
			continue
		}

		var tag string
		if conf == nil {
			tag = sf.Tag.Get("sql")
		} else {
			tag = conf[sf.Name]
		}

		if tag == "-" {
			continue
		}

		var parts = strings.Split(tag, ",")
		var sqlf = parts[0]
		if sqlf == "" {
			sqlf = strings.ToLower(sf.Name)
		}

		var bf = &boundField{sqlf, sf}
		t.sqlFields = append(t.sqlFields, bf)

		if len(parts) > 1 && parts[1] == "primary" && t.primaryKey == nil {
			t.primaryKey = bf
		}
	}
}

// FieldNames returns all known table field names based on the tag parsing done
// in newMagicTable
func (t *magicTable) FieldNames() []string {
	var names []string
	for _, bf := range t.sqlFields {
		names = append(names, bf.Name)
	}
	return names
}

// SaveFieldNames returns all fields' names except primary key (if one exists)
// to ease insert and update statements where the primary key isn't part of
// what's saved
func (t *magicTable) SaveFieldNames() []string {
	var names []string
	for _, bf := range t.sqlFields {
		if bf == t.primaryKey {
			continue
		}
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

// InsertSQL returns the SQL string for inserting a record into this table.
// This makes the assumption that the primary key is not being set, so it isn't
// part of the fields list of values placeholder.
func (t *magicTable) InsertSQL() string {
	var qList []string
	for _, bf := range t.sqlFields {
		if bf == t.primaryKey {
			continue
		}
		qList = append(qList, "?")
	}

	var fields = strings.Join(t.SaveFieldNames(), ",")
	var placeholders = strings.Join(qList, ",")
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t.name, fields, placeholders)
}

// InsertArgs sets up and returns an array suitable for passing to an SQL Exec
// call for doing an insert
func (t *magicTable) InsertArgs(source interface{}) []interface{} {
	var save []interface{}
	var rVal = reflect.ValueOf(source).Elem()

	for _, bf := range t.sqlFields {
		if bf == t.primaryKey {
			continue
		}
		var vf = rVal.FieldByName(bf.Field.Name)
		save = append(save, vf.Addr().Interface())
	}

	return save
}

// UpdateSQL returns the SQL string for updating a record in this table.
// Returns an empty string if there's no primary key.
func (t *magicTable) UpdateSQL() string {
	if t.primaryKey == nil {
		return ""
	}

	var setList []string
	for _, bf := range t.sqlFields {
		if bf == t.primaryKey {
			continue
		}
		setList = append(setList, fmt.Sprintf("%s = ?", bf.Name))
	}
	var sets = strings.Join(setList, ",")

	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", t.name, sets, t.primaryKey.Name)
}

// UpdateArgs sets up and returns an array suitable for passing to an SQL Exec
// call for doing an update.  Returns nil if there's no primary key.
func (t *magicTable) UpdateArgs(source interface{}) []interface{} {
	if t.primaryKey == nil {
		return nil
	}

	var save []interface{}
	var rVal = reflect.ValueOf(source).Elem()

	for _, bf := range t.sqlFields {
		if bf == t.primaryKey {
			continue
		}
		var vf = rVal.FieldByName(bf.Field.Name)
		save = append(save, vf.Addr().Interface())
	}

	save = append(save, rVal.FieldByName(t.primaryKey.Field.Name).Addr().Interface())
	return save
}
