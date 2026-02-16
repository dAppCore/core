//go:build darwin && arm64 && mlx

// Package tokenizer provides BPE/SentencePiece tokenization for Gemma models.
package tokenizer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Tokenizer handles text-to-token and token-to-text conversion.
type Tokenizer struct {
	vocab    map[string]int32
	invVocab map[int32]string
	merges   []mergePair
	special  map[string]int32

	bosToken int32
	eosToken int32
}

type mergePair struct {
	a, b string
	rank int
}

// tokenizerJSON is the HuggingFace tokenizer.json format.
type tokenizerJSON struct {
	Model struct {
		Type     string          `json:"type"`
		Vocab    json.RawMessage `json:"vocab"`
		Merges   []string        `json:"merges"`
		ByteFallback bool        `json:"byte_fallback"`
	} `json:"model"`
	AddedTokens []struct {
		ID      int32  `json:"id"`
		Content string `json:"content"`
		Special bool   `json:"special"`
	} `json:"added_tokens"`
}

// Load reads a tokenizer.json file and creates a Tokenizer.
func Load(path string) (*Tokenizer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("tokenizer: read %s: %w", path, err)
	}

	var tj tokenizerJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return nil, fmt.Errorf("tokenizer: parse: %w", err)
	}

	t := &Tokenizer{
		vocab:    make(map[string]int32),
		invVocab: make(map[int32]string),
		special:  make(map[string]int32),
	}

	// Parse vocab
	var vocab map[string]int32
	if err := json.Unmarshal(tj.Model.Vocab, &vocab); err != nil {
		return nil, fmt.Errorf("tokenizer: parse vocab: %w", err)
	}
	t.vocab = vocab
	for k, v := range vocab {
		t.invVocab[v] = k
	}

	// Parse merges
	for rank, merge := range tj.Model.Merges {
		parts := strings.SplitN(merge, " ", 2)
		if len(parts) == 2 {
			t.merges = append(t.merges, mergePair{a: parts[0], b: parts[1], rank: rank})
		}
	}

	// Parse special tokens
	for _, tok := range tj.AddedTokens {
		if tok.Special {
			t.special[tok.Content] = tok.ID
		}
		t.vocab[tok.Content] = tok.ID
		t.invVocab[tok.ID] = tok.Content
	}

	// Set BOS/EOS
	if id, ok := t.special["<bos>"]; ok {
		t.bosToken = id
	}
	if id, ok := t.special["<eos>"]; ok {
		t.eosToken = id
	}
	if id, ok := t.special["<end_of_turn>"]; ok {
		t.eosToken = id // Gemma uses end_of_turn as EOS
	}

	return t, nil
}

// Encode converts text to token IDs. Prepends BOS token.
func (t *Tokenizer) Encode(text string) []int32 {
	tokens := []int32{t.bosToken}

	// Simple BPE encoding — split into characters then merge
	// This is a simplified version. Full implementation handles
	// Unicode, byte fallback, and efficient BPE merging.
	chars := []string{}
	for _, r := range text {
		s := string(r)
		if s == " " {
			s = "▁" // SentencePiece space marker
		}
		chars = append(chars, s)
	}

	// Check for special tokens first
	remaining := text
	for remaining != "" {
		found := false
		for tok, id := range t.special {
			if strings.HasPrefix(remaining, tok) {
				tokens = append(tokens, id)
				remaining = remaining[len(tok):]
				found = true
				break
			}
		}
		if !found {
			// Encode character by character (simplified BPE)
			r := []rune(remaining)
			ch := "▁" + string(r[0])
			if id, ok := t.vocab[ch]; ok {
				tokens = append(tokens, id)
			} else if id, ok := t.vocab[string(r[0])]; ok {
				tokens = append(tokens, id)
			}
			remaining = string(r[1:])
		}
	}

	return tokens
}

// Decode converts token IDs back to text.
func (t *Tokenizer) Decode(tokens []int32) string {
	var sb strings.Builder
	for _, id := range tokens {
		if text, ok := t.invVocab[id]; ok {
			// Replace SentencePiece space marker
			text = strings.ReplaceAll(text, "▁", " ")
			sb.WriteString(text)
		}
	}
	result := sb.String()
	// Trim leading space from SentencePiece encoding
	if strings.HasPrefix(result, " ") {
		result = result[1:]
	}
	return result
}

// BOSToken returns the beginning-of-sequence token ID.
func (t *Tokenizer) BOSToken() int32 { return t.bosToken }

// EOSToken returns the end-of-sequence token ID.
func (t *Tokenizer) EOSToken() int32 { return t.eosToken }

// FormatGemmaPrompt applies the Gemma 3 chat template.
func FormatGemmaPrompt(prompt string) string {
	return fmt.Sprintf("<start_of_turn>user\n%s<end_of_turn>\n<start_of_turn>model\n", prompt)
}
