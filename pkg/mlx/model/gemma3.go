//go:build darwin && arm64 && mlx

// Package model provides transformer model architectures for MLX inference.
package model

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"

	"forge.lthn.ai/core/cli/pkg/mlx"
	"forge.lthn.ai/core/cli/pkg/mlx/cache"
	"forge.lthn.ai/core/cli/pkg/mlx/tokenizer"
)

// QuantizationConfig holds quantization parameters from config.json.
type QuantizationConfig struct {
	GroupSize int `json:"group_size"`
	Bits      int `json:"bits"`
}

// TextConfig holds Gemma 3 text model configuration.
type TextConfig struct {
	HiddenSize            int32   `json:"hidden_size"`
	NumHiddenLayers       int32   `json:"num_hidden_layers"`
	IntermediateSize      int32   `json:"intermediate_size"`
	NumAttentionHeads     int32   `json:"num_attention_heads"`
	NumKeyValueHeads      int32   `json:"num_key_value_heads"`
	HeadDim               int32   `json:"head_dim"`
	VocabSize             int32   `json:"vocab_size"`
	RMSNormEps            float32 `json:"rms_norm_eps"`
	RopeTheta             float32 `json:"rope_theta"`
	RopeLocalBaseFreq     float32 `json:"rope_local_base_freq"`
	MaxPositionEmbeddings int32   `json:"max_position_embeddings"`
	SlidingWindow         int32   `json:"sliding_window"`
	SlidingWindowPattern  int32   `json:"sliding_window_pattern"`

	Quantization *QuantizationConfig `json:"-"` // Parsed separately from top-level
	Scale        float32             `json:"-"` // Computed: 1/sqrt(head_dim)
}

// GemmaModel is the Gemma 3 text model.
type GemmaModel struct {
	EmbedTokens *mlx.Embedding
	Layers      []*DecoderLayer
	Norm        *mlx.RMSNormModule
	Output      *mlx.Linear // Tied to EmbedTokens

	// Precomputed (1 + weight) for Gemma-style RMSNorm
	NormScaled *mlx.Array

	Tok *tokenizer.Tokenizer
	Cfg *TextConfig
}

// DecoderLayer is a single transformer block.
type DecoderLayer struct {
	InputNorm    *mlx.RMSNormModule
	Attention    *Attention
	PostAttnNorm *mlx.RMSNormModule
	PreFFNorm    *mlx.RMSNormModule
	MLP          *MLP
	PostFFNorm   *mlx.RMSNormModule

	// Precomputed scaled weights
	InputNormScaled    *mlx.Array
	PostAttnNormScaled *mlx.Array
	PreFFNormScaled    *mlx.Array
	PostFFNormScaled   *mlx.Array

	IsSliding bool
	LayerIdx  int32
}

// Attention implements Gemma 3 attention with Q/K normalization.
type Attention struct {
	QProj *mlx.Linear
	KProj *mlx.Linear
	VProj *mlx.Linear
	OProj *mlx.Linear
	QNorm *mlx.RMSNormModule
	KNorm *mlx.RMSNormModule

	QNormScaled *mlx.Array
	KNormScaled *mlx.Array
}

// MLP is the feed-forward network.
type MLP struct {
	GateProj *mlx.Linear
	UpProj   *mlx.Linear
	DownProj *mlx.Linear
}

// compiledGELU is a singleton for the compiled GELU function.
var compiledGELU *mlx.CompiledFunc

func getCompiledGELU() *mlx.CompiledFunc {
	if compiledGELU == nil {
		compiledGELU = mlx.CompileShapeless(func(inputs []*mlx.Array) []*mlx.Array {
			return []*mlx.Array{geluApprox(inputs[0])}
		}, true)
	}
	return compiledGELU
}

// geluApprox computes GELU using the tanh approximation:
// 0.5 * x * (1 + tanh(sqrt(2/pi) * (x + 0.044715 * x^3)))
func geluApprox(x *mlx.Array) *mlx.Array {
	const sqrt2OverPi = 0.7978845608028654
	const coeff = 0.044715

	x3 := mlx.Mul(mlx.Mul(x, x), x)
	inner := mlx.Add(x, mlx.MulScalar(x3, coeff))
	scaled := mlx.MulScalar(inner, sqrt2OverPi)
	t := mlx.Tanh(scaled)
	onePlusT := mlx.AddScalar(t, 1.0)
	return mlx.Mul(mlx.MulScalar(x, 0.5), onePlusT)
}

// parseConfig handles both flat and nested (text_config) Gemma 3 configs.
func parseConfig(data []byte) (*TextConfig, error) {
	// Try parsing text_config from multimodal wrapper
	var wrapper struct {
		TextConfig   TextConfig          `json:"text_config"`
		ModelType    string              `json:"model_type"`
		Quantization *QuantizationConfig `json:"quantization"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}

	cfg := wrapper.TextConfig

	// If text_config was empty, try top-level
	if cfg.NumHiddenLayers == 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	// Quantization is always top-level
	cfg.Quantization = wrapper.Quantization

	// Compute scale (head_dim may be inferred later from weights if not in config)
	if cfg.HeadDim > 0 {
		cfg.Scale = float32(1.0 / math.Sqrt(float64(cfg.HeadDim)))
	}
	if cfg.RopeTheta == 0 {
		cfg.RopeTheta = 1000000
	}
	if cfg.RopeLocalBaseFreq == 0 {
		cfg.RopeLocalBaseFreq = 10000
	}
	if cfg.RMSNormEps == 0 {
		cfg.RMSNormEps = 1e-6
	}
	if cfg.SlidingWindowPattern == 0 {
		cfg.SlidingWindowPattern = 6
	}
	if cfg.VocabSize == 0 {
		cfg.VocabSize = 262208 // Gemma 3 default
	}

	return &cfg, nil
}

// resolveWeight looks up a weight with optional "language_model." prefix.
func resolveWeight(weights map[string]*mlx.Array, name string) *mlx.Array {
	if w, ok := weights[name]; ok {
		return w
	}
	if w, ok := weights["language_model."+name]; ok {
		return w
	}
	return nil
}

// LoadGemma3 loads a Gemma 3 text model from a directory.
func LoadGemma3(modelPath string) (*GemmaModel, error) {
	data, err := os.ReadFile(filepath.Join(modelPath, "config.json"))
	if err != nil {
		return nil, fmt.Errorf("gemma3: load config: %w", err)
	}

	cfg, err := parseConfig(data)
	if err != nil {
		return nil, fmt.Errorf("gemma3: parse config: %w", err)
	}

	// Load tokenizer
	tok, err := tokenizer.Load(filepath.Join(modelPath, "tokenizer.json"))
	if err != nil {
		return nil, fmt.Errorf("gemma3: load tokenizer: %w", err)
	}

	// Load weights from all safetensors files
	weights := make(map[string]*mlx.Array)
	matches, _ := filepath.Glob(filepath.Join(modelPath, "*.safetensors"))
	for _, path := range matches {
		for name, arr := range mlx.LoadSafetensors(path) {
			weights[name] = arr
		}
	}

	// Helper to resolve weight with language_model. prefix fallback
	w := func(name string) *mlx.Array { return resolveWeight(weights, name) }

	// Infer head_dim from q_proj weight shape when not in config.
	// Gemma 3 uses head_dim=256 which differs from hidden_size/num_heads.
	if cfg.HeadDim == 0 {
		qWeight := w("model.layers.0.self_attn.q_proj.weight")
		if qWeight != nil {
			qShape := qWeight.Shape()
			if len(qShape) > 0 {
				cfg.HeadDim = qShape[0] / cfg.NumAttentionHeads
				cfg.Scale = float32(1.0 / math.Sqrt(float64(cfg.HeadDim)))
				slog.Info("mlx: inferred head_dim from q_proj weight", "head_dim", cfg.HeadDim)
			}
		}
	}

	// Helper to create linear layer (quantized or dense)
	q := cfg.Quantization
	if q != nil {
		slog.Info("mlx: using quantized inference", "bits", q.Bits, "group_size", q.GroupSize)
	}
	linear := func(prefix string) *mlx.Linear {
		weight := w(prefix + ".weight")
		scales := w(prefix + ".scales")
		biases := w(prefix + ".biases")
		if scales != nil && q != nil {
			return mlx.NewQuantizedLinear(weight, scales, biases, nil, q.GroupSize, q.Bits)
		}
		return mlx.NewLinear(weight, nil)
	}

	// Create embedding (quantized or dense)
	embed := &mlx.Embedding{Weight: w("model.embed_tokens.weight")}
	if embedScales := w("model.embed_tokens.scales"); embedScales != nil && q != nil {
		embed.Scales = embedScales
		embed.Biases = w("model.embed_tokens.biases")
		embed.GroupSize = q.GroupSize
		embed.Bits = q.Bits
	}

	m := &GemmaModel{
		EmbedTokens: embed,
		Layers:      make([]*DecoderLayer, cfg.NumHiddenLayers),
		Norm:        &mlx.RMSNormModule{Weight: w("model.norm.weight")},
		Tok:         tok,
		Cfg:         cfg,
	}

	// Initialize layers
	for i := int32(0); i < cfg.NumHiddenLayers; i++ {
		prefix := fmt.Sprintf("model.layers.%d", i)
		m.Layers[i] = &DecoderLayer{
			InputNorm:    &mlx.RMSNormModule{Weight: w(prefix + ".input_layernorm.weight")},
			PostAttnNorm: &mlx.RMSNormModule{Weight: w(prefix + ".post_attention_layernorm.weight")},
			PreFFNorm:    &mlx.RMSNormModule{Weight: w(prefix + ".pre_feedforward_layernorm.weight")},
			PostFFNorm:   &mlx.RMSNormModule{Weight: w(prefix + ".post_feedforward_layernorm.weight")},
			Attention: &Attention{
				QProj: linear(prefix + ".self_attn.q_proj"),
				KProj: linear(prefix + ".self_attn.k_proj"),
				VProj: linear(prefix + ".self_attn.v_proj"),
				OProj: linear(prefix + ".self_attn.o_proj"),
				QNorm: &mlx.RMSNormModule{Weight: w(prefix + ".self_attn.q_norm.weight")},
				KNorm: &mlx.RMSNormModule{Weight: w(prefix + ".self_attn.k_norm.weight")},
			},
			MLP: &MLP{
				GateProj: linear(prefix + ".mlp.gate_proj"),
				UpProj:   linear(prefix + ".mlp.up_proj"),
				DownProj: linear(prefix + ".mlp.down_proj"),
			},
			LayerIdx:  i,
			IsSliding: isLayerSliding(i, cfg.SlidingWindowPattern),
		}
	}

	// Output head — check for separate lm_head first, else tie to embeddings
	lmHeadWeight := w("lm_head.weight")
	if lmHeadWeight != nil {
		lmHeadScales := w("lm_head.scales")
		if lmHeadScales != nil && q != nil {
			m.Output = mlx.NewQuantizedLinear(lmHeadWeight, lmHeadScales, w("lm_head.biases"), nil, q.GroupSize, q.Bits)
		} else {
			m.Output = mlx.NewLinear(lmHeadWeight, nil)
		}
	} else {
		// Tied embeddings — reuse embed_tokens weights (with quantization if present)
		m.Output = m.EmbedTokens.AsLinear()
	}

	// Materialize all weights
	var allArrays []*mlx.Array
	for _, a := range weights {
		allArrays = append(allArrays, a)
	}
	mlx.Materialize(allArrays...)

	// Precompute (1 + weight) for Gemma-style RMSNorm
	precomputeScaledWeights(m)

	return m, nil
}

func precomputeScaledWeights(m *GemmaModel) {
	m.NormScaled = mlx.AddScalar(m.Norm.Weight, 1.0)

	for _, layer := range m.Layers {
		layer.InputNormScaled = mlx.AddScalar(layer.InputNorm.Weight, 1.0)
		layer.PostAttnNormScaled = mlx.AddScalar(layer.PostAttnNorm.Weight, 1.0)
		layer.PreFFNormScaled = mlx.AddScalar(layer.PreFFNorm.Weight, 1.0)
		layer.PostFFNormScaled = mlx.AddScalar(layer.PostFFNorm.Weight, 1.0)
		layer.Attention.QNormScaled = mlx.AddScalar(layer.Attention.QNorm.Weight, 1.0)
		layer.Attention.KNormScaled = mlx.AddScalar(layer.Attention.KNorm.Weight, 1.0)
	}

	var scaled []*mlx.Array
	scaled = append(scaled, m.NormScaled)
	for _, layer := range m.Layers {
		scaled = append(scaled, layer.InputNormScaled, layer.PostAttnNormScaled,
			layer.PreFFNormScaled, layer.PostFFNormScaled,
			layer.Attention.QNormScaled, layer.Attention.KNormScaled)
	}
	mlx.Materialize(scaled...)
}

func isLayerSliding(layerIdx, pattern int32) bool {
	if pattern <= 0 {
		return false
	}
	return (layerIdx+1)%pattern != 0
}

// Forward runs the text model forward pass.
func (m *GemmaModel) Forward(tokens *mlx.Array, caches []cache.Cache) *mlx.Array {
	shape := tokens.Shape()
	B, L := shape[0], shape[1]

	h := m.EmbedTokens.Forward(tokens)
	h = mlx.MulScalar(h, float32(math.Sqrt(float64(m.Cfg.HiddenSize))))

	for i, layer := range m.Layers {
		h = layer.forward(h, caches[i], B, L, m.Cfg)
	}

	return m.Output.Forward(mlx.RMSNorm(h, m.NormScaled, m.Cfg.RMSNormEps))
}

func (l *DecoderLayer) forward(x *mlx.Array, c cache.Cache, B, L int32, cfg *TextConfig) *mlx.Array {
	normed := mlx.RMSNorm(x, l.InputNormScaled, cfg.RMSNormEps)
	attnOut := l.Attention.forward(normed, c, B, L, l.IsSliding, cfg)
	attnOut = mlx.RMSNorm(attnOut, l.PostAttnNormScaled, cfg.RMSNormEps)
	h := mlx.Add(x, attnOut)

	normed = mlx.RMSNorm(h, l.PreFFNormScaled, cfg.RMSNormEps)
	mlpOut := l.MLP.forward(normed)
	mlpOut = mlx.RMSNorm(mlpOut, l.PostFFNormScaled, cfg.RMSNormEps)
	return mlx.Add(h, mlpOut)
}

func (a *Attention) forward(x *mlx.Array, c cache.Cache, B, L int32, isSliding bool, cfg *TextConfig) *mlx.Array {
	q := a.QProj.Forward(x)
	k := a.KProj.Forward(x)
	v := a.VProj.Forward(x)

	// Reshape to [B, num_heads, L, head_dim]
	q = mlx.AsStrided(q, []int32{B, cfg.NumAttentionHeads, L, cfg.HeadDim},
		[]int64{int64(L * cfg.NumAttentionHeads * cfg.HeadDim), int64(cfg.HeadDim), int64(cfg.NumAttentionHeads * cfg.HeadDim), 1}, 0)
	k = mlx.AsStrided(k, []int32{B, cfg.NumKeyValueHeads, L, cfg.HeadDim},
		[]int64{int64(L * cfg.NumKeyValueHeads * cfg.HeadDim), int64(cfg.HeadDim), int64(cfg.NumKeyValueHeads * cfg.HeadDim), 1}, 0)
	v = mlx.AsStrided(v, []int32{B, cfg.NumKeyValueHeads, L, cfg.HeadDim},
		[]int64{int64(L * cfg.NumKeyValueHeads * cfg.HeadDim), int64(cfg.HeadDim), int64(cfg.NumKeyValueHeads * cfg.HeadDim), 1}, 0)

	// Q/K normalization
	q = mlx.RMSNorm(q, a.QNormScaled, cfg.RMSNormEps)
	k = mlx.RMSNorm(k, a.KNormScaled, cfg.RMSNormEps)

	// RoPE with appropriate theta
	ropeTheta := cfg.RopeTheta
	if isSliding {
		ropeTheta = cfg.RopeLocalBaseFreq
	}
	q = mlx.RoPE(q, int(cfg.HeadDim), false, ropeTheta, 1.0, c.Offset())
	k = mlx.RoPE(k, int(cfg.HeadDim), false, ropeTheta, 1.0, c.Offset())

	// Update cache
	k, v = c.Update(k, v, int(L))

	// GQA: repeat K/V heads
	repeatFactor := cfg.NumAttentionHeads / cfg.NumKeyValueHeads
	if repeatFactor > 1 {
		k = mlx.RepeatKV(k, repeatFactor)
		v = mlx.RepeatKV(v, repeatFactor)
	}

	// Scaled dot-product attention
	out := mlx.ScaledDotProductAttention(q, k, v, cfg.Scale, L > 1)
	out = mlx.Reshape(mlx.Transpose(out, 0, 2, 1, 3), B, L, cfg.NumAttentionHeads*cfg.HeadDim)
	return a.OProj.Forward(out)
}

func (m *MLP) forward(x *mlx.Array) *mlx.Array {
	gate := getCompiledGELU().Call(m.GateProj.Forward(x))[0]
	return m.DownProj.Forward(mlx.Mul(gate, m.UpProj.Forward(x)))
}

// NewCache creates per-layer caches for generation.
func (m *GemmaModel) NewCache() []cache.Cache {
	caches := make([]cache.Cache, len(m.Layers))
	for i := range caches {
		if m.Layers[i].IsSliding {
			caches[i] = cache.NewRotatingKVCache(int(m.Cfg.SlidingWindow))
		} else {
			caches[i] = cache.NewKVCache()
		}
	}
	return caches
}

// NumLayers returns the number of transformer layers.
func (m *GemmaModel) NumLayers() int { return len(m.Layers) }

// Tokenizer returns the model's tokenizer.
func (m *GemmaModel) Tokenizer() *tokenizer.Tokenizer { return m.Tok }
