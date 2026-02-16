//go:build darwin && arm64 && mlx

package mlx

/*
#include <stdlib.h>
#include "mlx/c/mlx.h"
*/
import "C"

import "unsafe"

// RMSNorm applies Root Mean Square normalization using a fused Metal kernel.
func RMSNorm(x, weight *Array, eps float32) *Array {
	out := New("FAST_RMSNORM", x)
	C.mlx_fast_rms_norm(&out.ctx, x.ctx, weight.ctx, C.float(eps), DefaultStream().ctx)
	return out
}

// LayerNorm applies Layer normalization using a fused Metal kernel.
func LayerNorm(x, weight, bias *Array, eps float32) *Array {
	out := New("FAST_LAYERNORM", x)
	C.mlx_fast_layer_norm(&out.ctx, x.ctx, weight.ctx, bias.ctx, C.float(eps), DefaultStream().ctx)
	return out
}

// RoPE applies Rotary Position Embeddings using a fused Metal kernel.
func RoPE(x *Array, dims int, traditional bool, base float32, scale float32, offset int) *Array {
	freqs := New("")
	out := New("FAST_ROPE", x, freqs)
	C.mlx_fast_rope(
		&out.ctx,
		x.ctx,
		C.int(dims),
		C._Bool(traditional),
		C.mlx_optional_float{
			value:     C.float(base),
			has_value: C._Bool(base != 0),
		},
		C.float(scale),
		C.int(offset),
		freqs.ctx,
		DefaultStream().ctx,
	)
	return out
}

// ScaledDotProductAttention computes attention using a fused Metal kernel.
// mask can be nil for causal masking, or set causal=true for auto causal mask.
func ScaledDotProductAttention(query, key, value *Array, scale float32, causal bool) *Array {
	var mask, sinks *Array
	if causal {
		mask = New("")
		sinks = New("")
	} else {
		mask = New("")
		sinks = New("")
	}

	mode := "causal"
	if !causal {
		mode = "none"
	}
	cMode := C.CString(mode)
	defer C.free(unsafe.Pointer(cMode))

	out := New("FAST_SDPA", query, key, value, mask, sinks)
	C.mlx_fast_scaled_dot_product_attention(&out.ctx, query.ctx, key.ctx, value.ctx, C.float(scale), cMode, mask.ctx, sinks.ctx, DefaultStream().ctx)
	return out
}

// ScaledDotProductAttentionWithMask computes attention with an explicit mask.
func ScaledDotProductAttentionWithMask(query, key, value, mask *Array, scale float32) *Array {
	sinks := New("")
	cMode := C.CString("none")
	defer C.free(unsafe.Pointer(cMode))

	out := New("FAST_SDPA", query, key, value, mask, sinks)
	C.mlx_fast_scaled_dot_product_attention(&out.ctx, query.ctx, key.ctx, value.ctx, C.float(scale), cMode, mask.ctx, sinks.ctx, DefaultStream().ctx)
	return out
}
