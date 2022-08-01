// Package yagfuncdata provides up-to-date information regarding the set of
// functions available in YAGPDB templates.
package yagfuncdata

import "context"

// Fetch retrieves the names of functions available in YAGPDB templates from the
// sources given. For fine-grained control over how sources are queried, use the
// Source.Fetch method directly.
func Fetch(ctx context.Context, sources []Source) (funcs []string, err error) {
	for _, src := range sources {
		result, err := src.Fetch(ctx)
		if err != nil {
			return nil, err
		}

		funcs = append(funcs, result...)
	}

	deduplicated := make([]string, 0, len(funcs))
	seen := make(map[string]bool, len(funcs))
	for _, name := range funcs {
		if !seen[name] {
			deduplicated = append(deduplicated, name)
			seen[name] = true
		}
	}
	return deduplicated, nil
}
