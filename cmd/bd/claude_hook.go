package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	claudeHookStop           = "Stop"
	claudeStopMemoryReminder = "Before you finish: if this session produced anything a later session in this project will need, save it now with `bd remember \"<fact>\"`. If there is nothing to save, just finish."
)

var (
	claudeHookMarkerDirOverride   string
	errClaudeTranscriptUnreadable = errors.New("read Claude transcript")
)

type claudeHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	HookEventName  string `json:"hook_event_name"`
	StopHookActive bool   `json:"stop_hook_active"`
}

type claudeStopDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

type claudeTranscriptEntry struct {
	Type    string `json:"type"`
	Message struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

var claudeHookCmd = &cobra.Command{
	Use:         "claude-hook <event>",
	Hidden:      true,
	Short:       "Run an internal Claude Code lifecycle hook",
	Args:        cobra.ExactArgs(1),
	Annotations: map[string]string{skipStoreAnnotation: "1"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runClaudeHook(cmd.Context(), args[0], os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(claudeHookCmd)
}

func runClaudeHook(_ context.Context, event string, stdin io.Reader, stdout io.Writer) error {
	var input claudeHookInput
	if err := json.NewDecoder(stdin).Decode(&input); err != nil && err != io.EOF {
		return err
	}
	if input.HookEventName != "" {
		event = input.HookEventName
	}
	if !strings.EqualFold(event, claudeHookStop) {
		return fmt.Errorf("unsupported Claude hook event %q", event)
	}
	return claudeHookStopReminder(input, stdout)
}

func claudeHookStopReminder(input claudeHookInput, stdout io.Writer) error {
	if input.TranscriptPath == "" {
		return errors.New("claude Stop hook input has no transcript_path")
	}
	markerPath := agentHookMarkerPath(claudeHookMarkerBaseDir(), input.SessionID, input.TranscriptPath)
	offset, markerErr := readClaudeStopMarker(markerPath)
	end, usedTools, scanErr := claudeTranscriptToolUseSince(input.TranscriptPath, offset)
	if errors.Is(scanErr, errClaudeTranscriptUnreadable) {
		return errors.Join(markerErr, scanErr)
	}
	if err := writeClaudeStopMarker(markerPath, end); err != nil {
		return errors.Join(markerErr, scanErr, err)
	}
	if err := errors.Join(markerErr, scanErr); err != nil {
		return err
	}
	if input.StopHookActive || !usedTools {
		return nil
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(claudeStopDecision{
		Decision: "block",
		Reason:   claudeStopMemoryReminder,
	})
}

func claudeTranscriptToolUseSince(path string, offset int64) (int64, bool, error) {
	file, err := os.Open(path) // #nosec G304 -- path is the transcript Claude Code names for this session
	if err != nil {
		return 0, false, fmt.Errorf("%w: %w", errClaudeTranscriptUnreadable, err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return 0, false, fmt.Errorf("%w: %w", errClaudeTranscriptUnreadable, err)
	}
	if offset > info.Size() {
		offset = 0
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return 0, false, fmt.Errorf("%w: %w", errClaudeTranscriptUnreadable, err)
	}
	reader := bufio.NewReader(file)
	end := offset
	usedTools := false
	var parseErrs []error
	for {
		line, err := reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			return end, usedTools, errors.Join(parseErrs...)
		}
		if err != nil {
			return 0, false, fmt.Errorf("%w: %w", errClaudeTranscriptUnreadable, err)
		}
		lineOffset := end
		end += int64(len(line))
		hasTool, err := claudeTranscriptLineUsesTool(bytes.TrimSpace(line))
		if err != nil {
			parseErrs = append(parseErrs, fmt.Errorf("parse Claude transcript %s at byte %d: %w", path, lineOffset, err))
			continue
		}
		usedTools = usedTools || hasTool
	}
}

func claudeTranscriptLineUsesTool(line []byte) (bool, error) {
	if len(line) == 0 {
		return false, nil
	}
	var entry claudeTranscriptEntry
	if err := json.Unmarshal(line, &entry); err != nil {
		return false, err
	}
	content := bytes.TrimSpace(entry.Message.Content)
	if entry.Type != "assistant" || len(content) == 0 || content[0] != '[' {
		return false, nil
	}
	var blocks []struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(content, &blocks); err != nil {
		return false, err
	}
	for _, block := range blocks {
		if block.Type == "tool_use" {
			return true, nil
		}
	}
	return false, nil
}

func readClaudeStopMarker(path string) (int64, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is derived by agentHookMarkerPath under the bd cache dir
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read Claude Stop hook marker: %w", err)
	}
	offset, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("corrupt Claude Stop hook marker %s: %q", path, strings.TrimSpace(string(data)))
	}
	return offset, nil
}

func writeClaudeStopMarker(path string, offset int64) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("write Claude Stop hook marker: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("write Claude Stop hook marker: %w", err)
	}
	_, writeErr := tmp.WriteString(strconv.FormatInt(offset, 10) + "\n")
	closeErr := tmp.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return errors.Join(fmt.Errorf("write Claude Stop hook marker: %w", err), os.Remove(tmp.Name()))
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return errors.Join(fmt.Errorf("write Claude Stop hook marker: %w", err), os.Remove(tmp.Name()))
	}
	return nil
}

func claudeHookMarkerBaseDir() string {
	return agentHookMarkerBaseDir("claude-hooks", claudeHookMarkerDirOverride)
}
