package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"markupbook/markdown"
)

// App struct
type App struct {
	ctx           context.Context
	store         *markdown.Store
	token         string
	authenticated bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	// store located in ./markups relative to the binary
	storeDir := filepath.Join(".", "markups")
	t := os.Getenv("MARKUPBOOK_TOKEN")
	if t == "" {
		t = "changeme"
	}
	return &App{store: markdown.NewStore(storeDir), token: t}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Authenticate verifies the token and sets authenticated state
func (a *App) Authenticate(provided string) (bool, error) {
	if provided == a.token {
		a.authenticated = true
		return true, nil
	}
	return false, errors.New("invalid token")
}

func (a *App) requireAuth() error {
	if !a.authenticated {
		return errors.New("not authenticated")
	}
	return nil
}

// ListPages returns titles of pages
func (a *App) ListPages() ([]string, error) {
	if err := a.requireAuth(); err != nil {
		return nil, err
	}
	return a.store.ListPages()
}

// LoadPage returns the HTML content for a page title
func (a *App) LoadPage(title string) (string, error) {
	if err := a.requireAuth(); err != nil {
		return "", err
	}
	return a.store.LoadPage(title)
}

// SavePage replaces a section by title
func (a *App) SavePage(oldTitle, newTitle, html string) error {
	if err := a.requireAuth(); err != nil {
		return err
	}
	return a.store.SavePage(oldTitle, newTitle, html)
}

// SavePageIfMatch saves only if ETag matches
func (a *App) SavePageIfMatch(oldTitle, newTitle, html, expectedETag string) error {
	if err := a.requireAuth(); err != nil {
		return err
	}
	return a.store.SavePageIfMatch(oldTitle, newTitle, html, expectedETag)
}

// GetETag returns the current ETag for the notebook
func (a *App) GetETag() (string, error) {
	if err := a.requireAuth(); err != nil {
		return "", err
	}
	return a.store.ComputeETag()
}

// NewPage inserts a new section
func (a *App) NewPage(newTitle string) error {
	if err := a.requireAuth(); err != nil {
		return err
	}
	return a.store.InsertNewSection(newTitle)
}

// RenamePage renames a section
func (a *App) RenamePage(oldTitle, newTitle string) error {
	if err := a.requireAuth(); err != nil {
		return err
	}
	return a.store.RenameSection(oldTitle, newTitle)
}
