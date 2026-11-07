// Package guide embeds the guide pages (docs/guide/<lang>/<page>.md).
package guide

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed en/*.md es/*.md
var files embed.FS

// Pages lists the available page names, e.g. "theft-before".
func Pages() []string {
	entries, _ := fs.Glob(files, "en/*.md")
	pages := make([]string, 0, len(entries))
	for _, e := range entries {
		pages = append(pages, strings.TrimSuffix(path.Base(e), ".md"))
	}
	sort.Strings(pages)
	return pages
}

// Page returns a page in a language ("en" or "es"), falling back to English.
func Page(lang, name string) (string, error) {
	for _, l := range []string{lang, "en"} {
		data, err := files.ReadFile(path.Join(l, name+".md"))
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("unknown guide page %q (available: %s)", name, strings.Join(Pages(), ", "))
}
