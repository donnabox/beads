package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/steveyegge/beads/internal/config"
)

// GraphClaudeStop configures only Steph's existing Stop handler in project
// settings and links Claude to existing graph guidance with an @-import.
// It does not replace instructions or install prime hooks, plugins or globals.
func GraphClaudeStop(projectDir string, check, remove bool) error {
	env, err := claudeEnvProvider()
	if err != nil {
		return err
	}
	env.projectDir = projectDir
	return graphClaudeStop(env, check, remove)
}

func graphClaudeStop(env claudeEnv, check, remove bool) error {
	if check && remove {
		return fmt.Errorf("choose --check or --remove, not both")
	}
	path := projectSettingsPath(env.projectDir)
	// Never follow a project settings symlink into another workspace or home.
	for _, candidate := range []string{filepath.Dir(path), path} {
		info, err := os.Lstat(candidate)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && (info.Mode()&os.ModeSymlink != 0 || (candidate == path && !info.Mode().IsRegular()) || (candidate != path && !info.IsDir())) {
			return fmt.Errorf("preserve non-regular graph Claude settings path %s; configure a project-local regular file", candidate)
		}
	}
	settings, exists, err := graphClaudeSettings(env, path)
	if err != nil {
		return err
	}
	hooks := map[string]interface{}{}
	if raw, present := settings["hooks"]; present {
		var ok bool
		hooks, ok = raw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("preserve malformed hooks in %s; expected an object", path)
		}
	}
	installed, err := graphClaudeStopPresent(hooks)
	if err != nil {
		return err
	}
	if !remove {
		// Global and legacy files are read only. An existing plugin or prime
		// recipe needs a deliberate operator decision, not a silent migration.
		for _, other := range []string{globalSettingsPath(env.homeDir), legacyProjectSettingsPath(env.projectDir)} {
			if _, _, err := graphClaudeSettings(env, other); err != nil {
				return err
			}
		}
		if hasBeadsPlugin(env) {
			return fmt.Errorf("Beads Claude plugin enables unsupported graph prime hooks; disable it for this project before installing the graph Stop hook (settings were preserved)")
		}
		for _, existing := range []string{path, globalSettingsPath(env.homeDir), legacyProjectSettingsPath(env.projectDir)} {
			if hasBeadsHooks(existing) {
				return fmt.Errorf("existing Beads prime hooks in %s are unsupported in graph mode; reconcile them before setup (settings were preserved)", existing)
			}
		}
	}
	var importPath string
	var importBefore, importAfter []byte
	if !remove {
		importPath, importBefore, importAfter, err = graphClaudeImport(env)
		if err != nil {
			return err
		}
	}
	if check {
		if !bytes.Equal(importBefore, importAfter) {
			return fmt.Errorf("Claude does not import graph instructions; run bd setup claude")
		}
		if !installed {
			return fmt.Errorf("graph Claude Stop hook is not installed; run bd setup claude")
		}
		_, _ = fmt.Fprintln(env.stdout, "Graph Claude Stop hook is installed.")
		return nil
	}
	// All settings/plugin/import refusals have completed before publication.
	if !remove && !bytes.Equal(importBefore, importAfter) {
		if err := graphClaudeWritePreservingMode(importPath, importAfter); err != nil {
			return fmt.Errorf("write graph guidance import: %w", err)
		}
	}
	if remove {
		if !exists || !installed {
			_, _ = fmt.Fprintf(env.stdout, "Graph Claude Stop hook is absent from %s; instructions preserved.\n", path)
			return nil
		}
		removeHookCommand(hooks, "Stop", claudeStopHookCommand)
	} else {
		if installed {
			_, _ = fmt.Fprintf(env.stdout, "Graph Claude Stop hook configured in %s; guidance loaded by %s.\n", path, importPath)
			return nil
		}
		addHookCommand(hooks, "Stop", claudeStopHookCommand)
	}
	settings["hooks"] = hooks
	data, err := marshalSettings(settings)
	if err != nil {
		return err
	}
	if err := env.ensureDir(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if existing, readErr := env.readFile(path); readErr != nil || !bytes.Equal(existing, data) {
		if err := graphClaudeWritePreservingMode(path, data); err != nil {
			return err
		}
	}
	if remove {
		_, _ = fmt.Fprintf(env.stdout, "Graph Claude Stop hook removed from %s; instructions preserved.\n", path)
	} else {
		_, _ = fmt.Fprintf(env.stdout, "Graph Claude Stop hook configured in %s; guidance loaded by %s.\n", path, importPath)
	}
	return nil
}

// Graph setup retains existing private file modes; ordinary setup is unchanged.
func graphClaudeWritePreservingMode(path string, data []byte) error {
	mode := os.FileMode(0644)
	if info, err := os.Lstat(path); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("preserve non-regular graph Claude file %s", path)
		}
		mode = info.Mode().Perm()
	} else if !os.IsNotExist(err) {
		return err
	}
	return atomicWriteFile(path, data, mode)
}

func graphClaudeSettings(env claudeEnv, path string) (map[string]interface{}, bool, error) {
	raw, err := env.readFile(path)
	if os.IsNotExist(err) {
		return map[string]interface{}{}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	// Shared plugin/prime detection still decodes numbers as float64. Retain
	// that admission range so an unrelated overflow cannot hide those hooks;
	// UseNumber below preserves the exact values of admitted settings.
	var legacySettings interface{}
	if err := json.Unmarshal(raw, &legacySettings); err != nil {
		return nil, true, fmt.Errorf("preserve Claude settings in %s: unsupported JSON or numeric range; reconcile the settings before setup (settings were preserved): %w", path, err)
	}
	var settings map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&settings); err != nil || settings == nil {
		return nil, true, fmt.Errorf("preserve invalid Claude settings object in %s", path)
	}
	if err := decoder.Decode(new(interface{})); err != io.EOF {
		return nil, true, fmt.Errorf("preserve trailing Claude settings content in %s", path)
	}
	return settings, true, nil
}

// Validate the changed event before using the shared merge primitives, which
// intentionally tolerate malformed legacy shapes in ordinary setup.
func graphClaudeStopPresent(hooks map[string]interface{}) (bool, error) {
	raw, present := hooks["Stop"]
	if !present {
		return false, nil
	}
	entries, ok := raw.([]interface{})
	if !ok {
		return false, fmt.Errorf("preserve malformed Stop hooks; expected an array")
	}
	found := false
	for _, entry := range entries {
		object, ok := entry.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("preserve malformed Stop entry; expected an object")
		}
		commands, ok := object["hooks"].([]interface{})
		if !ok {
			return false, fmt.Errorf("preserve malformed Stop entry hooks; expected an array")
		}
		for _, command := range commands {
			value, ok := command.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("preserve malformed Stop command; expected an object")
			}
			if value["command"] == claudeStopHookCommand {
				matcher, validMatcher := object["matcher"].(string)
				_, matcherPresent := object["matcher"]
				if value["type"] != "command" || matcher != "" || (matcherPresent && !validMatcher) {
					return false, fmt.Errorf("existing bd claude-hook stop has unsupported type or matcher; settings were preserved")
				}
				found = true
			}
		}
	}
	return found, nil
}

// Reuse Claude's established active-import parser rather than copying graph
// instructions into a second managed block. No existing content is removed.
func graphClaudeImport(env claudeEnv) (string, []byte, []byte, error) {
	path := filepath.Join(env.projectDir, claudeInstructionsFile)
	info, err := os.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, nil, err
	}
	if err == nil && !info.Mode().IsRegular() {
		return "", nil, nil, fmt.Errorf("preserve non-regular %s; use a regular project CLAUDE.md before graph setup", path)
	}
	before, err := env.readFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", nil, nil, err
	}
	agentsFile := config.SafeAgentsFile()
	if agentsFile == claudeInstructionsFile {
		return path, before, before, nil // The validated graph profile is already directly loaded.
	}
	if containsBeadsMarker(string(before)) {
		return "", nil, nil, fmt.Errorf("CLAUDE.md contains a managed Beads block; reconcile it with %s before graph setup (instructions were preserved)", agentsFile)
	}
	if isAgentsImportStub(string(before), agentsFile) {
		return path, before, before, nil
	}
	after := string(before)
	if after != "" && !strings.HasSuffix(after, "\n") {
		after += "\n"
	}
	if after != "" {
		after += "\n"
	}
	after += "@" + agentsFile + "\n"
	if !isAgentsImportStub(after, agentsFile) {
		return "", nil, nil, fmt.Errorf("CLAUDE.md has an unclosed code fence; close it before graph setup (instructions were preserved)")
	}
	return path, before, []byte(after), nil
}
