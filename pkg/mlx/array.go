//go:build darwin && arm64 && mlx

package mlx

/*
#include <stdlib.h>
#include "mlx/c/mlx.h"
*/
import "C"

import (
	"encoding/binary"
	"reflect"
	"runtime"
	"strings"
	"unsafe"
)

// Array wraps an mlx_array handle.
// Memory management relies on Go GC finalizers to call mlx_array_free,
// which decrements MLX-C's internal reference count. MLX-C handles all
// cross-array references internally — the Go wrapper does not track them.
type Array struct {
	ctx  C.mlx_array
	name string // debug label
}

// New creates a named Array and registers a GC finalizer.
// The inputs parameter is accepted for API compatibility but not stored —
// MLX-C tracks inter-array references via its own refcounting.
func New(name string, inputs ...*Array) *Array {
	t := &Array{name: name}
	runtime.SetFinalizer(t, finalizeArray)
	return t
}

// finalizeArray is called by Go GC to release the underlying C array handle.
func finalizeArray(t *Array) {
	if t != nil && t.ctx.ctx != nil {
		C.mlx_array_free(t.ctx)
		t.ctx.ctx = nil
	}
}

type scalarTypes interface {
	~bool | ~int | ~float32 | ~float64 | ~complex64
}

// FromValue creates a scalar Array from a Go value.
func FromValue[T scalarTypes](t T) *Array {
	Init()
	tt := New("")
	switch v := any(t).(type) {
	case bool:
		tt.ctx = C.mlx_array_new_bool(C.bool(v))
	case int:
		tt.ctx = C.mlx_array_new_int(C.int(v))
	case float32:
		tt.ctx = C.mlx_array_new_float32(C.float(v))
	case float64:
		tt.ctx = C.mlx_array_new_float64(C.double(v))
	case complex64:
		tt.ctx = C.mlx_array_new_complex(C.float(real(v)), C.float(imag(v)))
	default:
		panic("mlx: unsupported scalar type")
	}
	return tt
}

type arrayTypes interface {
	~bool | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64 |
		~complex64
}

// FromValues creates an Array from a Go slice with the given shape.
func FromValues[S ~[]E, E arrayTypes](s S, shape ...int) *Array {
	Init()
	if len(shape) == 0 {
		panic("mlx: shape required for non-scalar tensors")
	}

	cShape := make([]C.int, len(shape))
	for i := range shape {
		cShape[i] = C.int(shape[i])
	}

	var dtype DType
	switch reflect.TypeOf(s).Elem().Kind() {
	case reflect.Bool:
		dtype = DTypeBool
	case reflect.Uint8:
		dtype = DTypeUint8
	case reflect.Uint16:
		dtype = DTypeUint16
	case reflect.Uint32:
		dtype = DTypeUint32
	case reflect.Uint64:
		dtype = DTypeUint64
	case reflect.Int8:
		dtype = DTypeInt8
	case reflect.Int16:
		dtype = DTypeInt16
	case reflect.Int32:
		dtype = DTypeInt32
	case reflect.Int64:
		dtype = DTypeInt64
	case reflect.Float32:
		dtype = DTypeFloat32
	case reflect.Float64:
		dtype = DTypeFloat64
	case reflect.Complex64:
		dtype = DTypeComplex64
	default:
		panic("mlx: unsupported element type")
	}

	bts := make([]byte, binary.Size(s))
	if _, err := binary.Encode(bts, binary.LittleEndian, s); err != nil {
		panic(err)
	}

	tt := New("")
	tt.ctx = C.mlx_array_new_data(unsafe.Pointer(&bts[0]), unsafe.SliceData(cShape), C.int(len(cShape)), C.mlx_dtype(dtype))
	return tt
}

// Zeros creates a zero-filled Array with the given shape and dtype.
func Zeros(shape []int32, dtype DType) *Array {
	Init()
	cShape := make([]C.int, len(shape))
	for i, s := range shape {
		cShape[i] = C.int(s)
	}
	tt := New("ZEROS")
	C.mlx_zeros(&tt.ctx, unsafe.SliceData(cShape), C.size_t(len(cShape)), C.mlx_dtype(dtype), DefaultStream().ctx)
	return tt
}

// Set replaces this array's C handle with another's.
func (t *Array) Set(other *Array) {
	C.mlx_array_set(&t.ctx, other.ctx)
}

// Clone creates a new Go wrapper sharing the same C handle (increments C refcount).
func (t *Array) Clone() *Array {
	tt := New(t.name)
	C.mlx_array_set(&tt.ctx, t.ctx)
	return tt
}

// Valid reports whether this Array has a non-nil mlx handle.
func (t *Array) Valid() bool {
	return t.ctx.ctx != nil
}

// String returns a human-readable representation of the array.
func (t *Array) String() string {
	str := C.mlx_string_new()
	defer C.mlx_string_free(str)
	C.mlx_array_tostring(&str, t.ctx)
	return strings.TrimSpace(C.GoString(C.mlx_string_data(str)))
}

// Shape returns the dimensions as int32 slice.
func (t *Array) Shape() []int32 {
	dims := make([]int32, t.NumDims())
	for i := range dims {
		dims[i] = int32(t.Dim(i))
	}
	return dims
}

// Size returns the total number of elements.
func (t Array) Size() int { return int(C.mlx_array_size(t.ctx)) }

// NumBytes returns the total byte size.
func (t Array) NumBytes() int { return int(C.mlx_array_nbytes(t.ctx)) }

// NumDims returns the number of dimensions.
func (t Array) NumDims() int { return int(C.mlx_array_ndim(t.ctx)) }

// Dim returns the size of dimension i.
func (t Array) Dim(i int) int { return int(C.mlx_array_dim(t.ctx, C.int(i))) }

// Dims returns all dimensions as int slice.
func (t Array) Dims() []int {
	dims := make([]int, t.NumDims())
	for i := range dims {
		dims[i] = t.Dim(i)
	}
	return dims
}

// Dtype returns the array's data type.
func (t Array) Dtype() DType { return DType(C.mlx_array_dtype(t.ctx)) }

// Int extracts a scalar int64 value.
func (t Array) Int() int {
	var item C.int64_t
	C.mlx_array_item_int64(&item, t.ctx)
	return int(item)
}

// Float extracts a scalar float64 value.
func (t Array) Float() float64 {
	var item C.double
	C.mlx_array_item_float64(&item, t.ctx)
	return float64(item)
}

// Ints extracts all elements as int slice (from int32 data).
func (t Array) Ints() []int {
	ints := make([]int, t.Size())
	for i, f := range unsafe.Slice(C.mlx_array_data_int32(t.ctx), len(ints)) {
		ints[i] = int(f)
	}
	return ints
}

// DataInt32 extracts all elements as int32 slice.
func (t Array) DataInt32() []int32 {
	data := make([]int32, t.Size())
	for i, f := range unsafe.Slice(C.mlx_array_data_int32(t.ctx), len(data)) {
		data[i] = int32(f)
	}
	return data
}

// Floats extracts all elements as float32 slice.
func (t Array) Floats() []float32 {
	floats := make([]float32, t.Size())
	for i, f := range unsafe.Slice(C.mlx_array_data_float32(t.ctx), len(floats)) {
		floats[i] = float32(f)
	}
	return floats
}

// Free explicitly releases C array handles. Does not cascade — MLX-C's
// internal refcounting handles dependent arrays automatically.
func Free(s ...*Array) int {
	var n int
	for _, t := range s {
		if t != nil && t.Valid() {
			n += t.NumBytes()
			C.mlx_array_free(t.ctx)
			t.ctx.ctx = nil
			runtime.SetFinalizer(t, nil) // cancel finalizer
		}
	}
	return n
}
