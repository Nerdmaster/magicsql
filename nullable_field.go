package magicsql

import (
	"database/sql"
	"time"
)

// NullableField implements the sql Scanner interface to make null values suck
// a little less.  When a null value is encountered, it's simply ignored, so
// the actual source can be set to a value that represents null or left at its
// default.  Data loss can happen if the source fields aren't of proper size,
// and not all types are supported.
type NullableField struct {
	Value interface{}
}

// Scan implements the Scanner interface.  Always returns a nil error.  Only
// works with primitive types or direct mappings of time.Time fields.  e.g., if
// the database holds a string, a field of time.Time won't be usable.
func (nf *NullableField) Scan(src interface{}) error {
	// Create a nullable field based on the type of the destination data
	switch nf.Value.(type) {
	case *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		nf.storeInt(src)
	case *float32, *float64:
		nf.storeFloat(src)
	case *bool:
		nf.storeBool(src)
	case *string:
		nf.storeString(src)
	case *time.Time:
		nf.storeTime(src)
	}

	return nil
}

func (nf *NullableField) storeInt(src interface{}) {
	var n sql.NullInt64
	n.Scan(src)
	if !n.Valid {
		return
	}
	var i = n.Int64

	switch d := nf.Value.(type) {
	case *int:
		*d = int(i)
	case *int8:
		*d = int8(i)
	case *int16:
		*d = int16(i)
	case *int32:
		*d = int32(i)
	case *int64:
		*d = int64(i)
	case *uint:
		*d = uint(i)
	case *uint8:
		*d = uint8(i)
	case *uint16:
		*d = uint16(i)
	case *uint32:
		*d = uint32(i)
	case *uint64:
		*d = uint64(i)
	}
}

func (nf *NullableField) storeFloat(src interface{}) {
	var n sql.NullFloat64
	n.Scan(src)
	if !n.Valid {
		return
	}
	var f = n.Float64

	switch d := nf.Value.(type) {
	case *float32:
		*d = float32(f)
	case *float64:
		*d = float64(f)
	}
}

func (nf *NullableField) storeBool(src interface{}) {
	var n sql.NullBool
	n.Scan(src)
	if !n.Valid {
		return
	}
	d := nf.Value.(*bool)
	*d = n.Bool
}

func (nf *NullableField) storeString(src interface{}) {
	var n sql.NullString
	n.Scan(src)
	if !n.Valid {
		return
	}
	d := nf.Value.(*string)
	*d = n.String
}

func (nf *NullableField) storeTime(src interface{}) {
	d := nf.Value.(*time.Time)
	switch st := src.(type) {
	case time.Time:
		*d = st
	}
}
