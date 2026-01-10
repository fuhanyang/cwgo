package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// HasGlobMeta reports whether s contains any filepath.Glob metacharacters.
// Note: filepath.Glob supports *, ?, and character classes ([]).
func HasGlobMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// ExpandIDLPaths expands an IDL path pattern into concrete file paths.
//
// If pattern does not contain glob metacharacters, it is returned as a single
// element (after existence check).
//
// If pattern contains glob metacharacters, it is expanded via filepath.Glob and
// returns all matches (sorted).
//
// Also supports a semicolon-separated list of patterns/paths, e.g.
// "a.proto;b.proto" or "./idl/*.proto;./more/*.proto".
func ExpandIDLPaths(patternOrList string) ([]string, error) {
	if strings.TrimSpace(patternOrList) == "" {
		return nil, fmt.Errorf("idl path is empty")
	}

	parts := strings.Split(patternOrList, ";")
	var all []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if !HasGlobMeta(part) {
			// Keep original behavior but validate existence early for clearer errors.
			if _, err := os.Stat(part); err != nil {
				return nil, fmt.Errorf("idl path %s not found: %w", part, err)
			}
			all = append(all, part)
			continue
		}

		matches, err := filepath.Glob(part)
		if err != nil {
			return nil, fmt.Errorf("invalid idl glob pattern %s: %w", part, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("idl glob pattern %s matched no files", part)
		}
		all = append(all, matches...)
	}

	if len(all) == 0 {
		return nil, fmt.Errorf("idl path is empty")
	}

	// De-dup + stable order.
	sort.Strings(all)
	out := make([]string, 0, len(all))
	var last string
	for _, p := range all {
		if p == last {
			continue
		}
		out = append(out, p)
		last = p
	}
	return out, nil
}

// SelectRootIDLByService tries to pick a single "root" IDL file from multiple
// candidates, using a best-effort scan for service definitions.
//
// This is mainly used for `--idl "*.proto"` style inputs where the user expects
// cwgo to pick the file that defines the target service.
func SelectRootIDLByService(candidates []string, serviceName string) (string, error) {
	if len(candidates) == 0 {
		return "", fmt.Errorf("no idl candidates")
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	type hit struct {
		path     string
		services map[string]struct{}
	}
	hits := make([]hit, 0, len(candidates))
	filesWithService := make([]string, 0, len(candidates))
	for _, p := range candidates {
		svcs, err := scanIDLServiceNames(p)
		if err != nil {
			return "", err
		}
		h := hit{path: p, services: svcs}
		hits = append(hits, h)
		if len(svcs) > 0 {
			filesWithService = append(filesWithService, p)
		}
	}

	// If serviceName is provided, prefer exact match.
	if strings.TrimSpace(serviceName) != "" {
		for _, h := range hits {
			if _, ok := h.services[serviceName]; ok {
				return h.path, nil
			}
		}
	}

	// Otherwise if only one file defines any service, pick it.
	if len(filesWithService) == 1 {
		return filesWithService[0], nil
	}

	// Ambiguous: ask user to specify one explicitly.
	if len(filesWithService) > 1 {
		return "", fmt.Errorf(
			"idl pattern matched %d files; multiple files define services (%s). please specify a single idl file",
			len(candidates),
			strings.Join(filesWithService, ", "),
		)
	}
	return "", fmt.Errorf(
		"idl pattern matched %d files but no service definition found; please specify a single idl file",
		len(candidates),
	)
}

var (
	reBlockComment = regexp.MustCompile(`(?s)/\*.*?\*/`)
	reLineComment  = regexp.MustCompile(`//.*`)
	reService      = regexp.MustCompile(`\bservice\s+([A-Za-z_][A-Za-z0-9_]*)\b`)
)

func scanIDLServiceNames(path string) (map[string]struct{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read idl %s failed: %w", path, err)
	}
	s := string(b)
	// Remove comments (best-effort) so "service" inside comments doesn't count.
	s = reBlockComment.ReplaceAllString(s, " ")
	s = reLineComment.ReplaceAllString(s, " ")

	matches := reService.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return map[string]struct{}{}, nil
	}
	out := make(map[string]struct{}, len(matches))
	for _, m := range matches {
		if len(m) >= 2 {
			out[m[1]] = struct{}{}
		}
	}
	return out, nil
}
