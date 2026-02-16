package forgejo

import (
	"regexp"
	"strconv"

	forgejosdk "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/cli/pkg/jobrunner"
)

// epicChildRe matches checklist items: - [ ] #42 or - [x] #42
var epicChildRe = regexp.MustCompile(`- \[([ x])\] #(\d+)`)

// parseEpicChildren extracts child issue numbers from an epic body's checklist.
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
func findLinkedPR(prs []*forgejosdk.PullRequest, issueNumber int) *forgejosdk.PullRequest {
	target := strconv.Itoa(issueNumber)
	for _, pr := range prs {
		matches := linkedPRRe.FindAllStringSubmatch(pr.Body, -1)
		for _, m := range matches {
			if m[1] == target {
				return pr
			}
		}
	}
	return nil
}

// mapPRState maps Forgejo's PR state and merged flag to a canonical string.
func mapPRState(pr *forgejosdk.PullRequest) string {
	if pr.HasMerged {
		return "MERGED"
	}
	switch pr.State {
	case forgejosdk.StateOpen:
		return "OPEN"
	case forgejosdk.StateClosed:
		return "CLOSED"
	default:
		return "CLOSED"
	}
}

// mapMergeable maps Forgejo's boolean Mergeable field to a canonical string.
func mapMergeable(pr *forgejosdk.PullRequest) string {
	if pr.HasMerged {
		return "UNKNOWN"
	}
	if pr.Mergeable {
		return "MERGEABLE"
	}
	return "CONFLICTING"
}

// mapCombinedStatus maps a Forgejo CombinedStatus to SUCCESS/FAILURE/PENDING.
func mapCombinedStatus(cs *forgejosdk.CombinedStatus) string {
	if cs == nil || cs.TotalCount == 0 {
		return "PENDING"
	}
	switch cs.State {
	case forgejosdk.StatusSuccess:
		return "SUCCESS"
	case forgejosdk.StatusFailure, forgejosdk.StatusError:
		return "FAILURE"
	default:
		return "PENDING"
	}
}

// buildSignal creates a PipelineSignal from Forgejo API data.
func buildSignal(
	owner, repo string,
	epicNumber, childNumber int,
	pr *forgejosdk.PullRequest,
	checkStatus string,
) *jobrunner.PipelineSignal {
	sig := &jobrunner.PipelineSignal{
		EpicNumber:  epicNumber,
		ChildNumber: childNumber,
		PRNumber:    int(pr.Index),
		RepoOwner:   owner,
		RepoName:    repo,
		PRState:     mapPRState(pr),
		IsDraft:     false, // SDK v2.2.0 doesn't expose Draft; treat as non-draft
		Mergeable:   mapMergeable(pr),
		CheckStatus: checkStatus,
	}

	if pr.Head != nil {
		sig.LastCommitSHA = pr.Head.Sha
	}

	return sig
}
