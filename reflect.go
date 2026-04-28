// SPDX-License-Identifier: EUPL-1.2

// Reflection primitives — re-exports of Go's standard reflect package as
// Core types, so consumers never need to write `import "reflect"`.
// Reflection is a low-level tool; reach for it only when generic
// constraints and type switches genuinely cannot express the operation.
//
//	if core.TypeOf(v).Kind() == core.KindStruct { ... }
package core

import "reflect"

// Type is the runtime type descriptor returned by TypeOf.
//
//	t := core.TypeOf(opts)
//	name := t.Name()
type Type = reflect.Type

// Value is the runtime value handle returned by ValueOf.
//
//	v := core.ValueOf(42)
//	n := v.Int()
type Value = reflect.Value

// Kind classifies a Type into one of the basic Go kinds (Bool, Int,
// String, Struct, Map, Slice, etc.).
//
//	if core.TypeOf(opts).Kind() == core.KindStruct { ... }
type Kind = reflect.Kind

// Common Kind constants — re-exported from reflect so consumers can
// compare without importing reflect.
//
//	switch core.TypeOf(v).Kind() {
//	case core.KindString:
//	case core.KindInt, core.KindInt64:
//	}
const (
	KindInvalid       = reflect.Invalid
	KindBool          = reflect.Bool
	KindInt           = reflect.Int
	KindInt8          = reflect.Int8
	KindInt16         = reflect.Int16
	KindInt32         = reflect.Int32
	KindInt64         = reflect.Int64
	KindUint          = reflect.Uint
	KindUint8         = reflect.Uint8
	KindUint16        = reflect.Uint16
	KindUint32        = reflect.Uint32
	KindUint64        = reflect.Uint64
	KindFloat32       = reflect.Float32
	KindFloat64       = reflect.Float64
	KindComplex64     = reflect.Complex64
	KindComplex128    = reflect.Complex128
	KindArray         = reflect.Array
	KindChan          = reflect.Chan
	KindFunc          = reflect.Func
	KindInterface     = reflect.Interface
	KindMap           = reflect.Map
	KindPointer       = reflect.Pointer
	KindSlice         = reflect.Slice
	KindString        = reflect.String
	KindStruct        = reflect.Struct
	KindUnsafePointer = reflect.UnsafePointer
)

// TypeOf returns the runtime type of v. Returns nil if v is a nil
// interface value.
//
//	t := core.TypeOf(opts)
//	if t.Kind() == core.KindStruct { ... }
func TypeOf(v any) Type {
	return reflect.TypeOf(v)
}

// ValueOf returns a Value initialised to the concrete value stored in v.
// Returns the zero Value if v is a nil interface value.
//
//	val := core.ValueOf(42)
//	n := val.Int()  // 42
func ValueOf(v any) Value {
	return reflect.ValueOf(v)
}

// DeepEqual reports whether x and y are deeply equal — recursively
// comparing slices, maps, structs, etc. Differs from == in that it
// follows pointers and compares unexported fields.
//
//	core.DeepEqual([]int{1, 2, 3}, []int{1, 2, 3})  // true
//	core.DeepEqual(map[string]int{"a": 1}, map[string]int{"a": 1})  // true
func DeepEqual(x, y any) bool {
	return reflect.DeepEqual(x, y)
}
