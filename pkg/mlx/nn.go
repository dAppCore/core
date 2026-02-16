//go:build darwin && arm64 && mlx

package mlx

// Linear is a fully-connected layer: y = x @ W.T + bias.
// For quantized models, set Scales/Biases/GroupSize/Bits to use QuantizedMatmul.
type Linear struct {
	Weight    *Array `weight:"weight"`
	Scales    *Array `weight:"scales"`
	Biases    *Array `weight:"biases"`
	Bias      *Array `weight:"bias"`
	GroupSize int
	Bits      int
}

// NewLinear creates a dense Linear layer with optional bias.
func NewLinear(weight, bias *Array) *Linear {
	return &Linear{Weight: weight, Bias: bias}
}

// NewQuantizedLinear creates a quantized Linear layer.
func NewQuantizedLinear(weight, scales, biases, bias *Array, groupSize, bits int) *Linear {
	return &Linear{
		Weight:    weight,
		Scales:    scales,
		Biases:    biases,
		Bias:      bias,
		GroupSize: groupSize,
		Bits:      bits,
	}
}

// Forward computes the linear transformation.
// Uses QuantizedMatmul when quantization parameters are present.
func (l *Linear) Forward(x *Array) *Array {
	var out *Array
	if l.Scales != nil {
		out = QuantizedMatmul(x, l.Weight, l.Scales, l.Biases, true, l.GroupSize, l.Bits)
	} else {
		out = Matmul(x, Transpose(l.Weight))
	}
	if l.Bias != nil && l.Bias.Valid() {
		out = Add(out, l.Bias)
	}
	return out
}

// Embedding is a lookup table for token embeddings.
// For quantized models, set Scales/Biases/GroupSize/Bits to dequantize before lookup.
type Embedding struct {
	Weight    *Array `weight:"weight"`
	Scales    *Array `weight:"scales"`
	Biases    *Array `weight:"biases"`
	GroupSize int
	Bits      int
}

// Forward looks up embeddings for the given token indices.
func (e *Embedding) Forward(indices *Array) *Array {
	if e.Scales != nil {
		w := Dequantize(e.Weight, e.Scales, e.Biases, e.GroupSize, e.Bits)
		return Take(w, indices, 0)
	}
	return Take(e.Weight, indices, 0)
}

// AsLinear returns a Linear layer using the embedding weights (for tied output).
func (e *Embedding) AsLinear() *Linear {
	return &Linear{
		Weight:    e.Weight,
		Scales:    e.Scales,
		Biases:    e.Biases,
		GroupSize: e.GroupSize,
		Bits:      e.Bits,
	}
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
