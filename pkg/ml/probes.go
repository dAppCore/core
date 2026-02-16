package ml

import (
	"regexp"
	"strings"
)

// Probe defines a binary pass/fail capability check.
// Each probe sends a prompt to the model and evaluates the response
// with a Go function — no judge model needed.
type Probe struct {
	ID       string
	Category string
	Prompt   string
	Answer   string
	Check    func(response string) bool
}

// CapabilityProbes contains all 23 binary capability probes.
// Categories: arithmetic, algebra, probability, geometry, sequences,
// percentages, deduction, puzzles, sets, analogy, causal, spatial,
// temporal, pattern, code, word.
var CapabilityProbes = []Probe{
	// === MATH (8) ===
	{
		ID:       "math_01",
		Category: "arithmetic",
		Prompt:   "What is 347 × 29? Show your work and give the final answer.",
		Answer:   "10063",
		Check: func(r string) bool {
			clean := strings.ReplaceAll(strings.ReplaceAll(r, ",", ""), " ", "")
			return strings.Contains(clean, "10063")
		},
	},
	{
		ID:       "math_02",
		Category: "arithmetic",
		Prompt:   "A store sells apples for $1.25 each. If I buy 17 apples and pay with a $50 bill, how much change do I get?",
		Answer:   "28.75",
		Check: func(r string) bool {
			return strings.Contains(r, "28.75") || strings.Contains(r, "$28.75")
		},
	},
	{
		ID:       "math_03",
		Category: "algebra",
		Prompt:   "Solve for x: 3x + 7 = 2x - 5. What is x?",
		Answer:   "-12",
		Check: func(r string) bool {
			return regexp.MustCompile(`x\s*=\s*-\s*12|=\s*-12|-12`).MatchString(r)
		},
	},
	{
		ID:       "math_04",
		Category: "algebra",
		Prompt:   "If f(x) = 2x² - 3x + 1, what is f(4)?",
		Answer:   "21",
		Check: func(r string) bool {
			return regexp.MustCompile(`\b21\b`).MatchString(r)
		},
	},
	{
		ID:       "math_05",
		Category: "probability",
		Prompt:   "A bag has 3 red balls, 5 blue balls, and 2 green balls. What is the probability of drawing a blue ball? Express as a fraction and decimal.",
		Answer:   "1/2 or 0.5",
		Check: func(r string) bool {
			return strings.Contains(r, "1/2") || strings.Contains(r, "0.5") ||
				strings.Contains(r, "50%") || strings.Contains(r, "5/10")
		},
	},
	{
		ID:       "math_06",
		Category: "geometry",
		Prompt:   "A circle has a radius of 7cm. What is its area? Use pi = 3.14159.",
		Answer:   "153.94",
		Check: func(r string) bool {
			return regexp.MustCompile(`15[34]\.9|153\.9[0-9]|154\.0|49\s*[πpi]`).MatchString(r)
		},
	},
	{
		ID:       "math_07",
		Category: "sequences",
		Prompt:   "What is the next number in this sequence: 2, 6, 18, 54, ...?",
		Answer:   "162",
		Check: func(r string) bool {
			return strings.Contains(r, "162")
		},
	},
	{
		ID:       "math_08",
		Category: "percentages",
		Prompt:   "A laptop costs $800. It's on sale for 15% off. Then you have a coupon for 10% off the sale price. What is the final price?",
		Answer:   "612",
		Check: func(r string) bool {
			return regexp.MustCompile(`\$?612`).MatchString(r)
		},
	},
	// === LOGIC (5) ===
	{
		ID:       "logic_01",
		Category: "deduction",
		Prompt:   "All cats are animals. All animals need water. Does a cat need water? Explain your reasoning.",
		Answer:   "Yes",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)\byes\b`).MatchString(r)
		},
	},
	{
		ID:       "logic_02",
		Category: "deduction",
		Prompt:   "If it rains, the ground gets wet. The ground is wet. Can we conclude it rained? Why or why not?",
		Answer:   "No - affirming the consequent fallacy",
		Check: func(r string) bool {
			lower := strings.ToLower(r)
			return regexp.MustCompile(`\bno\b|\bcannot\b|\bcan't\b|not necessarily|fallac|other reason|doesn't mean`).MatchString(lower)
		},
	},
	{
		ID:       "logic_03",
		Category: "deduction",
		Prompt:   "In a room of 30 people, what is the minimum number of people that must share a birth month?",
		Answer:   "3",
		Check: func(r string) bool {
			lower := strings.ToLower(r)
			has3 := regexp.MustCompile(`\b3\b|three`).MatchString(lower)
			// Avoid matching "30" in the first 50 chars (restating the problem)
			prefix := lower
			if len(prefix) > 50 {
				prefix = prefix[:50]
			}
			has30 := regexp.MustCompile(`\b30\b`).MatchString(prefix)
			return has3 && !has30
		},
	},
	{
		ID:       "logic_04",
		Category: "puzzles",
		Prompt:   "A farmer needs to cross a river with a fox, a chicken, and a bag of grain. The boat only holds the farmer and one item. If left alone, the fox eats the chicken, and the chicken eats the grain. What is the first thing the farmer should take across?",
		Answer:   "The chicken",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)chicken|hen`).MatchString(r)
		},
	},
	{
		ID:       "logic_05",
		Category: "sets",
		Prompt:   "In a class of 40 students, 25 play football, 20 play basketball, and 10 play both. How many play neither?",
		Answer:   "5",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)\b5\b|five`).MatchString(r)
		},
	},
	// === REASONING (5) ===
	{
		ID:       "reason_01",
		Category: "analogy",
		Prompt:   "Complete the analogy: Book is to reading as fork is to ___",
		Answer:   "eating",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)eating|food|dining`).MatchString(r)
		},
	},
	{
		ID:       "reason_02",
		Category: "causal",
		Prompt:   "A car won't start. The battery is new. The fuel tank is full. The starter motor clicks but the engine doesn't turn. What is the most likely problem?",
		Answer:   "Starter motor / solenoid",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)starter|solenoid|connection|terminal|corros|ground|wire`).MatchString(r)
		},
	},
	{
		ID:       "reason_03",
		Category: "spatial",
		Prompt:   "You're facing north. You turn right 90 degrees, then turn right 90 degrees again. What direction are you facing?",
		Answer:   "South",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)\bsouth\b`).MatchString(r)
		},
	},
	{
		ID:       "reason_04",
		Category: "temporal",
		Prompt:   "Event A happened in 1995. Event B happened 12 years before Event A. Event C happened 8 years after Event B. In what year did Event C happen?",
		Answer:   "1991",
		Check: func(r string) bool {
			return strings.Contains(r, "1991")
		},
	},
	{
		ID:       "reason_05",
		Category: "pattern",
		Prompt:   "If APPLE = 50 (A=1, P=16, P=16, L=12, E=5), what does CAT equal using the same system?",
		Answer:   "24",
		Check: func(r string) bool {
			return regexp.MustCompile(`\b24\b`).MatchString(r)
		},
	},
	// === CODE (3) ===
	{
		ID:       "code_01",
		Category: "code",
		Prompt:   "What does this Python code print?\nx = [1, 2, 3, 4, 5]\nprint(x[1:3])",
		Answer:   "[2, 3]",
		Check: func(r string) bool {
			return strings.Contains(r, "[2, 3]") || strings.Contains(r, "[2,3]")
		},
	},
	{
		ID:       "code_02",
		Category: "code",
		Prompt:   "What is the output?\ndef f(n):\n    if n <= 1: return n\n    return f(n-1) + f(n-2)\nprint(f(6))",
		Answer:   "8",
		Check: func(r string) bool {
			return regexp.MustCompile(`\b8\b`).MatchString(r)
		},
	},
	{
		ID:       "code_03",
		Category: "code",
		Prompt:   "This code has a bug. What is it?\ndef average(numbers):\n    total = 0\n    for n in numbers:\n        total += n\n    return total / len(numbers)\nprint(average([]))",
		Answer:   "Division by zero",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)divis.*zero|zero.*divis|empty|len.*0|ZeroDivision`).MatchString(r)
		},
	},
	// === WORD PROBLEMS (2) ===
	{
		ID:       "word_01",
		Category: "word",
		Prompt:   "A train travels at 60 km/h. Another train travels at 80 km/h in the same direction from the same station, leaving 1 hour later. How long after the second train departs will it catch the first?",
		Answer:   "3 hours",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)\b3\b.*hour|three.*hour`).MatchString(r)
		},
	},
	{
		ID:       "word_02",
		Category: "word",
		Prompt:   "I have twice as many sisters as brothers. My sister has as many brothers as sisters. How many children are in my family? (I am male.)",
		Answer:   "7",
		Check: func(r string) bool {
			return regexp.MustCompile(`(?i)\b7\b|seven`).MatchString(r)
		},
	},
}

// ProbeCategories returns sorted unique categories from CapabilityProbes.
func ProbeCategories() []string {
	seen := make(map[string]bool)
	var cats []string
	for _, p := range CapabilityProbes {
		if !seen[p.Category] {
			seen[p.Category] = true
			cats = append(cats, p.Category)
		}
	}
	return cats
}

// StripThinkBlocks removes <think>...</think> blocks from DeepSeek R1 responses.
func StripThinkBlocks(s string) string {
	re := regexp.MustCompile(`(?s)<think>.*?</think>`)
	clean := strings.TrimSpace(re.ReplaceAllString(s, ""))
	if clean == "" && len(s) > 500 {
		return s[:500]
	}
	if clean == "" {
		return s
	}
	return clean
}
