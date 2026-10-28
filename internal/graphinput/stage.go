// Package graphinput stages bounded opaque input in caller-owned temporary space.
// Staging preserves bytes; it does not parse, authorize or admit an import.
package graphinput

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
)

var (
	ErrNilContext    = errors.New("graph input requires a context")
	ErrNilReader     = errors.New("graph input requires a reader")
	ErrEmptyDir      = errors.New("graph input requires a staging directory")
	ErrInvalidLimit  = errors.New("graph input requires a positive limit below MaxInt64")
	ErrLimitExceeded = errors.New("graph input exceeds its byte limit")
)

const chunkBytes = 32 * 1024
const maxEmptyReads = 100

// Options chooses an existing caller-owned directory and an intake byte bound.
// MaxBytes is not a semantic admission limit: callers must also respect the
// downstream parser and transaction bounds. The helper never removes Dir.
type Options struct {
	Dir      string
	MaxBytes int64
}

type stageFile interface {
	io.ReadWriteSeeker
	io.Closer
	Sync() error
	Stat() (os.FileInfo, error)
	Name() string
}

// fileOperations is a private fault seam, not a configurable filesystem API.
type fileOperations struct {
	mkdir  func(string) (string, error)
	create func(string) (stageFile, error)
	open   func(string) (stageFile, error)
	remove func(string) error
}

func diskOperations() fileOperations {
	return fileOperations{
		mkdir:  func(parent string) (string, error) { return os.MkdirTemp(parent, "graph-input-") },
		create: func(dir string) (stageFile, error) { return os.CreateTemp(dir, "payload-") },
		open: func(path string) (stageFile, error) {
			// #nosec G304 -- path is this operation's newly created file in its private directory.
			return os.Open(path)
		},
		remove: os.Remove,
	}
}

// StagedInput exposes one read-only cursor over complete staged bytes. Seek
// supports multipass validation without rereading the caller's source. Callers
// must Close it; there is no finalizer. Crash residue belongs to the caller's
// staging-directory lifecycle. Cursor operations and Close are not concurrent.
type StagedInput struct {
	file     stageFile
	dir      string
	path     string
	size     int64
	remove   func(string) error
	closed   bool
	closeErr error
}

// Stage consumes at most MaxBytes+1 source bytes, requiring EOF before success.
// It never closes input. Context is checked at I/O boundaries; an arbitrary
// blocking Reader or filesystem syscall cannot be interrupted by this API.
// Files/directories use at most 0600/0700 Unix permissions, restricted further
// by umask; this is not a separate Windows ACL guarantee.
//
// Sync and writer Close surface delayed write failures before the read-only
// handle is published. They do not promise crash recovery or power-loss
// durability. Success guarantees untransformed bytes and matching write/stat
// counts, not a second content-verification read. All failures return nil and
// attempt nonrecursive cleanup of only this operation's file and directory.
func Stage(ctx context.Context, input io.Reader, options Options) (*StagedInput, error) {
	return stage(ctx, input, options, diskOperations())
}

func stage(ctx context.Context, input io.Reader, options Options, ops fileOperations) (_ *StagedInput, err error) {
	if ctx == nil {
		return nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if input == nil {
		return nil, ErrNilReader
	}
	if options.Dir == "" {
		return nil, ErrEmptyDir
	}
	if options.MaxBytes <= 0 || options.MaxBytes == math.MaxInt64 {
		return nil, ErrInvalidLimit
	}
	parent, err := filepath.Abs(options.Dir)
	if err != nil {
		return nil, fmt.Errorf("resolve staging directory: %w", err)
	}
	dir, err := ops.mkdir(parent)
	if err != nil {
		return nil, fmt.Errorf("create staging directory: %w", err)
	}
	s := &StagedInput{dir: dir, remove: ops.remove}
	defer func() {
		if err != nil {
			err = errors.Join(err, s.Close())
		}
	}()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.file, err = ops.create(dir)
	if err != nil {
		return nil, fmt.Errorf("create staging file: %w", err)
	}
	s.path = s.file.Name()
	if err := s.copy(ctx, input, options.MaxBytes); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.file.Sync(); err != nil {
		return nil, fmt.Errorf("sync staging file: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.file.Close(); err != nil {
		return nil, fmt.Errorf("close staging writer: %w", err)
	}
	s.file = nil
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.file, err = ops.open(s.path)
	if err != nil {
		return nil, fmt.Errorf("open staged input: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	info, err := s.file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat staged input: %w", err)
	}
	if info.Size() != s.size {
		return nil, fmt.Errorf("staged input size mismatch: wrote %d, stored %d", s.size, info.Size())
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *StagedInput) copy(ctx context.Context, input io.Reader, limit int64) error {
	var buffer [chunkBytes]byte
	emptyReads := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		// limit excludes MaxInt64, and size never exceeds limit.
		p := buffer[:int(min(int64(len(buffer)), limit-s.size+1))]
		n, readErr := input.Read(p)
		if n < 0 || n > len(p) {
			return fmt.Errorf("read staged input: invalid byte count %d for buffer %d", n, len(p))
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if int64(n) > limit-s.size {
			return ErrLimitExceeded
		}
		if n > 0 {
			emptyReads = 0
			written, writeErr := s.file.Write(p[:n])
			if written != n {
				writeErr = errors.Join(writeErr, io.ErrShortWrite)
			}
			if writeErr != nil {
				return fmt.Errorf("write staged input: %w", writeErr)
			}
			s.size += int64(written)
		} else if readErr == nil {
			emptyReads++
			if emptyReads >= maxEmptyReads {
				return io.ErrNoProgress
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return nil
			}
			return fmt.Errorf("read staged input: %w", readErr)
		}
	}
}

func (s *StagedInput) Read(p []byte) (int, error) {
	if s.closed {
		return 0, os.ErrClosed
	}
	return s.file.Read(p)
}

func (s *StagedInput) Seek(offset int64, whence int) (int64, error) {
	if s.closed {
		return 0, os.ErrClosed
	}
	return s.file.Seek(offset, whence)
}

// Size returns the immutable count of successfully staged bytes.
func (s *StagedInput) Size() int64 { return s.size }

// Close closes the cursor and attempts owned file/directory removal, preserving
// every error. Repeated calls return the original outcome, including failures;
// they do not claim failed physical cleanup has succeeded.
func (s *StagedInput) Close() error {
	if s.closed {
		return s.closeErr
	}
	s.closed = true
	if s.file != nil {
		if err := s.file.Close(); err != nil && !errors.Is(err, fs.ErrClosed) {
			s.closeErr = fmt.Errorf("close staged input: %w", err)
		}
	}
	for _, path := range []string{s.path, s.dir} {
		if path == "" {
			continue
		}
		if err := s.remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			s.closeErr = errors.Join(s.closeErr, fmt.Errorf("remove staged input: %w", err))
		}
	}
	return s.closeErr
}
