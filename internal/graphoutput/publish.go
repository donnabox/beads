// Package graphoutput publishes detached, already validated export bytes.
// It owns neither the export format nor storage acquisition/cleanup. Callers
// must finish those operations successfully before invoking this package.
package graphoutput

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/steveyegge/beads/internal/atomicfile"
)

var (
	ErrNilContext = errors.New("graph output requires a context")
	ErrNilWriter  = errors.New("graph output requires a writer")
)

const writeChunkBytes = 32 * 1024

type publicationWriter interface {
	io.Writer
	Close() error
	Abort() error
}

type writerFactory func(string, os.FileMode) (publicationWriter, error)

// PublishFile atomically replaces path with data using caller-selected permissions.
// The caller owns data and must not modify it during publication. Cancellation
// is checked before creating the temporary file, before each write, and just
// before Close enters fsync/rename. Close itself is not interruptible: successful
// publication returns success even if cancellation races with or follows it.
// This protects readers from partial files, not power loss or missing directory
// fsync. Failures abort the owned writer, preserving operation and abort errors.
func PublishFile(ctx context.Context, path string, data []byte, perm os.FileMode) error {
	return publishFile(ctx, path, data, perm, func(path string, perm os.FileMode) (publicationWriter, error) {
		return atomicfile.Create(path, perm)
	})
}

func publishFile(ctx context.Context, path string, data []byte, perm os.FileMode, create writerFactory) (err error) {
	if err := contextError(ctx); err != nil {
		return err
	}
	w, err := create(path, perm)
	if err != nil {
		return fmt.Errorf("create graph output: %w", err)
	}
	defer func() {
		if err != nil {
			if abortErr := w.Abort(); abortErr != nil {
				err = errors.Join(err, fmt.Errorf("abort graph output: %w", abortErr))
			}
		}
	}()
	if err := WriteStream(ctx, w, data); err != nil {
		return err
	}
	if err := contextError(ctx); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("publish graph output: %w", err)
	}
	return nil
}

// WriteStream writes the entire detached byte slice, checking cancellation
// before each bounded write and reporting short writes. It does not close the
// caller's writer. For stdout/pipe output, bytes already delivered can remain
// after failure; there is no rollback. Writes are not interruptible. Once the
// final write succeeds, later cancellation does not turn complete output into
// a failure. The caller must not modify data during the call.
func WriteStream(ctx context.Context, w io.Writer, data []byte) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if w == nil {
		return ErrNilWriter
	}
	for start := 0; start < len(data); {
		if err := contextError(ctx); err != nil {
			return err
		}
		end := start + min(writeChunkBytes, len(data)-start)
		n, err := w.Write(data[start:end])
		if n != end-start {
			err = errors.Join(err, io.ErrShortWrite)
		}
		if err != nil {
			return fmt.Errorf("write graph output: %w", err)
		}
		start = end
	}
	return nil
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	return ctx.Err()
}
