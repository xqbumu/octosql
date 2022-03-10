// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Copyright 2019 The OctoSQL Authors

Licensed under the MIT license, as in the LICENSE file
*/

// Package sqltypes implements interfaces and types that represent SQL values.
package sqltypes

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	querypb "github.com/cube2222/octosql/parser/dependency/query"
)

// MakeString makes a VarBinary Value.
func MakeString(val []byte) Value {
	return MakeTrusted(VarBinary, val)
}

// BuildValue builds a value from any go type. sqltype.Value is
// also allowed.
func BuildValue(goval interface{}) (v Value, err error) {
	// Look for the most common types first.
	switch goval := goval.(type) {
	case nil:
		// no op
	case []byte:
		v = MakeTrusted(VarBinary, goval)
	case int64:
		v = MakeTrusted(Int64, strconv.AppendInt(nil, int64(goval), 10))
	case uint64:
		v = MakeTrusted(Uint64, strconv.AppendUint(nil, uint64(goval), 10))
	case float64:
		v = MakeTrusted(Float64, strconv.AppendFloat(nil, goval, 'f', -1, 64))
	case int:
		v = MakeTrusted(Int64, strconv.AppendInt(nil, int64(goval), 10))
	case int8:
		v = MakeTrusted(Int8, strconv.AppendInt(nil, int64(goval), 10))
	case int16:
		v = MakeTrusted(Int16, strconv.AppendInt(nil, int64(goval), 10))
	case int32:
		v = MakeTrusted(Int32, strconv.AppendInt(nil, int64(goval), 10))
	case uint:
		v = MakeTrusted(Uint64, strconv.AppendUint(nil, uint64(goval), 10))
	case uint8:
		v = MakeTrusted(Uint8, strconv.AppendUint(nil, uint64(goval), 10))
	case uint16:
		v = MakeTrusted(Uint16, strconv.AppendUint(nil, uint64(goval), 10))
	case uint32:
		v = MakeTrusted(Uint32, strconv.AppendUint(nil, uint64(goval), 10))
	case float32:
		v = MakeTrusted(Float32, strconv.AppendFloat(nil, float64(goval), 'f', -1, 64))
	case string:
		v = MakeTrusted(VarBinary, []byte(goval))
	case time.Time:
		v = MakeTrusted(Datetime, []byte(goval.Format("2006-01-02 15:04:05")))
	case Value:
		v = goval
	case *querypb.BindVariable:
		return ValueFromBytes(goval.Type, goval.Value)
	default:
		return v, fmt.Errorf("unexpected type %T: %v", goval, goval)
	}
	return v, nil
}

// BuildConverted is like BuildValue except that it tries to
// convert a string or []byte to an integral if the target type
// is an integral. We don't perform other implicit conversions
// because they're unsafe.
func BuildConverted(typ querypb.Type, goval interface{}) (v Value, err error) {
	if IsIntegral(typ) {
		switch goval := goval.(type) {
		case []byte:
			return ValueFromBytes(typ, goval)
		case string:
			return ValueFromBytes(typ, []byte(goval))
		case Value:
			if goval.IsQuoted() {
				return ValueFromBytes(typ, goval.Raw())
			}
		case *querypb.BindVariable:
			if IsQuoted(goval.Type) {
				return ValueFromBytes(typ, goval.Value)
			}
		}
	}
	return BuildValue(goval)
}

// ValueFromBytes builds a Value using typ and val. It ensures that val
// matches the requested type. If type is an integral it's converted to
// a cannonical form. Otherwise, the original representation is preserved.
func ValueFromBytes(typ querypb.Type, val []byte) (v Value, err error) {
	switch {
	case IsSigned(typ):
		signed, err := strconv.ParseInt(string(val), 0, 64)
		if err != nil {
			return NULL, err
		}
		v = MakeTrusted(typ, strconv.AppendInt(nil, signed, 10))
	case IsUnsigned(typ):
		unsigned, err := strconv.ParseUint(string(val), 0, 64)
		if err != nil {
			return NULL, err
		}
		v = MakeTrusted(typ, strconv.AppendUint(nil, unsigned, 10))
	case typ == Tuple:
		return NULL, errors.New("tuple not allowed for ValueFromBytes")
	case IsFloat(typ) || typ == Decimal:
		_, err := strconv.ParseFloat(string(val), 64)
		if err != nil {
			return NULL, err
		}
		// After verification, we preserve the original representation.
		fallthrough
	default:
		v = MakeTrusted(typ, val)
	}
	return v, nil
}

// BuildIntegral builds an integral type from a string representaion.
// The type will be Int64 or Uint64. Int64 will be preferred where possible.
func BuildIntegral(val string) (n Value, err error) {
	signed, err := strconv.ParseInt(val, 0, 64)
	if err == nil {
		return MakeTrusted(Int64, strconv.AppendInt(nil, signed, 10)), nil
	}
	unsigned, err := strconv.ParseUint(val, 0, 64)
	if err != nil {
		return Value{}, err
	}
	return MakeTrusted(Uint64, strconv.AppendUint(nil, unsigned, 10)), nil
}

// ToNative converts Value to a native go type.
// This does not work for sqltypes.Tuple. The function
// panics if there are inconsistencies.
func (v Value) ToNative() interface{} {
	var out interface{}
	var err error
	switch {
	case v.typ == Null:
		// no-op
	case IsSigned(v.typ):
		out, err = v.ParseInt64()
	case IsUnsigned(v.typ):
		out, err = v.ParseUint64()
	case IsFloat(v.typ):
		out, err = v.ParseFloat64()
	case v.typ == Tuple:
		err = errors.New("unexpected tuple")
	default:
		out = v.val
	}
	if err != nil {
		panic(err)
	}
	return out
}

// ToProtoValue converts Value to a querypb.Value.
func (v Value) ToProtoValue() *querypb.Value {
	return &querypb.Value{
		Type:  v.typ,
		Value: v.val,
	}
}

// ParseInt64 will parse a Value into an int64. It does
// not check the type.
func (v Value) ParseInt64() (val int64, err error) {
	return strconv.ParseInt(v.String(), 10, 64)
}

// ParseUint64 will parse a Value into a uint64. It does
// not check the type.
func (v Value) ParseUint64() (val uint64, err error) {
	return strconv.ParseUint(v.String(), 10, 64)
}

// ParseFloat64 will parse a Value into an float64. It does
// not check the type.
func (v Value) ParseFloat64() (val float64, err error) {
	return strconv.ParseFloat(v.String(), 64)
}

func writebyte(c byte, b BinWriter) {
	if err := b.WriteByte(c); err != nil {
		panic(err)
	}
}

func writebytes(val []byte, b BinWriter) {
	n, err := b.Write(val)
	if err != nil {
		panic(err)
	}
	if n != len(val) {
		panic(errors.New("short write"))
	}
}
