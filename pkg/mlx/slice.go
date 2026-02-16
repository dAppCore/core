//go:build darwin && arm64 && mlx

package mlx

/*
#include "mlx/c/mlx.h"
*/
import "C"

// Slice extracts a sub-array using start and end indices for each dimension.
// starts and ends must have the same length as the array's dimensions.
func Slice(a *Array, starts, ends []int32) *Array {
	out := New("SLICE", a)
	cStarts := make([]C.int, len(starts))
	cEnds := make([]C.int, len(ends))
	for i := range starts {
		cStarts[i] = C.int(starts[i])
		cEnds[i] = C.int(ends[i])
	}
	strides := make([]C.int, len(starts))
	for i := range strides {
		strides[i] = 1
	}
	C.mlx_slice(&out.ctx, a.ctx, &cStarts[0], C.int(len(cStarts)), &cEnds[0], C.int(len(cEnds)), &strides[0], C.int(len(strides)), DefaultStream().ctx)
	return out
}

// SliceAxis extracts a sub-array along a single axis.
func SliceAxis(a *Array, axis int, start, end int32) *Array {
	// Build full slice parameters
	ndim := a.NumDims()
	starts := make([]int32, ndim)
	ends := make([]int32, ndim)
	for i := 0; i < ndim; i++ {
		starts[i] = 0
		ends[i] = int32(a.Dim(i))
	}
	ax := axis
	if ax < 0 {
		ax = ndim + ax
	}
	starts[ax] = start
	ends[ax] = end
	return Slice(a, starts, ends)
}

// SliceUpdateInplace updates a slice of the array in-place.
// This is critical for KV cache updates.
func SliceUpdateInplace(a, update *Array, starts, ends []int32) *Array {
	out := New("SLICE_UPDATE", a, update)
	cStarts := make([]C.int, len(starts))
	cEnds := make([]C.int, len(ends))
	for i := range starts {
		cStarts[i] = C.int(starts[i])
		cEnds[i] = C.int(ends[i])
	}
	strides := make([]C.int, len(starts))
	for i := range strides {
		strides[i] = 1
	}
	C.mlx_slice_update(&out.ctx, a.ctx, update.ctx, &cStarts[0], C.int(len(cStarts)), &cEnds[0], C.int(len(cEnds)), &strides[0], C.int(len(strides)), DefaultStream().ctx)
	return out
}
