//go:build darwin && arm64 && mlx

package mlx

/*
#include "mlx/c/mlx.h"
*/
import "C"

// RandomCategorical samples from a categorical distribution defined by logprobs.
// Returns indices sampled according to the log-probability distribution along the last axis.
func RandomCategorical(logprobs *Array) *Array {
	out := New("RANDOM_CATEGORICAL", logprobs)
	// shape for output: same as input but last dim removed
	C.mlx_random_categorical_shape(
		&out.ctx,
		logprobs.ctx,
		C.int(-1),   // axis
		nil, C.int(0), // empty shape = infer from input
		nil,            // key (use default)
		DefaultStream().ctx,
	)
	return out
}

// RandomUniform generates uniform random values in [low, high).
func RandomUniform(low, high float32, shape []int32, dtype DType) *Array {
	out := New("RANDOM_UNIFORM")
	cShape := make([]C.int, len(shape))
	for i, s := range shape {
		cShape[i] = C.int(s)
	}
	lo := FromValue(low)
	hi := FromValue(high)
	C.mlx_random_uniform(
		&out.ctx,
		lo.ctx, hi.ctx,
		&cShape[0], C.int(len(cShape)),
		C.mlx_dtype(dtype),
		nil, // key
		DefaultStream().ctx,
	)
	return out
}
