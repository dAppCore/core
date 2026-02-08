package forge

import (
	"fmt"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"github.com/host-uk/core/pkg/cli"
	fg "github.com/host-uk/core/pkg/forge"
)

// Issues command flags.
var (
	issuesState string
	issuesTitle string
	issuesBody  string
)

// addIssuesCommand adds the 'issues' subcommand for listing and creating issues.
func addIssuesCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "issues [owner/repo]",
		Short: "List and manage issues",
		Long:  "List issues for a repository, or list all open issues across all your repos.",
		Args:  cli.MaximumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			if len(args) == 0 {
				return runListAllIssues()
			}

			owner, repo, err := splitOwnerRepo(args[0])
			if err != nil {
				return err
			}

			// If title is set, create an issue instead
			if issuesTitle != "" {
				return runCreateIssue(owner, repo)
			}

			return runListIssues(owner, repo)
		},
	}

	cmd.Flags().StringVar(&issuesState, "state", "open", "Filter by state (open, closed, all)")
	cmd.Flags().StringVar(&issuesTitle, "title", "", "Create issue with this title")
	cmd.Flags().StringVar(&issuesBody, "body", "", "Issue body (used with --title)")

	parent.AddCommand(cmd)
}

func runListAllIssues() error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	// Collect all repos: user repos + all org repos, deduplicated
	seen := make(map[string]bool)
	var allRepos []*forgejo.Repository

	userRepos, err := client.ListUserRepos()
	if err == nil {
		for _, r := range userRepos {
			if !seen[r.FullName] {
				seen[r.FullName] = true
				allRepos = append(allRepos, r)
			}
		}
	}

	orgs, err := client.ListMyOrgs()
	if err != nil {
		return err
	}

	for _, org := range orgs {
		repos, err := client.ListOrgRepos(org.UserName)
		if err != nil {
			continue
		}
		for _, r := range repos {
			if !seen[r.FullName] {
				seen[r.FullName] = true
				allRepos = append(allRepos, r)
			}
		}
	}

	total := 0
	cli.Blank()

	for _, repo := range allRepos {
		if repo.OpenIssues == 0 {
			continue
		}

		owner, name := repo.Owner.UserName, repo.Name
		issues, err := client.ListIssues(owner, name, fg.ListIssuesOpts{
			State: issuesState,
		})
		if err != nil || len(issues) == 0 {
			continue
		}

		cli.Print("  %s %s\n", repoStyle.Render(repo.FullName), dimStyle.Render(fmt.Sprintf("(%d)", len(issues))))
		for _, issue := range issues {
			printForgeIssue(issue)
		}
		cli.Blank()
		total += len(issues)
	}

	if total == 0 {
		cli.Text(fmt.Sprintf("No %s issues found.", issuesState))
	} else {
		cli.Print("  %s\n", dimStyle.Render(fmt.Sprintf("%d %s issues total", total, issuesState)))
	}
	cli.Blank()

	return nil
}

func runListIssues(owner, repo string) error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	issues, err := client.ListIssues(owner, repo, fg.ListIssuesOpts{
		State: issuesState,
	})
	if err != nil {
		return err
	}

	if len(issues) == 0 {
		cli.Text(fmt.Sprintf("No %s issues in %s/%s.", issuesState, owner, repo))
		return nil
	}

	cli.Blank()
	cli.Print("  %s\n\n", fmt.Sprintf("%d %s issues in %s/%s", len(issues), issuesState, owner, repo))

	for _, issue := range issues {
		printForgeIssue(issue)
	}

	return nil
}

func runCreateIssue(owner, repo string) error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	issue, err := client.CreateIssue(owner, repo, forgejo.CreateIssueOption{
		Title: issuesTitle,
		Body:  issuesBody,
	})
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Created issue #%d: %s", issue.Index, issue.Title))
	cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(issue.HTMLURL))
	cli.Blank()

	return nil
}

func printForgeIssue(issue *forgejo.Issue) {
	num := numberStyle.Render(fmt.Sprintf("#%d", issue.Index))
	title := valueStyle.Render(cli.Truncate(issue.Title, 60))

	line := fmt.Sprintf("  %s %s", num, title)

	// Add labels
	if len(issue.Labels) > 0 {
		var labels []string
		for _, l := range issue.Labels {
			labels = append(labels, l.Name)
		}
		line += " " + warningStyle.Render("["+strings.Join(labels, ", ")+"]")
	}

	// Add assignees
	if len(issue.Assignees) > 0 {
		var assignees []string
		for _, a := range issue.Assignees {
			assignees = append(assignees, "@"+a.UserName)
		}
		line += " " + infoStyle.Render(strings.Join(assignees, ", "))
	}

	cli.Text(line)
}
