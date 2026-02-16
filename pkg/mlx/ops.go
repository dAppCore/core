//go:build darwin && arm64 && mlx

package mlx

/*
#include <stdlib.h>
#include "mlx/c/mlx.h"
*/
import "C"

import "unsafe"

// --- Element-wise arithmetic ---

// Add returns element-wise a + b.
func Add(a, b *Array) *Array {
	out := New("ADD", a, b)
	C.mlx_add(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// AddScalar returns a + scalar (broadcast).
func AddScalar(a *Array, s float32) *Array {
	scalar := FromValue(s)
	return Add(a, scalar)
}

// Mul returns element-wise a * b.
func Mul(a, b *Array) *Array {
	out := New("MUL", a, b)
	C.mlx_multiply(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// MulScalar returns a * scalar (broadcast).
func MulScalar(a *Array, s float32) *Array {
	scalar := FromValue(s)
	return Mul(a, scalar)
}

// Divide returns element-wise a / b.
func Divide(a, b *Array) *Array {
	out := New("DIV", a, b)
	C.mlx_divide(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// Subtract returns element-wise a - b.
func Subtract(a, b *Array) *Array {
	out := New("SUB", a, b)
	C.mlx_subtract(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// Negative returns element-wise -a.
func Negative(a *Array) *Array {
	out := New("NEG", a)
	C.mlx_negative(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// --- Math functions ---

// Exp returns element-wise exp(a).
func Exp(a *Array) *Array {
	out := New("EXP", a)
	C.mlx_exp(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// Tanh returns element-wise tanh(a).
func Tanh(a *Array) *Array {
	out := New("TANH", a)
	C.mlx_tanh(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// Sqrt returns element-wise sqrt(a).
func Sqrt(a *Array) *Array {
	out := New("SQRT", a)
	C.mlx_sqrt(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// Rsqrt returns element-wise 1/sqrt(a).
func Rsqrt(a *Array) *Array {
	out := New("RSQRT", a)
	C.mlx_rsqrt(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// Reciprocal returns element-wise 1/a.
func Reciprocal(a *Array) *Array {
	out := New("RECIPROCAL", a)
	C.mlx_reciprocal(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// Square returns element-wise a^2.
func Square(a *Array) *Array {
	out := New("SQUARE", a)
	C.mlx_square(&out.ctx, a.ctx, DefaultStream().ctx)
	return out
}

// Power returns element-wise a^b.
func Power(a, b *Array) *Array {
	out := New("POWER", a, b)
	C.mlx_power(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// Maximum returns element-wise max(a, b).
func Maximum(a, b *Array) *Array {
	out := New("MAX", a, b)
	C.mlx_maximum(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// Minimum returns element-wise min(a, b).
func Minimum(a, b *Array) *Array {
	out := New("MIN", a, b)
	C.mlx_minimum(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// --- Matrix operations ---

// Matmul returns the matrix product of a and b.
func Matmul(a, b *Array) *Array {
	out := New("MATMUL", a, b)
	C.mlx_matmul(&out.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// QuantizedMatmul performs quantized matrix multiplication.
func QuantizedMatmul(x, w, scales, biases *Array, transpose bool, groupSize, bits int) *Array {
	out := New("QMATMUL", x, w, scales, biases)
	gs := C.mlx_optional_int{value: C.int(groupSize), has_value: C._Bool(true)}
	b := C.mlx_optional_int{value: C.int(bits), has_value: C._Bool(true)}
	mode := C.CString("affine")
	defer C.free(unsafe.Pointer(mode))
	C.mlx_quantized_matmul(
		&out.ctx, x.ctx, w.ctx, scales.ctx, biases.ctx,
		C._Bool(transpose), gs, b, mode,
		DefaultStream().ctx,
	)
	return out
}

// --- Reductions ---

// Softmax returns softmax along the last axis.
func Softmax(a *Array) *Array {
	out := New("SOFTMAX", a)
	axis := []C.int{C.int(-1)}
	C.mlx_softmax_axes(&out.ctx, a.ctx, &axis[0], C.size_t(1), C._Bool(false), DefaultStream().ctx)
	return out
}

// Argmax returns the index of the maximum value along an axis.
func Argmax(a *Array, axis int, keepDims bool) *Array {
	out := New("ARGMAX", a)
	C.mlx_argmax_axis(&out.ctx, a.ctx, C.int(axis), C._Bool(keepDims), DefaultStream().ctx)
	return out
}

// TopK returns the top k values along the last axis.
func TopK(a *Array, k int) *Array {
	out := New("TOPK", a)
	C.mlx_topk_axis(&out.ctx, a.ctx, C.int(k), C.int(-1), DefaultStream().ctx)
	return out
}

// Sum reduces by summation along the given axis.
func Sum(a *Array, axis int, keepDims bool) *Array {
	out := New("SUM", a)
	axes := []C.int{C.int(axis)}
	C.mlx_sum_axes(&out.ctx, a.ctx, &axes[0], C.size_t(1), C._Bool(keepDims), DefaultStream().ctx)
	return out
}

// Mean reduces by averaging along the given axis.
func Mean(a *Array, axis int, keepDims bool) *Array {
	out := New("MEAN", a)
	axes := []C.int{C.int(axis)}
	C.mlx_mean_axes(&out.ctx, a.ctx, &axes[0], C.size_t(1), C._Bool(keepDims), DefaultStream().ctx)
	return out
}

// --- Shape operations ---

// Reshape changes the shape of an array.
func Reshape(a *Array, shape ...int32) *Array {
	out := New("RESHAPE", a)
	cShape := make([]C.int, len(shape))
	for i, s := range shape {
		cShape[i] = C.int(s)
	}
	C.mlx_reshape(&out.ctx, a.ctx, &cShape[0], C.size_t(len(cShape)), DefaultStream().ctx)
	return out
}

// Transpose permutes dimensions. If no axes given, reverses all dims.
func Transpose(a *Array, axes ...int) *Array {
	out := New("TRANSPOSE", a)
	if len(axes) == 0 {
		C.mlx_transpose(&out.ctx, a.ctx, DefaultStream().ctx)
	} else {
		cAxes := make([]C.int, len(axes))
		for i, ax := range axes {
			cAxes[i] = C.int(ax)
		}
		C.mlx_transpose_axes(&out.ctx, a.ctx, &cAxes[0], C.size_t(len(cAxes)), DefaultStream().ctx)
	}
	return out
}

// ExpandDims inserts a new axis at the given position.
func ExpandDims(a *Array, axis int) *Array {
	out := New("EXPAND_DIMS", a)
	C.mlx_expand_dims(&out.ctx, a.ctx, C.int(axis), DefaultStream().ctx)
	return out
}

// Squeeze removes dimensions of size 1.
func Squeeze(a *Array, axes ...int) *Array {
	out := New("SQUEEZE", a)
	cAxes := make([]C.int, len(axes))
	for i, ax := range axes {
		cAxes[i] = C.int(ax)
	}
	C.mlx_squeeze_axes(&out.ctx, a.ctx, &cAxes[0], C.size_t(len(cAxes)), DefaultStream().ctx)
	return out
}

// Concatenate joins arrays along the given axis.
func Concatenate(arrays []*Array, axis int) *Array {
	vector := C.mlx_vector_array_new()
	defer C.mlx_vector_array_free(vector)

	inputs := make([]*Array, len(arrays))
	for i, a := range arrays {
		C.mlx_vector_array_append_value(vector, a.ctx)
		inputs[i] = a
	}

	out := New("CONCAT", inputs...)
	C.mlx_concatenate_axis(&out.ctx, vector, C.int(axis), DefaultStream().ctx)
	return out
}

// BroadcastTo broadcasts an array to the given shape.
func BroadcastTo(a *Array, shape []int32) *Array {
	out := New("BROADCAST", a)
	cShape := make([]C.int, len(shape))
	for i, s := range shape {
		cShape[i] = C.int(s)
	}
	C.mlx_broadcast_to(&out.ctx, a.ctx, &cShape[0], C.size_t(len(cShape)), DefaultStream().ctx)
	return out
}

// AsType casts an array to a different dtype.
func AsType(a *Array, dtype DType) *Array {
	out := New("ASTYPE", a)
	C.mlx_astype(&out.ctx, a.ctx, C.mlx_dtype(dtype), DefaultStream().ctx)
	return out
}

// AsStrided creates a view with custom strides.
func AsStrided(a *Array, shape []int32, strides []int64, offset int64) *Array {
	out := New("AS_STRIDED", a)
	cShape := make([]C.int, len(shape))
	for i, s := range shape {
		cShape[i] = C.int(s)
	}
	cStrides := make([]C.int64_t, len(strides))
	for i, s := range strides {
		cStrides[i] = C.int64_t(s)
	}
	C.mlx_as_strided(&out.ctx, a.ctx, &cShape[0], C.size_t(len(cShape)), &cStrides[0], C.size_t(len(cStrides)), C.size_t(offset), DefaultStream().ctx)
	return out
}

// Take gathers elements from a along axis using indices.
func Take(a, indices *Array, axis int) *Array {
	out := New("TAKE", a, indices)
	C.mlx_take_axis(&out.ctx, a.ctx, indices.ctx, C.int(axis), DefaultStream().ctx)
	return out
}

// Where selects elements from a or b based on condition.
func Where(condition, a, b *Array) *Array {
	out := New("WHERE", condition, a, b)
	C.mlx_where(&out.ctx, condition.ctx, a.ctx, b.ctx, DefaultStream().ctx)
	return out
}

// Argpartition partially sorts and returns indices for top-k selection.
func Argpartition(a *Array, kth, axis int) *Array {
	out := New("ARGPARTITION", a)
	C.mlx_argpartition_axis(&out.ctx, a.ctx, C.int(kth), C.int(axis), DefaultStream().ctx)
	return out
}

// Dequantize restores a quantized array to full precision.
func Dequantize(w, scales, biases *Array, groupSize, bits int) *Array {
	out := New("DEQUANTIZE", w, scales, biases)
	gs := C.mlx_optional_int{value: C.int(groupSize), has_value: C._Bool(true)}
	b := C.mlx_optional_int{value: C.int(bits), has_value: C._Bool(true)}
	mode := C.CString("affine")
	defer C.free(unsafe.Pointer(mode))
	noDtype := C.mlx_optional_dtype{has_value: C._Bool(false)}
	C.mlx_dequantize(&out.ctx, w.ctx, scales.ctx, biases.ctx, gs, b, mode, noDtype, DefaultStream().ctx)
	return out
}

// PutAlongAxis places values into array at indices along axis.
func PutAlongAxis(a, indices, values *Array, axis int) *Array {
	out := New("PUT_ALONG_AXIS", a, indices, values)
	// Use scatter approach: src[indices] = values
	C.mlx_put_along_axis(&out.ctx, a.ctx, indices.ctx, values.ctx, C.int(axis), DefaultStream().ctx)
	return out
}
