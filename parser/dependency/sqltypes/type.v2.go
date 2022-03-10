/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqltypes

import (
	querypb "github.com/cube2222/octosql/parser/dependency/query"
)

// If you add to this map, make sure you add a test case
// in tabletserver/endtoend.
var mysqlToType = map[int64]querypb.Type{
	0:   Decimal,
	1:   Int8,
	2:   Int16,
	3:   Int32,
	4:   Float32,
	5:   Float64,
	6:   Null,
	7:   Timestamp,
	8:   Int64,
	9:   Int24,
	10:  Date,
	11:  Time,
	12:  Datetime,
	13:  Year,
	15:  VarChar,
	16:  Bit,
	17:  Timestamp,
	18:  Datetime,
	19:  Time,
	245: TypeJSON,
	246: Decimal,
	247: Enum,
	248: Set,
	249: Text,
	250: Text,
	251: Text,
	252: Text,
	253: VarChar,
	254: Char,
	255: Geometry,
}

// AreTypesEquivalent returns whether two types are equivalent.
func AreTypesEquivalent(mysqlTypeFromBinlog, mysqlTypeFromSchema querypb.Type) bool {
	return (mysqlTypeFromBinlog == mysqlTypeFromSchema) ||
		(mysqlTypeFromBinlog == VarChar && mysqlTypeFromSchema == VarBinary) ||
		// Binlog only has base type. But doesn't have per-column-flags to differentiate
		// various logical types. For Binary, Enum, Set types, binlog only returns Char
		// as data type.
		(mysqlTypeFromBinlog == Char && mysqlTypeFromSchema == Binary) ||
		(mysqlTypeFromBinlog == Char && mysqlTypeFromSchema == Enum) ||
		(mysqlTypeFromBinlog == Char && mysqlTypeFromSchema == Set) ||
		(mysqlTypeFromBinlog == Text && mysqlTypeFromSchema == Blob) ||
		(mysqlTypeFromBinlog == Int8 && mysqlTypeFromSchema == Uint8) ||
		(mysqlTypeFromBinlog == Int16 && mysqlTypeFromSchema == Uint16) ||
		(mysqlTypeFromBinlog == Int24 && mysqlTypeFromSchema == Uint24) ||
		(mysqlTypeFromBinlog == Int32 && mysqlTypeFromSchema == Uint32) ||
		(mysqlTypeFromBinlog == Int64 && mysqlTypeFromSchema == Uint64)
}
