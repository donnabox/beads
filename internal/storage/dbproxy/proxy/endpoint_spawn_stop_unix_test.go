//go:build unix

package proxy

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spawnStopFailAfter only detects a hung start or Shutdown. It is far above
// shutdownSpawnWaitDeadline's components on purpose: a loaded -race runner
// can take seconds to fork/exec the child, and that must not fail the test.
const spawnStopFailAfter = 60 * time.Second

// holdStartBeforeExec makes the next proxy start block in its
// release-lock-before-exec window. entered closes once the start is parked
// there with its spawn marker published; calling release lets it exec.
func holdStartBeforeExec(t *testing.T, workers *sync.WaitGroup, observe func()) (root string, entered <-chan struct{}, release func()) {
	t.Helper()
	root, err := os.MkdirTemp("", "beads-spawn-stop-")
	require.NoError(t, err)
	childPath := filepath.Join(root, "exit-child.sh")

	previousResolve := ResolveExecutable
	previousHook := beforeProxyChildStart
	previousObserved := afterActiveSpawnMarker
	ResolveExecutable = func() (string, error) { return childPath, nil }
	enteredCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var once sync.Once
	var releaseOnce sync.Once
	release = func() { releaseOnce.Do(func() { close(releaseCh) }) }
	beforeProxyChildStart = func() {
		once.Do(func() {
			close(enteredCh)
			<-releaseCh
		})
	}
	if observe != nil {
		afterActiveSpawnMarker = observe
	}
	t.Cleanup(func() {
		release()
		joined := make(chan struct{})
		go func() {
			workers.Wait()
			close(joined)
		}()
		select {
		case <-joined:
			ResolveExecutable = previousResolve
			beforeProxyChildStart = previousHook
			afterActiveSpawnMarker = previousObserved
			require.NoError(t, os.RemoveAll(root))
		case <-time.After(spawnStopFailAfter):
			// Outstanding workers may still read these hooks or the root.
			// Preserve both rather than racing their restoration/removal.
			t.Errorf("owned spawn/stop workers did not join; retained hooks and root %s", root)
		}
	})
	require.NoError(t, os.WriteFile(childPath, []byte("#!/bin/sh\nexit 0\n"), 0o700))
	return root, enteredCh, release
}

// runShutdownDuringHeldStart parks a start before exec, runs Shutdown
// concurrently, holds the window for hold, then lets the start proceed and
// asserts the start aborts and Shutdown succeeds without leaving a marker.
func runShutdownDuringHeldStart(t *testing.T, hold time.Duration) {
	t.Helper()
	var workers sync.WaitGroup
	observed := make(chan struct{})
	var observedOnce sync.Once
	root, entered, release := holdStartBeforeExec(t, &workers, func() {
		observedOnce.Do(func() { close(observed) })
	})

	startDone := make(chan error, 1)
	workers.Add(1)
	go func() {
		defer workers.Done()
		_, err := GetCreateDatabaseProxyServerEndpoint(root, externalOpenOpts(root))
		startDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(spawnStopFailAfter):
		t.Fatal("start did not reach injected pre-exec delay")
	}

	marker, err := readSpawnMarker(root)
	require.NoError(t, err)
	require.NotNil(t, marker)
	t.Logf("held marker: %+v", marker)
	stopDone := make(chan error, 1)
	workers.Add(1)
	go func() {
		defer workers.Done()
		stopDone <- Shutdown(root)
	}()
	select {
	case <-observed:
	case err := <-stopDone:
		t.Fatalf("Shutdown returned before observing marker: %v", err)
	case <-time.After(spawnStopFailAfter):
		t.Fatal("Shutdown did not observe live starter")
	}
	epoch, err := readStopEpoch(root)
	require.NoError(t, err)
	require.NotEqual(t, marker.StopEpoch, epoch)
	t.Logf("shutdown observed live marker; stop epoch=%s", epoch)
	select {
	case err := <-stopDone:
		t.Fatalf("Shutdown returned during delayed start: %v", err)
	case <-time.After(hold):
		// The injected hook, rather than scheduler timing, holds the exact
		// historical release-lock-before-Start window open.
	}

	release()
	select {
	case err := <-startDone:
		t.Logf("starter finished: %v", err)
		remaining, markerErr := readSpawnMarker(root)
		t.Logf("marker after starter completion: %+v; read error: %v", remaining, markerErr)
		require.Error(t, err)
		assert.True(t, errors.Is(err, errStartInterrupted), "start error = %v", err)
	case <-time.After(spawnStopFailAfter):
		t.Fatal("start did not terminate after concurrent shutdown")
	}
	select {
	case err := <-stopDone:
		require.NoError(t, err)
	case <-time.After(spawnStopFailAfter):
		t.Fatal("Shutdown did not finish after delayed start resolved")
	}

	assert.NoFileExists(t, filepath.Join(root, spawnMarkerFileName))
}

func TestShutdownWaitsForStartDelayedBeforeExec(t *testing.T) {
	runShutdownDuringHeldStart(t, 150*time.Millisecond)
}

// A live starter that is slow to exec (loaded host, -race) must be waited
// out, not reported as "proxy left running" once shutdownConfirmDeadline
// elapses: the start is still making progress and will abort on the advanced
// stop epoch.
func TestShutdownWaitsForLiveStarterPastConfirmDeadline(t *testing.T) {
	runShutdownDuringHeldStart(t, shutdownConfirmDeadline+time.Second)
}

// Cleanup failures must not disappear behind the interruption they accompany.
// A malformed marker is deliberately installed only after the starter owns its
// original marker; this exercises the actual failed-start deferred cleanup.
func TestInterruptedStartReportsSpawnMarkerCleanupFailure(t *testing.T) {
	var workers sync.WaitGroup
	root, entered, release := holdStartBeforeExec(t, &workers, nil)
	startDone := make(chan error, 1)
	workers.Add(1)
	go func() {
		defer workers.Done()
		_, err := GetCreateDatabaseProxyServerEndpoint(root, externalOpenOpts(root))
		startDone <- err
	}()
	select {
	case <-entered:
	case <-time.After(spawnStopFailAfter):
		t.Fatal("start did not reach injected pre-exec delay")
	}
	markerPath := filepath.Join(root, spawnMarkerFileName)
	marker, err := readSpawnMarker(root)
	require.NoError(t, err)
	require.NotNil(t, marker)
	require.NoError(t, advanceStopEpoch(root))
	require.NoError(t, os.Remove(markerPath))
	require.NoError(t, os.Mkdir(markerPath, 0o700))
	release()
	select {
	case err := <-startDone:
		require.ErrorIs(t, err, errStartInterrupted)
		require.ErrorIs(t, err, syscall.EISDIR)
		require.ErrorContains(t, err, "clear own proxy spawn marker")
		require.ErrorContains(t, err, marker.Birth)
	case <-time.After(spawnStopFailAfter):
		t.Fatal("start did not terminate after interrupted cleanup")
	}
}
