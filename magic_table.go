package magicsql

import (
  "fmt"
  "reflect"
  "strings"
)

type boundField struct {
  Name     string
  Field    reflect.StructField
}

// MagicTable represents a named database table for reading data from a single
// table into a tagged structure
type MagicTable struct {
	db          *DB
	generator   func() interface{}
  name        string
  RType       reflect.Type
  sqlFields   []*boundField
	err         errorable
}

// Err returns the first error encountered on any operation the parent DB
// object oversees
func (t *MagicTable) Err() error {
	return t.err.Err()
}

// reflect traverses the wrapped structure to figure out which fields map to
// database table fields and how
func (t *MagicTable) reflect() {
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

func (t *MagicTable) fieldNames() []string {
  var names []string
  for _, bf := range t.sqlFields {
    names = append(names, bf.Name)
  }
  return names
}

func (t *MagicTable) scanStruct(dest interface{}) []interface{} {
  var fields = make([]interface{}, len(t.sqlFields))
  var rVal = reflect.ValueOf(dest).Elem()
  for i, bf := range t.sqlFields {
		var vf = rVal.FieldByName(bf.Field.Name)
    fields[i] = &NullableField{Value: vf.Addr().Interface()}
  }

  return fields
}

func (t *MagicTable) buildQuerySQL(where string) string {
  var format = "SELECT %s FROM %s%s"
  var selectFieldList = strings.Join(t.fieldNames(), ",")
  var whereClause string
  if where != "" {
    whereClause = " WHERE " + where
  }

  return fmt.Sprintf(format, selectFieldList, t.name, whereClause)
}

// FindAll runs a select query with the given where clause and bound args,
// returning an array of whatever type was given the DB.RegisterTable as the
// generator function output
func (t *MagicTable) FindAll(where string, args ...interface{}) []interface{} {
	if t.err.Err() != nil {
		return nil
	}

  var sql = t.buildQuerySQL(where)
	var stmt = t.db.Prepare(sql)
  var rows = stmt.Query(args...)
	var data []interface{}
	for rows.Next() {
		var obj = t.generator()
		rows.Scan(t.scanStruct(obj)...)
		data = append(data, obj)
	}
  return data
}
