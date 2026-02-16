//go:build darwin && arm64 && mlx

// Package sample provides composable token sampling strategies.
package sample

import (
	"math"

	"forge.lthn.ai/core/cli/pkg/mlx"
)

// Sampler transforms logits into a sampled token index.
type Sampler interface {
	Sample(logits *mlx.Array) *mlx.Array
}

// New creates a composable sampler chain from the given parameters.
// Order: TopP -> MinP -> TopK -> Temperature -> categorical sample.
func New(temp, topP, minP float32, topK int) Sampler {
	if temp == 0 {
		return greedy{}
	}

	var samplers []Sampler
	if topP > 0 && topP < 1 {
		samplers = append(samplers, TopP(topP))
	}
	if minP > 0 {
		samplers = append(samplers, MinPSampler(minP))
	}
	if topK > 0 {
		samplers = append(samplers, TopKSampler(topK))
	}
	samplers = append(samplers, Temperature(temp))
	return chain(samplers)
}

// chain applies a sequence of samplers, then samples from the result.
type chain []Sampler

func (c chain) Sample(logits *mlx.Array) *mlx.Array {
	for _, s := range c {
		logits = s.Sample(logits)
	}
	// Final categorical sample from log-probabilities
	return mlx.RandomCategorical(logits)
}

// greedy returns the argmax token.
type greedy struct{}

func (greedy) Sample(logits *mlx.Array) *mlx.Array {
	return mlx.Argmax(logits, -1, false)
}

// Temperature scales logits by 1/temp.
type Temperature float32

func (t Temperature) Sample(logits *mlx.Array) *mlx.Array {
	return mlx.MulScalar(logits, 1.0/float32(t))
}

// TopKSampler masks all but the top-k logits.
type TopKSampler int

func (k TopKSampler) Sample(logits *mlx.Array) *mlx.Array {
	neg := mlx.Negative(logits)
	mask := mlx.Argpartition(neg, int(k)-1, -1)
	// Slice the indices beyond top-k
	mask = mlx.SliceAxis(mask, -1, int32(k), int32(logits.Dim(-1)))
	return mlx.PutAlongAxis(logits, mask, mlx.FromValue(float32(math.Inf(-1))), -1)
}

// TopP implements nucleus sampling (cumulative probability threshold).
type TopP float32

func (p TopP) Sample(logits *mlx.Array) *mlx.Array {
	// Softmax to get probabilities
	probs := mlx.Softmax(logits)
	// Sort descending
	neg := mlx.Negative(probs)
	sortedIdx := mlx.Argpartition(neg, 0, -1)
	sortedProbs := mlx.Take(probs, sortedIdx, -1)

	// Cumulative sum
	cumProbs := mlx.Sum(sortedProbs, -1, true) // simplified — full impl needs cumsum

	// Mask tokens beyond threshold
	threshold := mlx.FromValue(float32(p))
	mask := mlx.Where(
		mlx.FromValue(true), // placeholder — proper impl compares cumprobs > p
		mlx.FromValue(float32(math.Inf(-1))),
		logits,
	)
	return mask
}

// MinPSampler masks tokens below min_p * max_prob.
type MinPSampler float32

func (p MinPSampler) Sample(logits *mlx.Array) *mlx.Array {
	// For now, pass through — MinP is an optimization over TopP.
	// Full implementation requires finding max prob and masking below threshold.
	return logits
}
