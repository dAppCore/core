//go:build darwin && arm64 && mlx

package mlx

// #include "mlx/c/mlx.h"
import "C"

import "encoding/json"

// DType represents an MLX array data type.
type DType C.mlx_dtype

const (
	DTypeBool      DType = C.MLX_BOOL
	DTypeUint8     DType = C.MLX_UINT8
	DTypeUint16    DType = C.MLX_UINT16
	DTypeUint32    DType = C.MLX_UINT32
	DTypeUint64    DType = C.MLX_UINT64
	DTypeInt8      DType = C.MLX_INT8
	DTypeInt16     DType = C.MLX_INT16
	DTypeInt32     DType = C.MLX_INT32
	DTypeInt64     DType = C.MLX_INT64
	DTypeFloat16   DType = C.MLX_FLOAT16
	DTypeFloat32   DType = C.MLX_FLOAT32
	DTypeFloat64   DType = C.MLX_FLOAT64
	DTypeBFloat16  DType = C.MLX_BFLOAT16
	DTypeComplex64 DType = C.MLX_COMPLEX64
)

var dtypeNames = map[DType]string{
	DTypeBool:      "bool",
	DTypeUint8:     "uint8",
	DTypeUint16:    "uint16",
	DTypeUint32:    "uint32",
	DTypeUint64:    "uint64",
	DTypeInt8:      "int8",
	DTypeInt16:     "int16",
	DTypeInt32:     "int32",
	DTypeInt64:     "int64",
	DTypeFloat16:   "float16",
	DTypeFloat32:   "float32",
	DTypeFloat64:   "float64",
	DTypeBFloat16:  "bfloat16",
	DTypeComplex64: "complex64",
}

func (d DType) String() string {
	if s, ok := dtypeNames[d]; ok {
		return s
	}
	return "unknown"
}

var dtypeFromString = map[string]DType{
	"bool": DTypeBool, "BOOL": DTypeBool,
	"uint8": DTypeUint8, "U8": DTypeUint8,
	"uint16": DTypeUint16, "U16": DTypeUint16,
	"uint32": DTypeUint32, "U32": DTypeUint32,
	"uint64": DTypeUint64, "U64": DTypeUint64,
	"int8": DTypeInt8, "I8": DTypeInt8,
	"int16": DTypeInt16, "I16": DTypeInt16,
	"int32": DTypeInt32, "I32": DTypeInt32,
	"int64": DTypeInt64, "I64": DTypeInt64,
	"float16": DTypeFloat16, "F16": DTypeFloat16,
	"float32": DTypeFloat32, "F32": DTypeFloat32,
	"float64": DTypeFloat64, "F64": DTypeFloat64,
	"bfloat16": DTypeBFloat16, "BF16": DTypeBFloat16,
	"complex64": DTypeComplex64,
}

// UnmarshalJSON parses a DType from JSON strings like "F32", "BF16", etc.
func (d *DType) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if dt, ok := dtypeFromString[s]; ok {
		*d = dt
		return nil
	}
	*d = DTypeFloat32 // default
	return nil
}
