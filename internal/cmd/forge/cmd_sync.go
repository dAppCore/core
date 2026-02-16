package forge

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	forgejo "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"

	"forge.lthn.ai/core/cli/pkg/cli"
	fg "forge.lthn.ai/core/cli/pkg/forge"
)

// Sync command flags.
var (
	syncOrg      string
	syncBasePath string
	syncSetup    bool
)

// addSyncCommand adds the 'sync' subcommand for syncing GitHub repos to Forgejo upstream branches.
func addSyncCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "sync <owner/repo> [owner/repo...]",
		Short: "Sync GitHub repos to Forgejo upstream branches",
		Long: `Push local GitHub content to Forgejo as 'upstream' branches.

Each repo gets:
  - An 'upstream' branch tracking the GitHub default branch
  - A 'main' branch (default) for private tasks, processes, and AI workflows

Use --setup on first run to create the Forgejo repos and configure remotes.
Without --setup, updates existing upstream branches from local clones.`,
		Args: cli.MinimumNArgs(0),
		RunE: func(cmd *cli.Command, args []string) error {
			return runSync(args)
		},
	}

	cmd.Flags().StringVar(&syncOrg, "org", "Host-UK", "Forgejo organisation")
	cmd.Flags().StringVar(&syncBasePath, "base-path", "~/Code/host-uk", "Base path for local repo clones")
	cmd.Flags().BoolVar(&syncSetup, "setup", false, "Initial setup: create repos, configure remotes, push upstream branches")

	parent.AddCommand(cmd)
}

// syncRepoEntry holds info for a repo to sync.
type syncRepoEntry struct {
	name          string
	localPath     string
	defaultBranch string
}

func runSync(args []string) error {
	client, err := fg.NewFromConfig("", "")
	if err != nil {
		return err
	}

	// Expand base path
	basePath := syncBasePath
	if strings.HasPrefix(basePath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to resolve home directory: %w", err)
		}
		basePath = filepath.Join(home, basePath[2:])
	}

	// Build repo list: either from args or from the Forgejo org
	repos, err := buildSyncRepoList(client, args, basePath)
	if err != nil {
		return err
	}

	if len(repos) == 0 {
		cli.Text("No repos to sync.")
		return nil
	}

	forgeURL := client.URL()

	if syncSetup {
		return runSyncSetup(client, repos, forgeURL)
	}

	return runSyncUpdate(repos, forgeURL)
}

func buildSyncRepoList(client *fg.Client, args []string, basePath string) ([]syncRepoEntry, error) {
	var repos []syncRepoEntry

	if len(args) > 0 {
		for _, arg := range args {
			name := arg
			if parts := strings.SplitN(arg, "/", 2); len(parts) == 2 {
				name = parts[1]
			}
			localPath := filepath.Join(basePath, name)
			branch := syncDetectDefaultBranch(localPath)
			repos = append(repos, syncRepoEntry{
				name:          name,
				localPath:     localPath,
				defaultBranch: branch,
			})
		}
	} else {
		orgRepos, err := client.ListOrgRepos(syncOrg)
		if err != nil {
			return nil, err
		}
		for _, r := range orgRepos {
			localPath := filepath.Join(basePath, r.Name)
			branch := syncDetectDefaultBranch(localPath)
			repos = append(repos, syncRepoEntry{
				name:          r.Name,
				localPath:     localPath,
				defaultBranch: branch,
			})
		}
	}

	return repos, nil
}

func runSyncSetup(client *fg.Client, repos []syncRepoEntry, forgeURL string) error {
	cli.Blank()
	cli.Print("  Setting up %d repos in %s with upstream branches...\n\n", len(repos), syncOrg)

	var succeeded, failed int

	for _, repo := range repos {
		cli.Print("  %s %s\n", dimStyle.Render(">>"), repoStyle.Render(repo.name))

		// Step 1: Delete existing repo if it exists
		cli.Print("     Deleting existing repo... ")
		err := client.DeleteRepo(syncOrg, repo.name)
		if err != nil {
			cli.Print("%s (may not exist)\n", dimStyle.Render("skipped"))
		} else {
			cli.Print("%s\n", successStyle.Render("done"))
		}

		// Step 2: Create empty repo
		cli.Print("     Creating repo... ")
		_, err = client.CreateOrgRepo(syncOrg, forgejo.CreateRepoOption{
			Name:          repo.name,
			AutoInit:      false,
			DefaultBranch: "main",
		})
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}
		cli.Print("%s\n", successStyle.Render("done"))

		// Step 3: Add forge remote to local clone
		cli.Print("     Configuring remote... ")
		remoteURL := fmt.Sprintf("%s/%s/%s.git", forgeURL, syncOrg, repo.name)
		err = syncConfigureForgeRemote(repo.localPath, remoteURL)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}
		cli.Print("%s\n", successStyle.Render("done"))

		// Step 4: Push default branch as 'upstream' to Forgejo
		cli.Print("     Pushing %s -> upstream... ", repo.defaultBranch)
		err = syncPushUpstream(repo.localPath, repo.defaultBranch)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}
		cli.Print("%s\n", successStyle.Render("done"))

		// Step 5: Create 'main' branch from 'upstream' on Forgejo
		cli.Print("     Creating main branch... ")
		err = syncCreateMainFromUpstream(client, syncOrg, repo.name)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "409") {
				cli.Print("%s\n", dimStyle.Render("exists"))
			} else {
				cli.Print("%s\n", errorStyle.Render(err.Error()))
				failed++
				continue
			}
		} else {
			cli.Print("%s\n", successStyle.Render("done"))
		}

		// Step 6: Set default branch to 'main'
		cli.Print("     Setting default branch... ")
		_, _, err = client.API().EditRepo(syncOrg, repo.name, forgejo.EditRepoOption{
			DefaultBranch: strPtr("main"),
		})
		if err != nil {
			cli.Print("%s\n", warningStyle.Render(err.Error()))
		} else {
			cli.Print("%s\n", successStyle.Render("main"))
		}

		succeeded++
		cli.Blank()
	}

	cli.Print("  %s", successStyle.Render(fmt.Sprintf("%d repos set up", succeeded)))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	cli.Blank()

	return nil
}

func runSyncUpdate(repos []syncRepoEntry, forgeURL string) error {
	cli.Blank()
	cli.Print("  Syncing %d repos to %s upstream branches...\n\n", len(repos), syncOrg)

	var succeeded, failed int

	for _, repo := range repos {
		cli.Print("  %s -> upstream  ", repoStyle.Render(repo.name))

		// Ensure remote exists
		remoteURL := fmt.Sprintf("%s/%s/%s.git", forgeURL, syncOrg, repo.name)
		_ = syncConfigureForgeRemote(repo.localPath, remoteURL)

		// Fetch latest from GitHub (origin)
		err := syncGitFetch(repo.localPath, "origin")
		if err != nil {
			cli.Print("%s\n", errorStyle.Render("fetch failed: "+err.Error()))
			failed++
			continue
		}

		// Push to Forgejo upstream branch
		err = syncPushUpstream(repo.localPath, repo.defaultBranch)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render(err.Error()))
			failed++
			continue
		}

		cli.Print("%s\n", successStyle.Render("ok"))
		succeeded++
	}

	cli.Blank()
	cli.Print("  %s", successStyle.Render(fmt.Sprintf("%d synced", succeeded)))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(fmt.Sprintf("%d failed", failed)))
	}
	cli.Blank()

	return nil
}

func syncDetectDefaultBranch(path string) string {
	out, err := exec.Command("git", "-C", path, "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		if parts := strings.Split(ref, "/"); len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}

	out, err = exec.Command("git", "-C", path, "branch", "--show-current").Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		if branch != "" {
			return branch
		}
	}

	return "main"
}

func syncConfigureForgeRemote(localPath, remoteURL string) error {
	out, err := exec.Command("git", "-C", localPath, "remote", "get-url", "forge").Output()
	if err == nil {
		existing := strings.TrimSpace(string(out))
		if existing != remoteURL {
			cmd := exec.Command("git", "-C", localPath, "remote", "set-url", "forge", remoteURL)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to update remote: %w", err)
			}
		}
		return nil
	}

	cmd := exec.Command("git", "-C", localPath, "remote", "add", "forge", remoteURL)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	return nil
}

func syncPushUpstream(localPath, defaultBranch string) error {
	refspec := fmt.Sprintf("refs/remotes/origin/%s:refs/heads/upstream", defaultBranch)
	cmd := exec.Command("git", "-C", localPath, "push", "--force", "forge", refspec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}

	return nil
}

func syncGitFetch(localPath, remote string) error {
	cmd := exec.Command("git", "-C", localPath, "fetch", remote)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

func syncCreateMainFromUpstream(client *fg.Client, org, repo string) error {
	_, _, err := client.API().CreateBranch(org, repo, forgejo.CreateBranchOption{
		BranchName:    "main",
		OldBranchName: "upstream",
	})
	if err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	return nil
}
