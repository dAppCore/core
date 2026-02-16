package ml

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// ChatMessage is a single message in the chat training format.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// TrainingExample is a single training example in chat JSONL format.
type TrainingExample struct {
	Messages []ChatMessage `json:"messages"`
}

// ValidatePercentages checks that train+valid+test percentages sum to 100
// and that none are negative.
func ValidatePercentages(trainPct, validPct, testPct int) error {
	if trainPct < 0 || validPct < 0 || testPct < 0 {
		return fmt.Errorf("percentages must be non-negative: train=%d, valid=%d, test=%d", trainPct, validPct, testPct)
	}
	sum := trainPct + validPct + testPct
	if sum != 100 {
		return fmt.Errorf("percentages must sum to 100, got %d (train=%d + valid=%d + test=%d)", sum, trainPct, validPct, testPct)
	}
	return nil
}

// FilterResponses removes responses with empty content, "ERROR:" prefix,
// or response length < 50 characters.
func FilterResponses(responses []Response) []Response {
	var filtered []Response
	for _, r := range responses {
		if r.Response == "" {
			continue
		}
		if strings.HasPrefix(r.Response, "ERROR:") {
			continue
		}
		if len(r.Response) < 50 {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// SplitData shuffles responses with a deterministic seed and splits them
// into train, valid, and test sets by the given percentages.
func SplitData(responses []Response, trainPct, validPct, testPct int, seed int64) (train, valid, test []Response) {
	shuffled := make([]Response, len(responses))
	copy(shuffled, responses)

	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	n := len(shuffled)
	trainN := n * trainPct / 100
	validN := n * validPct / 100
	_ = testPct

	train = shuffled[:trainN]
	valid = shuffled[trainN : trainN+validN]
	test = shuffled[trainN+validN:]

	return train, valid, test
}

// WriteTrainingJSONL writes responses in chat JSONL format suitable for
// MLX LoRA fine-tuning.
func WriteTrainingJSONL(path string, responses []Response) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	for _, r := range responses {
		example := TrainingExample{
			Messages: []ChatMessage{
				{Role: "user", Content: r.Prompt},
				{Role: "assistant", Content: r.Response},
			},
		}

		data, err := json.Marshal(example)
		if err != nil {
			return fmt.Errorf("marshal example: %w", err)
		}

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("write line: %w", err)
		}
		if _, err := w.WriteString("\n"); err != nil {
			return fmt.Errorf("write newline: %w", err)
		}
	}

	return nil
}
