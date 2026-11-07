package check

import (
	"fmt"
	"io/fs"
	"path"
	"sort"

	"gopkg.in/yaml.v3"
)

// Load reads and validates every check of a platform from fsys
// (laid out as <platform>/<name>.yaml), sorted by ID.
func Load(fsys fs.FS, platform string) ([]Check, error) {
	files, err := fs.Glob(fsys, path.Join(platform, "*.yaml"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no checks for platform %q", platform)
	}
	seen := map[string]string{}
	checks := make([]Check, 0, len(files))
	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return nil, err
		}
		var c Check
		if err := yaml.Unmarshal(data, &c); err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		if c.Platform != platform {
			return nil, fmt.Errorf("%s: platform is %q, expected %q", file, c.Platform, platform)
		}
		if prev, dup := seen[c.ID]; dup {
			return nil, fmt.Errorf("%s: duplicate id %q (also in %s)", file, c.ID, prev)
		}
		seen[c.ID] = file
		if err := c.compile(); err != nil {
			return nil, fmt.Errorf("%s (%s): %w", file, c.ID, err)
		}
		checks = append(checks, c)
	}
	sort.Slice(checks, func(i, j int) bool { return checks[i].ID < checks[j].ID })
	return checks, nil
}
