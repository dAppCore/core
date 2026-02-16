package ml

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Pre-compiled regex patterns for GSM8K answer extraction.
var (
	// hashAnswer matches the #### delimiter pattern used in GSM8K.
	hashAnswer = regexp.MustCompile(`####\s*([\d,.\-]+)`)

	// lastNumber matches the last number in a response.
	lastNumber = regexp.MustCompile(`(?:^|[\s=])(-?[\d,]+(?:\.\d+)?)`)
)

// scoreGSM8K extracts a numeric answer from a model response and compares
// it to the correct answer using exact match (within epsilon of 0.01).
func scoreGSM8K(response, correctAnswer string) *StandardScores {
	correct := false

	// Empty or error response.
	if response == "" || strings.HasPrefix(response, "ERROR") {
		return &StandardScores{
			Correct:   &correct,
			Extracted: "",
			Expected:  correctAnswer,
		}
	}

	// Try #### delimiter first.
	var extracted string
	if m := hashAnswer.FindStringSubmatch(response); len(m) > 1 {
		extracted = m[1]
	} else {
		// Find the last number in the response.
		matches := lastNumber.FindAllStringSubmatch(response, -1)
		if len(matches) > 0 {
			extracted = matches[len(matches)-1][1]
		}
	}

	// No number found.
	if extracted == "" {
		return &StandardScores{
			Correct:   &correct,
			Extracted: "",
			Expected:  correctAnswer,
		}
	}

	// Clean commas and parse both numbers.
	cleanExtracted := strings.ReplaceAll(extracted, ",", "")
	cleanExpected := strings.ReplaceAll(correctAnswer, ",", "")

	extVal, errExt := strconv.ParseFloat(cleanExtracted, 64)
	expVal, errExp := strconv.ParseFloat(cleanExpected, 64)

	if errExt != nil || errExp != nil {
		return &StandardScores{
			Correct:   &correct,
			Extracted: extracted,
			Expected:  correctAnswer,
		}
	}

	correct = math.Abs(expVal-extVal) < 0.01

	return &StandardScores{
		Correct:   &correct,
		Extracted: extracted,
		Expected:  correctAnswer,
	}
}
