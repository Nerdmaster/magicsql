package magicsql

import (
	"fmt"
	"reflect"
	"strings"
)

type boundField struct {
	Name     string
	Field    reflect.StructField
	NoInsert bool
	NoUpdate bool
}

// MagicTable represents a named database table for reading data from a single
// table into a tagged structure
type MagicTable struct {
	Object     interface{}
	Name       string
	RType      reflect.Type
	sqlFields  []*boundField
	primaryKey *boundField
}

// Table registers a table name and an object's type for use in database
// operations.  The returned MagicTable is pre-configured using the object's
// structure tags.
func Table(name string, obj interface{}) *MagicTable {
	var mt = &MagicTable{Name: name, Object: obj}
	mt.Configure(nil)
	return mt
}

// Configure traverses the wrapped structure to figure out which fields map to
// database table columns and how.  If conf is non-nil, that is used in place
// of struct tags.  This is run if a table is created with Table(), but can be
// useful for reconfiguring a table with explicit ConfigTags.
func (t *MagicTable) Configure(conf ConfigTags) {
	t.sqlFields = nil
	t.RType = reflect.TypeOf(t.Object).Elem()
	var rVal = reflect.ValueOf(t.Object).Elem()

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
			sqlf = toUnderscore(sf.Name)
		}

		var bf = &boundField{sqlf, sf, false, false}
		t.sqlFields = append(t.sqlFields, bf)

		if len(parts) > 1 {
			for _, part := range parts[1:] {
				switch part {
				case "primary":
					if t.primaryKey == nil {
						bf.NoInsert = true
						bf.NoUpdate = true
						t.primaryKey = bf
					}
				case "readonly":
					bf.NoInsert = true
					bf.NoUpdate = true
				case "noinsert":
					bf.NoInsert = true
				case "noupdate":
					bf.NoUpdate = true
				}
			}
		}
	}
}

// FieldNames returns all known table field names based on the tag parsing done
// in newMagicTable
func (t *MagicTable) FieldNames() []string {
	var names []string
	for _, bf := range t.sqlFields {
		names = append(names, bf.Name)
	}
	return names
}

// ScanStruct sets up a structure suitable for calling Scan to populate dest
func (t *MagicTable) ScanStruct(dest interface{}) []interface{} {
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
func (t *MagicTable) InsertSQL() string {
	var fList []string
	var qList []string
	for _, bf := range t.sqlFields {
		if bf.NoInsert {
			continue
		}
		fList = append(fList, bf.Name)
		qList = append(qList, "?")
	}

	var fields = strings.Join(fList, ",")
	var placeholders = strings.Join(qList, ",")
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", t.Name, fields, placeholders)
}

// InsertArgs sets up and returns an array suitable for passing to an SQL Exec
// call for doing an insert
func (t *MagicTable) InsertArgs(source interface{}) []interface{} {
	var save []interface{}
	var rVal = reflect.ValueOf(source).Elem()

	for _, bf := range t.sqlFields {
		// TODO: Make this check for empty val so we can force primary keys in explicit Insert calls
		if bf.NoInsert {
			continue
		}
		var vf = rVal.FieldByName(bf.Field.Name)
		save = append(save, vf.Addr().Interface())
	}

	return save
}

// UpdateSQL returns the SQL string for updating a record in this table.
// Returns an empty string if there's no primary key.
func (t *MagicTable) UpdateSQL() string {
	if t.primaryKey == nil {
		return ""
	}

	var setList []string
	for _, bf := range t.sqlFields {
		if bf.NoUpdate {
			continue
		}
		setList = append(setList, fmt.Sprintf("%s = ?", bf.Name))
	}
	var sets = strings.Join(setList, ",")

	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", t.Name, sets, t.primaryKey.Name)
}

// UpdateArgs sets up and returns an array suitable for passing to an SQL Exec
// call for doing an update.  Returns nil if there's no primary key.
func (t *MagicTable) UpdateArgs(source interface{}) []interface{} {
	if t.primaryKey == nil {
		return nil
	}

	var save []interface{}
	var rVal = reflect.ValueOf(source).Elem()

	for _, bf := range t.sqlFields {
		if bf.NoUpdate {
			continue
		}
		var vf = rVal.FieldByName(bf.Field.Name)
		save = append(save, vf.Addr().Interface())
	}

	save = append(save, rVal.FieldByName(t.primaryKey.Field.Name).Addr().Interface())
	return save
}
