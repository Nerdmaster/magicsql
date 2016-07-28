package magicsql

import (
	"reflect"
)

// OperationTable ties together a memorized MagicTable definition with an
// in-progress SQL operation
type OperationTable struct {
	op *Operation
	t  *MagicTable
}

// Reconfigure sends explicit ConfigTags data to the underlying MagicTable in
// order to override the reflected structure's tags
func (ot *OperationTable) Reconfigure(conf ConfigTags) {
	ot.t.Configure(conf)
}

// Select simply instantiates a Select instance with the OperationTable set up
// for it to use for gathering fields and running the query
func (ot *OperationTable) Select() Select {
	return Select{ot: ot}
}

// Save determines if an INSERT or UPDATE is necessary (primary key of 0 means
// INSERT here), generates the SQL and arguments, and runs the Exec call on the
// database.  Behavior may be unpredictable if a MagicTable was manually
// registered with a structure of a different type than obj.
func (ot *OperationTable) Save(obj interface{}) *Result {
	// Check for object's primary key field being zero
	var rVal = reflect.ValueOf(obj).Elem()
	var pkValField = rVal.FieldByName(ot.t.primaryKey.Field.Name)

	if pkValField.Interface() == reflect.Zero(pkValField.Type()).Interface() {
		var res = ot.op.Exec(ot.t.InsertSQL(), ot.t.InsertArgs(obj)...)
		pkValField.SetInt(res.LastInsertId())
		return res
	}

	return ot.op.Exec(ot.t.UpdateSQL(), ot.t.UpdateArgs(obj)...)
}

// Insert forces an insert, ignoring any primary key tagging.  Note that
// directly inserting via this method will *not* auto-set the primary key to
// the last insert id.
//
// TODO: make the DB querying backend more flexible so it doesn't care about
// tagging for generating SQL and arg lists, instead relying on field and
// column name mappings
func (ot *OperationTable) Insert(obj interface{}) *Result {
	var pkField = ot.t.primaryKey
	ot.t.primaryKey = nil
	var res = ot.op.Exec(ot.t.InsertSQL(), ot.t.InsertArgs(obj)...)
	ot.t.primaryKey = pkField

	return res
}
