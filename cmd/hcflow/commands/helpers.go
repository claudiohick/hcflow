// helpers.go — common setup for commands that need git/github/config context.
package commands

import (
	"fmt"
	"os"

	"github.com/hickstein/hcflow/internal/config"
	"github.com/hickstein/hcflow/internal/git"
	"github.com/hickstein/hcflow/internal/github"
)

// context bundles the common services used across commands.
type cmdContext struct {
	Dir    string
	Config *config.Config
	Git    *git.Service
	GitHub *github.Service
	Owner  string
	Repo   string
}

// loadContext loads config + creates git/github services from cwd.
// Returns an error if .hcflow.yml is not found.
func loadContext() (*cmdContext, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg, cfgPath, err := config.Load(dir)
	if err != nil {
		return nil, fmt.Errorf("%w\n  Tip: run 'hcflow init' to set up this repository", err)
	}

	root := config.RootDir(cfgPath)
	gitSvc := git.New(root)

	url := gitSvc.RemoteOrigin()
	owner, repo, parseErr := git.ParseGitHubRemote(url)
	if parseErr != nil {
		// non-fatal — some commands work without GitHub
		return &cmdContext{
			Dir:    root,
			Config: cfg,
			Git:    gitSvc,
		}, nil
	}

	ghSvc := github.New(root, owner, repo)
	return &cmdContext{
		Dir:    root,
		Config: cfg,
		Git:    gitSvc,
		GitHub: ghSvc,
		Owner:  owner,
		Repo:   repo,
	}, nil
}

// requireGitHub returns an error if GitHub service is not available.
func (c *cmdContext) requireGitHub() error {
	if c.GitHub == nil {
		return fmt.Errorf("no GitHub remote found — is this a GitHub repository?")
	}
	ok, _ := c.GitHub.IsAuthenticated()
	if !ok {
		return fmt.Errorf("not authenticated with GitHub — run: gh auth login")
	}
	return nil
}

// requireCleanOrConfirm checks working tree; if dirty, warns user.
func (c *cmdContext) isCleanTree() (bool, error) {
	return c.Git.IsClean()
}
