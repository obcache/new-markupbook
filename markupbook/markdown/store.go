package markdown

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

var h2Re = regexp.MustCompile(`(?m)^##\s+(.*)$`)

type Store struct {
	Dir string
}

func NewStore(dir string) *Store {
	return &Store{Dir: dir}
}

func (s *Store) ensureDir() error {
	return os.MkdirAll(s.Dir, 0o755)
}

func (s *Store) ReadAll() (string, error) {
	if err := s.ensureDir(); err != nil {
		return "", err
	}
	path := filepath.Join(s.Dir, "notebook.md")
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Store) WriteAll(text string) error {
	if err := s.ensureDir(); err != nil {
		return err
	}
	path := filepath.Join(s.Dir, "notebook.md")
	return os.WriteFile(path, []byte(text), 0o644)
}

type Section struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (s *Store) SplitSections(md string) []Section {
	matches := h2Re.FindAllStringSubmatchIndex(md, -1)
	if len(matches) == 0 {
		return nil
	}
	sections := make([]Section, 0, len(matches))
	for i, m := range matches {
		title := md[m[2]:m[3]]
		contentStart := m[1]
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(md)
		}
		content := md[contentStart:end]
		sections = append(sections, Section{Title: title, Content: content})
	}
	return sections
}

func (s *Store) ListPages() ([]string, error) {
	md, err := s.ReadAll()
	if err != nil {
		return nil, err
	}
	secs := s.SplitSections(md)
	titles := make([]string, 0, len(secs))
	for _, sec := range secs {
		titles = append(titles, sec.Title)
	}
	return titles, nil
}

func (s *Store) LoadPage(title string) (string, error) {
	md, err := s.ReadAll()
	if err != nil {
		return "", err
	}
	for _, sec := range s.SplitSections(md) {
		if sec.Title == title {
			return sec.Content, nil
		}
	}
	return "", errors.New("not found")
}

func (s *Store) SavePage(oldTitle, newTitle, html string) error {
	md, err := s.ReadAll()
	if err != nil {
		return err
	}
	sections := s.SplitSections(md)
	if len(sections) == 0 {
		return errors.New("no sections")
	}
	found := -1
	var start, end int
	// locate by scanning headings
	matches := h2Re.FindAllStringSubmatchIndex(md, -1)
	for i, sec := range sections {
		if sec.Title == oldTitle {
			found = i
			start = matches[i][0]
			if i+1 < len(matches) {
				end = matches[i+1][0]
			} else {
				end = len(md)
			}
			break
		}
	}
	if found == -1 {
		return errors.New("section not found")
	}
	newBlock := "## " + newTitle + "\n\n" + html + "\n"
	newMd := md[:start] + newBlock + md[end:]
	return s.WriteAll(newMd)
}

// ComputeETag returns a SHA256 hex of the current notebook contents.
func (s *Store) ComputeETag() (string, error) {
	md, err := s.ReadAll()
	if err != nil {
		return "", err
	}
	h := sha256.Sum256([]byte(md))
	return hex.EncodeToString(h[:]), nil
}

// SavePageIfMatch saves the page only if the current ETag matches expected.
// If expectedETag is empty, it saves unconditionally.
func (s *Store) SavePageIfMatch(oldTitle, newTitle, html, expectedETag string) error {
	if expectedETag != "" {
		etag, err := s.ComputeETag()
		if err != nil {
			return err
		}
		if etag != expectedETag {
			return errors.New("etag mismatch")
		}
	}
	return s.SavePage(oldTitle, newTitle, html)
}

func (s *Store) InsertNewSection(newTitle string) error {
	md, err := s.ReadAll()
	if err != nil {
		return err
	}
	stub := "\n\n## " + newTitle + "\n\n<p><em>New page.</em></p>\n"
	var out string
	if len(md) > 0 {
		out = md + stub
	} else {
		out = "# Notebook\n\n" + stub
	}
	return s.WriteAll(out)
}

func (s *Store) RenameSection(oldTitle, newTitle string) error {
	md, err := s.ReadAll()
	if err != nil {
		return err
	}
	matches := h2Re.FindAllStringSubmatchIndex(md, -1)
	sections := s.SplitSections(md)
	for i, sec := range sections {
		if sec.Title == oldTitle {
			start := matches[i][0]
			// find heading line end
			lineEnd := -1
			for j := start; j < len(md); j++ {
				if md[j] == '\n' {
					lineEnd = j
					break
				}
			}
			if lineEnd == -1 {
				lineEnd = len(md)
			}
			newLine := "## " + newTitle
			newMd := md[:start] + newLine + md[lineEnd:]
			return s.WriteAll(newMd)
		}
	}
	return errors.New("section not found")
}
