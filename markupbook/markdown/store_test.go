package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitSectionsAndList(t *testing.T) {
	tmp := t.TempDir()
	s := NewStore(tmp)
	md := "# Notebook\n\n## First\n\n<p>one</p>\n## Second\n\n<p>two</p>\n"
	if err := s.WriteAll(md); err != nil {
		t.Fatal(err)
	}
	pages, err := s.ListPages()
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) != 2 || pages[0] != "First" || pages[1] != "Second" {
		t.Fatalf("unexpected pages: %#v", pages)
	}
}

func TestLoadAndSave(t *testing.T) {
	tmp := t.TempDir()
	s := NewStore(tmp)
	md := "# Notebook\n\n## One\n\n<p>a</p>\n"
	if err := s.WriteAll(md); err != nil {
		t.Fatal(err)
	}
	content, err := s.LoadPage("One")
	if err != nil {
		t.Fatal(err)
	}
	if content == "" {
		t.Fatalf("empty content")
	}
	// save with rename
	if err := s.SavePage("One", "Uno", "<p>A</p>"); err != nil {
		t.Fatal(err)
	}
	pages, _ := s.ListPages()
	if len(pages) != 1 || pages[0] != "Uno" {
		t.Fatalf("rename failed: %#v", pages)
	}
}

func TestInsertAndRename(t *testing.T) {
	tmp := t.TempDir()
	s := NewStore(tmp)
	if err := s.InsertNewSection("NewPage"); err != nil {
		t.Fatal(err)
	}
	pages, _ := s.ListPages()
	if len(pages) != 1 || pages[0] != "NewPage" {
		t.Fatalf("insert failed: %#v", pages)
	}
	if err := s.RenameSection("NewPage", "Renamed"); err != nil {
		t.Fatal(err)
	}
	pages, _ = s.ListPages()
	if pages[0] != "Renamed" {
		t.Fatalf("rename failed: %#v", pages)
	}
}

func TestETagAndSaveIfMatch(t *testing.T) {
	tmp := t.TempDir()
	s := NewStore(tmp)
	if err := s.WriteAll("# Notebook\n\n## A\n\n<p>a</p>\n"); err != nil {
		t.Fatal(err)
	}
	etag, err := s.ComputeETag()
	if err != nil {
		t.Fatal(err)
	}
	// save with correct etag
	if err := s.SavePageIfMatch("A", "A", "<p>A</p>", etag); err != nil {
		t.Fatal(err)
	}
	// save with wrong etag should fail
	if err := s.SavePageIfMatch("A", "A", "<p>A</p>", "deadbeef"); err == nil {
		t.Fatalf("expected etag mismatch")
	}
}

func TestEdgeCases(t *testing.T) {
	tmp := t.TempDir()
	s := NewStore(tmp)

	// ETag of empty notebook should be consistent
	if err := s.WriteAll(""); err != nil {
		t.Fatal(err)
	}
	etag1, err := s.ComputeETag()
	if err != nil {
		t.Fatal(err)
	}
	// rewrite empty and recompute
	if err := s.WriteAll(""); err != nil {
		t.Fatal(err)
	}
	etag2, err := s.ComputeETag()
	if err != nil {
		t.Fatal(err)
	}
	if etag1 != etag2 {
		t.Fatalf("empty etag inconsistent: %s vs %s", etag1, etag2)
	}

	// SavePage should error when no sections exist
	if err := s.SavePage("X", "Y", "<p>x</p>"); err == nil {
		t.Fatalf("expected error saving with no sections")
	}

	// Load non-existent section
	if _, err := s.LoadPage("Nope"); err == nil {
		t.Fatalf("expected not found error")
	}

	// Simulate concurrent modification: get etag, then write file externally, then SavePageIfMatch should fail
	if err := s.WriteAll("# Notebook\n\n## A\n\n<p>a</p>\n"); err != nil {
		t.Fatal(err)
	}
	etag, _ := s.ComputeETag()
	// external write
	if err := os.WriteFile(filepath.Join(tmp, "notebook.md"), []byte("# Notebook\n\n## A\n\n<p>changed</p>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := s.SavePageIfMatch("A", "A", "<p>a</p>", etag); err == nil {
		t.Fatalf("expected etag mismatch after external change")
	}
}
