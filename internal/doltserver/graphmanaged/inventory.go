package graphmanaged

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type filePin struct {
	path, digest string
	identity     os.FileInfo
	executable   bool
}
type rootPin struct {
	path     string
	identity os.FileInfo
}
type admitted struct {
	roots                                          []rootPin
	executable, cwd, profile, config, registration string
	environment                                    []string
	pins                                           []filePin
	directories                                    []directoryPin
}
type directoryPin struct {
	path     string
	entries  []string
	identity os.FileInfo
}

// inspect produces owned inert metadata, not approval to execute an adapter.
// No production adapter artifact is registered by this package.
func inspect(ctx context.Context, raw []byte) (admitted, error) {
	var a admitted
	if !platformSupported() {
		return a, errUnsupported
	}
	deadline := time.Now().Add(admissionDuration)
	if err := checkWork(ctx, deadline); err != nil {
		return a, err
	}
	b := budget{}
	d, err := decodeDescription(raw, &b)
	if err != nil {
		return a, err
	}
	if err = charge(&b.retained, int64(len(raw)), maxRetainedBytes, "description"); err != nil {
		return a, err
	}
	a.profile = d.ProfileSHA256
	h := sha256.Sum256(raw)
	a.config = hex.EncodeToString(h[:])
	a.cwd, err = decodeCanonical(ctx, deadline, d.Cwd, &b)
	if err != nil {
		return a, err
	}
	cwdInfo, err := checkedDirectory(ctx, deadline, a.cwd)
	if err != nil {
		return a, err
	}
	a.roots = append(a.roots, rootPin{a.cwd, cwdInfo})
	a.executable, err = decodeCanonical(ctx, deadline, d.Executable, &b)
	if err != nil {
		return a, err
	}
	roots := map[string]struct{}{}
	for _, encoded := range d.SecurityRoots {
		p, e := decodeCanonical(ctx, deadline, encoded, &b)
		if e != nil {
			return a, e
		}
		if _, ok := roots[p]; ok {
			return a, errInput
		}
		roots[p] = struct{}{}
		info, e := checkedDirectory(ctx, deadline, p)
		if e != nil {
			return a, e
		}
		a.roots = append(a.roots, rootPin{p, info})
	}
	if len(d.SecurityRoots) == 0 {
		return a, errInput
	}
	names := map[string]struct{}{}
	physical := []os.FileInfo{}
	for _, db := range d.Databases {
		if err = charge(&b.candidates, 1, maxCandidates, "candidates"); err != nil {
			return a, err
		}
		name := strings.ToLower(db.Name)
		if name == "" || name != db.Name || strings.TrimSpace(name) != name || db.Branch == "" {
			return a, errInput
		}
		if _, ok := names[name]; ok {
			return a, errInput
		}
		names[name] = struct{}{}
		p, e := decodeCanonical(ctx, deadline, db.Root, &b)
		if e != nil {
			return a, e
		}
		info, e := checkedDirectory(ctx, deadline, p)
		if e != nil {
			return a, e
		}
		for _, prior := range physical {
			if os.SameFile(info, prior) {
				return a, errInput
			}
		}
		physical = append(physical, info)
		a.roots = append(a.roots, rootPin{p, info})
	}
	if len(d.Databases) == 0 {
		return a, errInput
	}
	// Preserve declared order and all engine-owned strings; this is correlation,
	// not engine-specific normalization or parsing of repository state.
	registrations, err := json.Marshal(d.Databases)
	if err != nil {
		return a, err
	}
	if err = charge(&b.retained, int64(len(registrations)), maxRetainedBytes, "registrations"); err != nil {
		return a, err
	}
	h = sha256.Sum256(registrations)
	a.registration = hex.EncodeToString(h[:])
	usedEnv := map[string]struct{}{}
	for _, e := range d.Environment {
		if err = charge(&b.environment, 1, maxEnvironment, "environment entries"); err != nil {
			return a, err
		}
		if err = charge(&b.environmentBytes, int64(len(e.Name)+len(e.Value)+1), maxEnvironmentBytes, "environment bytes"); err != nil {
			return a, err
		}
		if _, ok := usedEnv[e.Name]; ok {
			return a, errInput
		}
		usedEnv[e.Name] = struct{}{}
		// A future source-qualified adapter may widen this closed manifest. Ambient
		// HOME, DOLT_*, loader, credential and debugging settings never pass here.
		switch e.Name {
		case "TMPDIR":
			if !filepath.IsAbs(e.Value) {
				return a, errInput
			}
			if _, err = checkedDirectory(ctx, deadline, e.Value); err != nil {
				return a, err
			}
		default:
			return a, errInput
		}
		if strings.ContainsAny(e.Name+e.Value, "\x00\r\n") {
			return a, errInput
		}
		if err = charge(&b.retained, int64(len(e.Name)+len(e.Value)+1), maxRetainedBytes, "environment"); err != nil {
			return a, err
		}
		a.environment = append(a.environment, e.Name+"="+e.Value)
	}
	for _, source := range d.Sources {
		if err = charge(&b.files, 1, maxFiles, "file occurrences"); err != nil {
			return a, err
		}
		p, e := decodeCanonical(ctx, deadline, source.Path, &b)
		if e != nil {
			return a, e
		}
		pin, e := hashFile(ctx, deadline, p, false, &b)
		if e != nil {
			return a, e
		}
		if !digestValid(source.SHA256) || pin.digest != source.SHA256 {
			return a, errInput
		}
		for _, prior := range a.pins {
			if os.SameFile(pin.identity, prior.identity) {
				return a, errInput
			}
		}
		a.pins = append(a.pins, pin)
	}
	pin, err := hashFile(ctx, deadline, a.executable, true, &b)
	if err != nil {
		return a, err
	}
	if pin.digest != d.ExecutableSHA256 {
		return a, errInput
	}
	a.pins = append(a.pins, pin)
	seenDirs := map[string]struct{}{}
	for _, dir := range d.Directories {
		p, e := decodeCanonical(ctx, deadline, dir.Path, &b)
		if e != nil {
			return a, e
		}
		rel, e := filepath.Rel(a.cwd, p)
		if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return a, errInput
		}
		if rel != "." && len(strings.Split(rel, string(filepath.Separator))) > maxDirectoryDepth {
			return a, fmt.Errorf("%w: directory depth", errBudget)
		}
		if _, ok := seenDirs[p]; ok {
			return a, errInput
		}
		seenDirs[p] = struct{}{}
		observed, e := inventoryDirectory(ctx, deadline, p, &b)
		if e != nil {
			return a, e
		}
		expected := make([]string, 0, len(dir.Entries))
		for _, encoded := range dir.Entries {
			name, e := nativePath(encoded, &b)
			if e != nil {
				return a, e
			}
			if filepath.Base(name) != name || name == "." || name == ".." {
				return a, errInput
			}
			expected = append(expected, name)
		}
		slices.Sort(expected)
		if !slices.Equal(expected, observed.entries) {
			return a, errInput
		}
		a.directories = append(a.directories, observed)
	}
	return a, checkWork(ctx, deadline)
}

func decodeCanonical(ctx context.Context, deadline time.Time, encoded string, b *budget) (string, error) {
	if err := checkWork(ctx, deadline); err != nil {
		return "", err
	}
	p, err := nativePath(encoded, b)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(p) || filepath.Clean(p) != p {
		return "", errInput
	}
	if len(strings.Split(p, string(filepath.Separator))) > maxComponents {
		return "", fmt.Errorf("%w: path components", errBudget)
	}
	current := string(filepath.Separator)
	for _, part := range strings.Split(strings.TrimPrefix(p, current), string(filepath.Separator)) {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)
		if err := checkWork(ctx, deadline); err != nil {
			return "", err
		}
		info, e := os.Lstat(current)
		if err := checkWork(ctx, deadline); err != nil {
			return "", errors.Join(e, err)
		}
		if e != nil {
			return "", e
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errInput
		}
	}
	return p, checkWork(ctx, deadline)
}

func hashStream(ctx context.Context, deadline time.Time, r io.Reader, limit int64, b *budget, executable bool) (string, int64, error) {
	h := sha256.New()
	buf := make([]byte, 64<<10)
	var consumed int64
	for {
		if err := checkWork(ctx, deadline); err != nil {
			return "", consumed, err
		}
		want := min(int64(len(buf)), limit-consumed+1)
		n, err := r.Read(buf[:want])
		if n > 0 {
			if e := charge(&consumed, int64(n), limit, "file bytes"); e != nil {
				return "", consumed, e
			}
			if !executable {
				if e := charge(&b.sources, int64(n), maxSourceBytes, "source bytes"); e != nil {
					return "", consumed, e
				}
			}
			if _, e := h.Write(buf[:n]); e != nil {
				return "", consumed, e
			}
		}
		if errors.Is(err, io.EOF) {
			return hex.EncodeToString(h.Sum(nil)), consumed, nil
		}
		if err != nil {
			return "", consumed, err
		}
		if n == 0 {
			return "", consumed, io.ErrNoProgress
		}
	}
}
func hashFile(ctx context.Context, deadline time.Time, path string, executable bool, b *budget) (pin filePin, result error) {
	if _, err := checkedDirectory(ctx, deadline, filepath.Dir(path)); err != nil {
		return pin, err
	}
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	f, err := openRegular(path)
	if err != nil {
		return pin, err
	}
	defer func() { result = errors.Join(result, f.Close()) }()
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	before, err := f.Stat()
	if err != nil {
		return pin, err
	}
	limit := int64(maxFileBytes)
	if executable {
		limit = maxExecutableBytes
	}
	if before.Size() < 0 || before.Size() > limit {
		return pin, fmt.Errorf("%w: file bytes", errBudget)
	}
	if before.Mode().Perm()&0022 != 0 || !ownedFile(before) || executable && before.Mode().Perm()&0111 == 0 {
		return pin, errInput
	}
	digest, n, err := hashStream(ctx, deadline, f, limit, b, executable)
	if err != nil {
		return pin, err
	}
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	after, err := f.Stat()
	if err != nil {
		return pin, err
	}
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	named, err := os.Lstat(path)
	if workErr := checkWork(ctx, deadline); workErr != nil {
		return pin, errors.Join(err, workErr)
	}
	if err != nil {
		return pin, err
	}
	if !sameStable(before, after) || !sameStable(after, named) || n != after.Size() {
		return pin, errInput
	}
	return filePin{path: path, digest: digest, identity: after, executable: executable}, nil
}
func sameStable(a, b os.FileInfo) bool {
	return os.SameFile(a, b) && a.Mode() == b.Mode() && a.Size() == b.Size() && a.ModTime() == b.ModTime()
}
func inventoryDirectory(ctx context.Context, deadline time.Time, path string, b *budget) (pin directoryPin, result error) {
	before, err := checkedDirectory(ctx, deadline, path)
	if err != nil {
		return pin, err
	}
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	f, err := openDirectory(path)
	if err != nil {
		return pin, err
	}
	defer func() { result = errors.Join(result, f.Close()) }()
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	opened, err := f.Stat()
	if err != nil || !sameStable(before, opened) {
		return pin, errors.Join(err, errInput)
	}
	pin = directoryPin{path: path, identity: before}
	for {
		if err = checkWork(ctx, deadline); err != nil {
			return pin, err
		}
		entries, e := f.ReadDir(64)
		for _, entry := range entries {
			if err = charge(&b.entries, 1, maxEntries, "directory entries"); err != nil {
				return pin, err
			}
			name := entry.Name()
			if len(filepath.Join(path, name)) > maxPathBytes {
				return pin, fmt.Errorf("%w: joined path", errBudget)
			}
			if err = charge(&b.paths, int64(len(name)), maxPathsBytes, "directory paths"); err != nil {
				return pin, err
			}
			if err = charge(&b.retained, int64(len(name)), maxRetainedBytes, "directory names"); err != nil {
				return pin, err
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return pin, errInput
			}
			pin.entries = append(pin.entries, name)
		}
		if errors.Is(e, io.EOF) {
			break
		}
		if e != nil {
			return pin, e
		}
	}
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	after, err := f.Stat()
	if err != nil {
		return pin, err
	}
	if err := checkWork(ctx, deadline); err != nil {
		return pin, err
	}
	named, err := os.Lstat(path)
	if workErr := checkWork(ctx, deadline); workErr != nil {
		return pin, errors.Join(err, workErr)
	}
	if err != nil {
		return pin, err
	}
	if !sameStable(before, after) || !sameStable(after, named) {
		return pin, errInput
	}
	slices.Sort(pin.entries)
	return pin, nil
}
func (a admitted) recheck(ctx context.Context) error {
	b := budget{}
	deadline := time.Now().Add(admissionDuration)
	if err := checkWork(ctx, deadline); err != nil {
		return err
	}
	for _, pin := range a.roots {
		got, err := checkedDirectory(ctx, deadline, pin.path)
		if err != nil {
			return err
		}
		if !os.SameFile(pin.identity, got) || pin.identity.Mode() != got.Mode() {
			return errInput
		}
	}
	for _, pin := range a.pins {
		got, err := hashFile(ctx, deadline, pin.path, pin.executable, &b)
		if err != nil {
			return err
		}
		if !sameStable(pin.identity, got.identity) || pin.digest != got.digest {
			return errInput
		}
	}
	for _, pin := range a.directories {
		got, err := inventoryDirectory(ctx, deadline, pin.path, &b)
		if err != nil {
			return err
		}
		if !sameStable(pin.identity, got.identity) || !slices.Equal(pin.entries, got.entries) {
			return errInput
		}
	}
	return checkWork(ctx, deadline)
}

func checkedDirectory(ctx context.Context, deadline time.Time, path string) (os.FileInfo, error) {
	if err := checkWork(ctx, deadline); err != nil {
		return nil, err
	}
	info, err := trustedDirectory(path)
	if workErr := checkWork(ctx, deadline); workErr != nil {
		return nil, errors.Join(err, workErr)
	}
	return info, err
}
