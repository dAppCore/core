package ml

import (
	"context"
	"fmt"
	"time"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/cli/pkg/ml"
)

var (
	scoreInput   string
	scoreSuites  string
	scoreOutput  string
	scoreConcur  int
)

var scoreCmd = &cli.Command{
	Use:   "score",
	Short: "Score responses with heuristic and LLM judges",
	Long:  "Reads a JSONL file of prompt/response pairs and scores them across configured suites.",
	RunE:  runScore,
}

func init() {
	scoreCmd.Flags().StringVar(&scoreInput, "input", "", "Input JSONL file with prompt/response pairs (required)")
	scoreCmd.Flags().StringVar(&scoreSuites, "suites", "all", "Comma-separated scoring suites (heuristic,semantic,content,exact,truthfulqa,donotanswer,toxigen)")
	scoreCmd.Flags().StringVar(&scoreOutput, "output", "", "Output JSON file for scores")
	scoreCmd.Flags().IntVar(&scoreConcur, "concurrency", 4, "Number of concurrent scoring workers")
	scoreCmd.MarkFlagRequired("input")
}

func runScore(cmd *cli.Command, args []string) error {
	responses, err := ml.ReadResponses(scoreInput)
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	var judge *ml.Judge
	if judgeURL != "" {
		backend := ml.NewHTTPBackend(judgeURL, judgeModel)
		judge = ml.NewJudge(backend)
	}

	engine := ml.NewEngine(judge, scoreConcur, scoreSuites)

	ctx := context.Background()
	perPrompt := engine.ScoreAll(ctx, responses)
	averages := ml.ComputeAverages(perPrompt)

	if scoreOutput != "" {
		output := &ml.ScorerOutput{
			Metadata: ml.Metadata{
				JudgeModel: judgeModel,
				JudgeURL:   judgeURL,
				ScoredAt:   time.Now(),
				Suites:     ml.SplitComma(scoreSuites),
			},
			ModelAverages: averages,
			PerPrompt:     perPrompt,
		}
		if err := ml.WriteScores(scoreOutput, output); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		fmt.Printf("Scores written to %s\n", scoreOutput)
	} else {
		for model, avgs := range averages {
			fmt.Printf("%s:\n", model)
			for field, val := range avgs {
				fmt.Printf("  %-25s %.3f\n", field, val)
			}
		}
	}

	return nil
}
