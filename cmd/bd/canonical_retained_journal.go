package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const canonicalRetainedJournalLimit int64 = 131072

// This is local custody, not a server principal or evidence of remote commit.
// Owner describes the OS/config/provider context, never a token or token digest.
// The server alone binds the idempotency namespace to its authenticated user.
type canonicalRetainedJournalRecord struct {
	Version  int                       `json:"version"`
	Owner    string                    `json:"credentialOwnerContext"`
	Provider string                    `json:"credentialProvider"`
	Request  canonicalRetainedPrepared `json:"request"`
}

// Only a durable save or a verified recovery load constructs a saved handle.
// The POST adapter has no arm accepting a plain prepared/preview request.
type canonicalRetainedSaved struct {
	path   string
	record canonicalRetainedJournalRecord
	bytes  []byte
}

func canonicalRetainedPrivateDirectory(root string, create bool) error {
	if !filepath.IsAbs(root) || filepath.Clean(root) != root {
		return fmt.Errorf("retained journal requires a clean absolute path")
	}
	// Inspect each component before creating or using it: MkdirAll would traverse
	// a symlink silently. Existing ancestors keep their own permissions; only the
	// dedicated journal directory is required to be private.
	var chain []string
	for p := root; ; p = filepath.Dir(p) {
		chain = append(chain, p)
		if filepath.Dir(p) == p {
			break
		}
	}
	for i := len(chain) - 1; i >= 0; i-- {
		p := chain[i]
		info, err := os.Lstat(p)
		if os.IsNotExist(err) && create {
			if err := os.Mkdir(p, 0700); err != nil && !os.IsExist(err) {
				return err
			}
			info, err = os.Lstat(p)
			if err == nil {
				// Durably publish every newly created directory in its parent.
				err = canonicalRetainedSyncDirectory(filepath.Dir(p))
			}
		}
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("retained journal path contains a non-directory or symlink")
		}
		if p == root && info.Mode().Perm()&0077 != 0 {
			return fmt.Errorf("retained journal directory must have mode 0700")
		}
	}
	return nil
}

func canonicalRetainedSyncDirectory(path string) error {
	directory, err := os.Open(path) // #nosec G304 -- only a journal directory or its parent selected by the checked private-directory/save path is synced.
	if err != nil {
		return err
	}
	syncErr := directory.Sync()
	closeErr := directory.Close()
	return errors.Join(syncErr, closeErr)
}

func canonicalRetainedJournalBytes(record canonicalRetainedJournalRecord) ([]byte, error) {
	if record.Version != 1 || record.Owner == "" || record.Request.Scope == "" || record.Request.Contract != canonicalRetainedContract || record.Request.Method != "POST" || !canonicalRetainedKey(record.Request.Key) {
		return nil, fmt.Errorf("incomplete retained invocation custody")
	}
	switch record.Request.Operation {
	case retainedCreateBead, retainedUpdateBead, retainedDeleteBead, retainedCreateLink, retainedUpdateLink, retainedDeleteLink:
	case retainedAtomicBatch:
		if _, err := canonicalRetainedBatchMembers(record.Request.Body); err != nil {
			return nil, err
		}
		if !canonicalRetainedBatchOwnedTarget(record.Request.Scope, record.Request.Target) {
			return nil, fmt.Errorf("batch recovery target is outside original Scope")
		}
	default:
		return nil, fmt.Errorf("journal accepts selected retained operations only")
	}
	if len(record.Request.Body) == 0 || len(record.Request.Body) > 65536 || !json.Valid(record.Request.Body) {
		return nil, fmt.Errorf("invalid retained request body")
	}
	raw, err := canonicalMutationMarshalNative(record)
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > canonicalRetainedJournalLimit {
		return nil, fmt.Errorf("retained journal exceeds its bound")
	}
	return append(raw, '\n'), nil
}

func saveCanonicalRetainedInvocation(root string, record canonicalRetainedJournalRecord) (canonicalRetainedSaved, error) {
	raw, err := canonicalRetainedJournalBytes(record)
	if err != nil {
		return canonicalRetainedSaved{}, err
	}
	if err := canonicalRetainedPrivateDirectory(root, true); err != nil {
		return canonicalRetainedSaved{}, err
	}
	digest := sha256.Sum256(raw)
	path := filepath.Join(root, hex.EncodeToString(digest[:])+".json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600) // #nosec G304 -- digest-only filename under the checked private root; exclusive create never overwrites existing custody.
	if os.IsExist(err) {
		saved, loadErr := loadCanonicalRetainedInvocation(root, path)
		if loadErr != nil {
			return canonicalRetainedSaved{}, loadErr
		}
		if !bytes.Equal(saved.bytes, raw) {
			return canonicalRetainedSaved{}, fmt.Errorf("retained invocation digest collision or changed bytes")
		}
		// A previous save may have stopped after file publication but before either
		// sync completed. Re-establish both barriers before admitting submission.
		existing, openErr := os.Open(path) // #nosec G304 -- same internally generated digest path; load above checked confinement, private mode, identity and exact bytes before durability barriers.
		if openErr != nil {
			return canonicalRetainedSaved{}, openErr
		}
		syncErr := existing.Sync()
		closeErr := existing.Close()
		if err := errors.Join(syncErr, closeErr, canonicalRetainedSyncDirectory(root)); err != nil {
			return canonicalRetainedSaved{}, err
		}
		return saved, nil
	}
	if err != nil {
		return canonicalRetainedSaved{}, err
	}
	n, writeErr := file.Write(raw)
	if writeErr == nil && n != len(raw) {
		writeErr = io.ErrShortWrite
	}
	var syncErr error
	if writeErr == nil {
		syncErr = file.Sync()
	}
	closeErr := file.Close()
	// Leave partial files untouched and fail closed. A later recovery load rejects
	// their digest; no overwrite, automatic resubmission or manufactured success.
	if err := errors.Join(writeErr, syncErr, closeErr); err != nil {
		return canonicalRetainedSaved{}, err
	}
	if err := canonicalRetainedSyncDirectory(root); err != nil {
		return canonicalRetainedSaved{}, err
	}
	return canonicalRetainedSaved{path: path, record: record, bytes: bytes.Clone(raw)}, nil
}

func loadCanonicalRetainedInvocation(root, path string) (canonicalRetainedSaved, error) {
	if err := canonicalRetainedPrivateDirectory(root, false); err != nil {
		return canonicalRetainedSaved{}, err
	}
	if !filepath.IsAbs(path) || filepath.Clean(path) != path || filepath.Dir(path) != root {
		return canonicalRetainedSaved{}, fmt.Errorf("recovery file is outside the admitted private journal")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return canonicalRetainedSaved{}, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 || info.Size() > canonicalRetainedJournalLimit {
		return canonicalRetainedSaved{}, fmt.Errorf("unsafe retained recovery file")
	}
	file, err := os.Open(path) // #nosec G304 -- clean absolute path confined to the checked private root; Lstat/Stat identity, private mode, bound and digest are checked before recovery admission.
	if err != nil {
		return canonicalRetainedSaved{}, err
	}
	opened, statErr := file.Stat()
	if statErr != nil || !os.SameFile(info, opened) {
		return canonicalRetainedSaved{}, errors.Join(statErr, file.Close(), fmt.Errorf("recovery file changed during open"))
	}
	raw, readErr := io.ReadAll(io.LimitReader(file, canonicalRetainedJournalLimit+1))
	closeErr := file.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return canonicalRetainedSaved{}, err
	}
	if int64(len(raw)) > canonicalRetainedJournalLimit {
		return canonicalRetainedSaved{}, fmt.Errorf("oversized retained recovery file")
	}
	digest := sha256.Sum256(raw)
	if filepath.Base(path) != hex.EncodeToString(digest[:])+".json" {
		return canonicalRetainedSaved{}, fmt.Errorf("retained recovery digest mismatch")
	}
	var record canonicalRetainedJournalRecord
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return canonicalRetainedSaved{}, err
	}
	encoded, err := canonicalRetainedJournalBytes(record)
	if err != nil {
		return canonicalRetainedSaved{}, err
	}
	if !bytes.Equal(raw, encoded) {
		return canonicalRetainedSaved{}, fmt.Errorf("recovery file is not the exact saved invocation")
	}
	return canonicalRetainedSaved{path: path, record: record, bytes: bytes.Clone(raw)}, nil
}
