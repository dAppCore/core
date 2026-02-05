package github

import (
	"regexp"
	"strconv"
	"time"

	"github.com/host-uk/core/pkg/jobrunner"
)

// ghIssue is a minimal GitHub issue response.
type ghIssue struct {
	Number int       `json:"number"`
	Title  string    `json:"title"`
	Body   string    `json:"body"`
	Labels []ghLabel `json:"labels"`
	State  string    `json:"state"`
}

// ghLabel is a GitHub label.
type ghLabel struct {
	Name string `json:"name"`
}

// ghPR is a minimal GitHub pull request response.
type ghPR struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	State          string `json:"state"`
	Draft          bool   `json:"draft"`
	MergeableState string `json:"mergeable_state"`
	Head           ghRef  `json:"head"`
}

// ghRef is a Git reference (branch head).
type ghRef struct {
	SHA string `json:"sha"`
	Ref string `json:"ref"`
}

// ghCheckSuites is the response for the check-suites endpoint.
type ghCheckSuites struct {
	TotalCount  int            `json:"total_count"`
	CheckSuites []ghCheckSuite `json:"check_suites"`
}

// ghCheckSuite is a single check suite.
type ghCheckSuite struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`     // queued, in_progress, completed
	Conclusion string `json:"conclusion"` // success, failure, neutral, cancelled, etc.
}

// epicChildRe matches checklist items in epic bodies: - [ ] #42 or - [x] #42
var epicChildRe = regexp.MustCompile(`- \[([ x])\] #(\d+)`)

// parseEpicChildren extracts child issue numbers from an epic body's checklist.
// Returns two slices: unchecked (pending) and checked (done) issue numbers.
func parseEpicChildren(body string) (unchecked []int, checked []int) {
	matches := epicChildRe.FindAllStringSubmatch(body, -1)
	for _, m := range matches {
		num, err := strconv.Atoi(m[2])
		if err != nil {
			continue
		}
		if m[1] == "x" {
			checked = append(checked, num)
		} else {
			unchecked = append(unchecked, num)
		}
	}
	return unchecked, checked
}

// linkedPRRe matches "#N" references in PR bodies.
var linkedPRRe = regexp.MustCompile(`#(\d+)`)

// findLinkedPR finds the first PR whose body references the given issue number.
func findLinkedPR(prs []ghPR, issueNumber int) *ghPR {
	target := strconv.Itoa(issueNumber)
	for i := range prs {
		matches := linkedPRRe.FindAllStringSubmatch(prs[i].Body, -1)
		for _, m := range matches {
			if m[1] == target {
				return &prs[i]
			}
		}
	}
	return nil
}

// aggregateCheckStatus returns SUCCESS, FAILURE, or PENDING based on check suites.
func aggregateCheckStatus(suites []ghCheckSuite) string {
	if len(suites) == 0 {
		return "PENDING"
	}

	allComplete := true
	for _, s := range suites {
		if s.Status != "completed" {
			allComplete = false
			break
		}
	}

	if !allComplete {
		return "PENDING"
	}

	for _, s := range suites {
		if s.Conclusion != "success" && s.Conclusion != "neutral" && s.Conclusion != "skipped" {
			return "FAILURE"
		}
	}

	return "SUCCESS"
}

// mergeableToString maps GitHub's mergeable_state to a canonical string.
func mergeableToString(state string) string {
	switch state {
	case "clean", "has_hooks", "unstable":
		return "MERGEABLE"
	case "dirty", "blocked":
		return "CONFLICTING"
	default:
		return "UNKNOWN"
	}
}

// buildSignal creates a PipelineSignal from parsed GitHub API data.
func buildSignal(
	owner, repo string,
	epicNumber, childNumber int,
	pr *ghPR,
	checkStatus string,
) *jobrunner.PipelineSignal {
	prState := "OPEN"
	switch pr.State {
	case "closed":
		prState = "CLOSED"
	case "open":
		prState = "OPEN"
	}

	return &jobrunner.PipelineSignal{
		EpicNumber:    epicNumber,
		ChildNumber:   childNumber,
		PRNumber:      pr.Number,
		RepoOwner:     owner,
		RepoName:      repo,
		PRState:       prState,
		IsDraft:       pr.Draft,
		Mergeable:     mergeableToString(pr.MergeableState),
		CheckStatus:   checkStatus,
		LastCommitSHA: pr.Head.SHA,
		LastCommitAt:  time.Time{}, // Not available from list endpoint
		LastReviewAt:  time.Time{}, // Not available from list endpoint
	}
}
