package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Repo struct {
	Root      string
	Name      string
	RemoteURL string
	Head      string
	BaseRef   string
}

func findRepo() (Repo, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		wd, _ := os.Getwd()
		return Repo{Root: wd, Name: filepath.Base(wd)}, nil
	}
	root := strings.TrimSpace(string(out))
	return Repo{
		Root:      root,
		Name:      filepath.Base(root),
		RemoteURL: gitOutput(root, "remote", "get-url", "origin"),
		Head:      gitOutput(root, "rev-parse", "HEAD"),
		BaseRef:   defaultBaseRef(root),
	}, nil
}

func defaultExcludes() []string {
	return []string{
		"node_modules",
		".turbo",
		".next",
		"dist",
		"dist-runtime",
		".cache",
		".pnpm-store",
		".git/lfs",
		".git/logs",
		".git/rr-cache",
		".git/worktrees",
	}
}

func configuredExcludes(cfg Config) []string {
	return appendUniqueStrings(defaultExcludes(), cfg.Sync.Excludes...)
}

func allowedEnv(allow []string) map[string]string {
	out := map[string]string{}
	for _, env := range os.Environ() {
		k, v, ok := strings.Cut(env, "=")
		if !ok {
			continue
		}
		if envAllowed(k, allow) {
			out[k] = v
		}
	}
	return out
}

func envAllowed(name string, allow []string) bool {
	for _, pattern := range allow {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.HasSuffix(pattern, "*") {
			if strings.HasPrefix(name, strings.TrimSuffix(pattern, "*")) {
				return true
			}
			continue
		}
		if name == pattern {
			return true
		}
	}
	return false
}

func gitOutput(root string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func defaultBaseRef(root string) string {
	originHead := gitOutput(root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	if originHead != "" {
		return strings.TrimPrefix(originHead, "origin/")
	}
	branch := gitOutput(root, "branch", "--show-current")
	if branch != "" {
		return branch
	}
	return ""
}

func syncFingerprint(repo Repo, cfg Config) (string, error) {
	if repo.Head == "" {
		return "", nil
	}
	paths, err := changedSyncPaths(repo.Root, configuredExcludes(cfg))
	if err != nil {
		return "", err
	}
	h := sha256.New()
	fmt.Fprintf(h, "v2\nhead=%s\n", repo.Head)
	fmt.Fprintf(h, "delete=%t\nchecksum=%t\n", cfg.Sync.Delete, cfg.Sync.Checksum)
	for _, exclude := range configuredExcludes(cfg) {
		fmt.Fprintf(h, "exclude=%s\n", exclude)
	}
	for _, rel := range paths {
		fmt.Fprintf(h, "path=%s\n", rel)
		full := filepath.Join(repo.Root, filepath.FromSlash(rel))
		info, err := os.Lstat(full)
		if err != nil {
			fmt.Fprintf(h, "missing\n")
			continue
		}
		fmt.Fprintf(h, "mode=%s size=%d\n", info.Mode().String(), info.Size())
		if info.IsDir() {
			continue
		}
		file, err := os.Open(full)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(h, file); err != nil {
			_ = file.Close()
			return "", err
		}
		_ = file.Close()
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func changedSyncPaths(root string, excludes []string) ([]string, error) {
	sets := [][]string{}
	for _, args := range [][]string{
		{"diff", "--name-only", "-z"},
		{"diff", "--cached", "--name-only", "-z"},
		{"ls-files", "--others", "--exclude-standard", "-z"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		sets = append(sets, splitNul(out))
	}
	seen := map[string]bool{}
	for _, set := range sets {
		for _, rel := range set {
			rel = filepath.ToSlash(rel)
			if rel == "" || pathExcluded(rel, excludes) {
				continue
			}
			seen[rel] = true
		}
	}
	out := make([]string, 0, len(seen))
	for rel := range seen {
		out = append(out, rel)
	}
	sort.Strings(out)
	return out, nil
}

func splitNul(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	parts := bytes.Split(data, []byte{0})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 0 {
			out = append(out, string(part))
		}
	}
	return out
}

func pathExcluded(rel string, excludes []string) bool {
	rel = filepath.ToSlash(rel)
	for _, exclude := range excludes {
		exclude = strings.Trim(filepath.ToSlash(strings.TrimSpace(exclude)), "/")
		if exclude == "" {
			continue
		}
		if rel == exclude || strings.HasPrefix(rel, exclude+"/") {
			return true
		}
		if ok, _ := filepath.Match(exclude, filepath.Base(rel)); ok {
			return true
		}
		if ok, _ := filepath.Match(exclude, rel); ok {
			return true
		}
	}
	return false
}
