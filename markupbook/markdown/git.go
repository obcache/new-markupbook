package markdown

import (
    "time"

    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/config"
    "github.com/go-git/go-git/v5/plumbing/object"
)

// ensureRepo ensures there is a git repo at dir and returns it
func ensureRepo(dir string) (*git.Repository, error) {
    // try opening
    r, err := git.PlainOpen(dir)
    if err == nil {
        return r, nil
    }
    // init new repo
    r, err = git.PlainInit(dir, false)
    if err != nil {
        return nil, err
    }
    // set default branch
    r.CreateBranch(&config.Branch{Name: "main", Merge: "refs/heads/main"})
    return r, nil
}

func CommitNotebook(dir, message string) error {
    // repository should be at dir
    r, err := ensureRepo(dir)
    if err != nil {
        return err
    }

    wt, err := r.Worktree()
    if err != nil {
        return err
    }

    // add file
    if err := wt.Add("notebook.md"); err != nil {
        return err
    }

    // commit
    _, err = wt.Commit(message, &git.CommitOptions{
        Author: &object.Signature{Name: "Markupbook", Email: "markupbook@local", When: time.Now()},
    })
    if err != nil {
        return err
    }

    // ensure HEAD exists
    _, err = r.Head()
    return err
}
