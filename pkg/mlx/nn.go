//go:build darwin && arm64 && mlx

package mlx

// Linear is a fully-connected layer: y = x @ W.T + bias.
type Linear struct {
	Weight *Array `weight:"weight"`
	Bias   *Array `weight:"bias"`
}

// NewLinear creates a Linear layer with optional bias.
func NewLinear(weight, bias *Array) *Linear {
	return &Linear{Weight: weight, Bias: bias}
}

// Forward computes the linear transformation.
func (l *Linear) Forward(x *Array) *Array {
	out := Matmul(x, Transpose(l.Weight))
	if l.Bias != nil && l.Bias.Valid() {
		out = Add(out, l.Bias)
	}
	return out
}

// Embedding is a lookup table for token embeddings.
type Embedding struct {
	Weight *Array `weight:"weight"`
}

// Forward looks up embeddings for the given token indices.
func (e *Embedding) Forward(indices *Array) *Array {
	return Take(e.Weight, indices, 0)
}

// RMSNormModule is an RMS normalization layer wrapping the fused kernel.
type RMSNormModule struct {
	Weight *Array `weight:"weight"`
}

// Forward applies RMS normalization.
func (r *RMSNormModule) Forward(x *Array, eps float32) *Array {
	return RMSNorm(x, r.Weight, eps)
}

// RepeatKV repeats key/value heads for grouped-query attention.
// Input shape: [B, num_kv_heads, L, D]
// Output shape: [B, num_kv_heads * factor, L, D]
func RepeatKV(x *Array, factor int32) *Array {
	if factor <= 1 {
		return x
	}
	shape := x.Shape()
	B, H, L, D := shape[0], shape[1], shape[2], shape[3]

	// Expand: [B, H, 1, L, D] then broadcast to [B, H, factor, L, D]
	expanded := ExpandDims(x, 2)
	expanded = BroadcastTo(expanded, []int32{B, H, factor, L, D})
	return Reshape(expanded, B, H*factor, L, D)
}
