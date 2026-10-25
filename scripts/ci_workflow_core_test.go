package scripts_test

// ciWorkflow and its supporting types/helpers parse a .github/workflows/*.yml
// file into a form the policy tests can query. Helper-only file (no Test
// functions, like repo_helpers_test.go): shared by scripts_test's
// ci_workflow_test.go and nogo_integration_scan_test.go's dedicated go_test
// target (independent review of #7486, cost fix), so each can declare only
// the data it reads without duplicating this parser.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type ciWorkflow struct {
	Env  map[string]string        `yaml:"env"`
	Jobs map[string]ciWorkflowJob `yaml:"jobs"`
}

type ciWorkflowJob struct {
	Name            string                `yaml:"name"`
	Uses            string                `yaml:"uses"`
	With            map[string]string     `yaml:"with"`
	Secrets         any                   `yaml:"secrets"`
	Permissions     any                   `yaml:"permissions"`
	Needs           ciWorkflowStringList  `yaml:"needs"`
	Steps           []ciWorkflowStep      `yaml:"steps"`
	RunsOn          string                `yaml:"runs-on"`
	If              string                `yaml:"if"`
	ContinueOnError bool                  `yaml:"continue-on-error"`
	TimeoutMinutes  int                   `yaml:"timeout-minutes"`
	Strategy        ciWorkflowStrategy    `yaml:"strategy"`
	Env             map[string]string     `yaml:"env"`
	Outputs         map[string]string     `yaml:"outputs"`
	Environment     ciWorkflowEnvironment `yaml:"environment"`
}

type ciWorkflowStrategy struct {
	FailFast bool             `yaml:"fail-fast"`
	Matrix   ciWorkflowMatrix `yaml:"matrix"`
}

type ciWorkflowMatrix struct {
	OS     []string `yaml:"os"`
	Runner []string `yaml:"runner"`
	Shard  []int    `yaml:"shard"`
	Target []string `yaml:"target"`
	Check  []string `yaml:"check"`
	Group  []string `yaml:"group"`
	// Venue and Flavor back main.yml's "venue matrix" (F7b spec §4) push-to-
	// main-only Blacksmith cache savers: test-windows's
	// `venue: [blacksmith, github]` and
	// blacksmith-go-build-cache's `flavor: [race, non-race]`. Pinned by
	// TestBlacksmithSaverVenueAndFlavorMatricesAreComplete (F7b review B2/S7
	// follow-up) so dropping either leg from these lists -- silently losing
	// that leg's cache seed or (for flavor) its race coverage -- fails loudly.
	Venue   []string                  `yaml:"venue"`
	Flavor  []string                  `yaml:"flavor"`
	Include []ciWorkflowMatrixInclude `yaml:"include"`
}

type ciWorkflowMatrixInclude struct {
	OS           string `yaml:"os"`
	ExpectedGOOS string `yaml:"expected_goos"`
	Coverage     bool   `yaml:"coverage"`
	TestFlags    string `yaml:"test-flags"`
	Runner       string `yaml:"runner"`
	// F7b review fix (S1): no typed `shard` field here. pr-preflight-
	// platforms' Windows shard split (the one thing in this file that ever
	// used it) was reverted as not measurably helpful, and a typed field
	// here is shared by every workflow's `strategy.matrix.include` tuples,
	// whose keys may be any type, so a mismatched type here would break
	// parsing of an unrelated workflow. Untyped matrix keys fall through to
	// Extra.
	Extra map[string]any `yaml:",inline"`
}

type ciWorkflowStep struct {
	Name            string            `yaml:"name"`
	ID              string            `yaml:"id"`
	If              string            `yaml:"if"`
	Uses            string            `yaml:"uses"`
	Run             string            `yaml:"run"`
	Shell           string            `yaml:"shell"`
	ContinueOnError any               `yaml:"continue-on-error"`
	TimeoutMinutes  int               `yaml:"timeout-minutes"`
	Env             map[string]string `yaml:"env"`
	With            map[string]string `yaml:"with"`
}

type ciWorkflowStringList []string

func (items *ciWorkflowStringList) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		*items = []string{node.Value}
		return nil
	case yaml.SequenceNode:
		values := make([]string, 0, len(node.Content))
		for _, item := range node.Content {
			if item.Kind != yaml.ScalarNode {
				return fmt.Errorf("needs item must be scalar, got YAML kind %d", item.Kind)
			}
			values = append(values, item.Value)
		}
		*items = values
		return nil
	default:
		return fmt.Errorf("needs must be scalar or sequence, got YAML kind %d", node.Kind)
	}
}

func readCIWorkflow(t *testing.T, name string) ciWorkflow {
	t.Helper()

	path := filepath.Join(sourceRepoRoot(t), ".github", "workflows", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	workflow, err := decodeCIWorkflowForRBE(name, data, false)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return workflow
}

// decodeCIWorkflowForRBE keeps timeout fields typed as integers for all
// policy tests. Only bazel-test's two reviewed mode-dependent expressions may
// be resolved here; this is not a general GitHub expression evaluator.
func decodeCIWorkflowForRBE(name string, data []byte, enabled bool) (ciWorkflow, error) {
	var workflow ciWorkflow
	if err := decodeCIWorkflowYAMLForRBE(name, data, enabled, &workflow); err != nil {
		return ciWorkflow{}, err
	}
	return workflow, nil
}

// decodeCIWorkflowYAMLForRBE preserves each policy test's destination schema
// while applying the same narrow timeout normalization before typed decoding.
func decodeCIWorkflowYAMLForRBE(name string, data []byte, enabled bool, destination any) error {
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return err
	}
	if name == "bazel.yml" {
		if err := resolveBazelTestTimeouts(&document, enabled); err != nil {
			return err
		}
	}
	return document.Decode(destination)
}

func ciWorkflowMapValue(mapping *yaml.Node, key string) (*yaml.Node, error) {
	if mapping.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping at %q, got YAML kind %d", key, mapping.Kind)
	}
	var found *yaml.Node
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value != key {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("duplicate %q key", key)
		}
		found = mapping.Content[i+1]
	}
	if found == nil {
		return nil, fmt.Errorf("missing %q key", key)
	}
	return found, nil
}

func resolveBazelTestTimeouts(document *yaml.Node, enabled bool) error {
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
		return fmt.Errorf("bazel workflow must contain one YAML document")
	}
	jobs, err := ciWorkflowMapValue(document.Content[0], "jobs")
	if err != nil {
		return err
	}
	job, err := ciWorkflowMapValue(jobs, "bazel-test")
	if err != nil {
		return err
	}
	jobTimeout, err := ciWorkflowMapValue(job, "timeout-minutes")
	if err != nil {
		return err
	}
	steps, err := ciWorkflowMapValue(job, "steps")
	if err != nil {
		return err
	}
	if steps.Kind != yaml.SequenceNode {
		return fmt.Errorf("bazel-test steps must be a sequence")
	}
	var testStep *yaml.Node
	for _, step := range steps.Content {
		if step.Kind != yaml.MappingNode {
			return fmt.Errorf("bazel-test step must be a mapping")
		}
		for i := 0; i < len(step.Content); i += 2 {
			if step.Content[i].Value == "id" && step.Content[i+1].Value == "test" {
				if testStep != nil {
					return fmt.Errorf("duplicate bazel-test id:test step")
				}
				testStep = step
			}
		}
	}
	if testStep == nil {
		return fmt.Errorf("missing bazel-test id:test step")
	}
	stepName, err := ciWorkflowMapValue(testStep, "name")
	if err != nil {
		return err
	}
	if stepName.Value != "bazel test //... --config=ci" {
		return fmt.Errorf("bazel-test id:test step has unexpected name %q", stepName.Value)
	}
	stepTimeout, err := ciWorkflowMapValue(testStep, "timeout-minutes")
	if err != nil {
		return err
	}
	const jobExpression = "${{ needs.rbe.outputs.enabled == 'true' && 115 || 130 }}"
	const stepExpression = "${{ needs.rbe.outputs.enabled == 'true' && 110 || 125 }}"
	if jobTimeout.Kind != yaml.ScalarNode || jobTimeout.Value != jobExpression {
		return fmt.Errorf("bazel-test job timeout must be exact approved expression, got %q", jobTimeout.Value)
	}
	if stepTimeout.Kind != yaml.ScalarNode || stepTimeout.Value != stepExpression {
		return fmt.Errorf("bazel-test id:test timeout must be exact approved expression, got %q", stepTimeout.Value)
	}
	var rejectOtherTimeoutExpressions func(*yaml.Node) error
	rejectOtherTimeoutExpressions = func(node *yaml.Node) error {
		if node.Kind == yaml.MappingNode {
			for i := 0; i < len(node.Content); i += 2 {
				key, value := node.Content[i], node.Content[i+1]
				if key.Value == "timeout-minutes" && value != jobTimeout && value != stepTimeout &&
					(value.Kind == yaml.AliasNode || strings.Contains(value.Value, "${{")) {
					return fmt.Errorf("unexpected timeout expression or alias at line %d", value.Line)
				}
			}
		}
		for _, child := range node.Content {
			if err := rejectOtherTimeoutExpressions(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := rejectOtherTimeoutExpressions(document); err != nil {
		return err
	}
	jobMinutes, stepMinutes := 130, 125
	if enabled {
		jobMinutes, stepMinutes = 115, 110
	}
	for node, minutes := range map[*yaml.Node]int{jobTimeout: jobMinutes, stepTimeout: stepMinutes} {
		node.Value = strconv.Itoa(minutes)
		node.Tag = "!!int"
		node.Style = 0
	}
	return nil
}

func (workflow ciWorkflow) job(t *testing.T, name string) ciWorkflowJob {
	t.Helper()

	job, ok := workflow.Jobs[name]
	if !ok {
		t.Fatalf("workflow has no %q job", name)
	}
	return job
}

func (job ciWorkflowJob) step(t *testing.T, name string) ciWorkflowStep {
	t.Helper()

	for _, step := range job.Steps {
		if step.Name == name {
			return step
		}
	}
	t.Fatalf("job has no %q step", name)
	return ciWorkflowStep{}
}

func (job ciWorkflowJob) stepIndex(t *testing.T, name string) int {
	t.Helper()

	index := -1
	for i, step := range job.Steps {
		if step.Name != name {
			continue
		}
		if index >= 0 {
			t.Fatalf("job has more than one %q step", name)
		}
		index = i
	}
	if index < 0 {
		t.Fatalf("job has no %q step", name)
	}
	return index
}

func assertJobRunsExactly(t *testing.T, job ciWorkflowJob, want string) {
	t.Helper()

	for _, step := range job.Steps {
		if strings.TrimSpace(step.Run) == want {
			return
		}
	}
	t.Errorf("job has no step that runs exactly %q", want)
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// --- Bazel lane (.github/workflows/bazel.yml, gated through pr.yml) ----------

const (
	bazelWorkflowName = "bazel.yml"
	bazelJobName      = "bazel-test"
	bazelPureJobName  = "bazel-pure"
	// The release-target cross-compilation, split out of bazel-pure so it
	// runs in parallel with it (same runner and skip rule).
	bazelReleaseCrossJobName = "bazel-release-cross"
	bazelDoltJobName         = "bazel-doltserver"
	bazelEmbedJobName        = "bazel-embedded"
	bazelRBEJobName          = "rbe"
	bazelIntegJobName        = "bazel-integration"
	bazelProxiedJobName      = "bazel-proxied"
	bazelServerJobName       = "bazel-server-storage"
	bazelCmdDoltJobName      = "bazel-cmd-dolt"
	// F3: the MCP and npm package gates, moved here from pr.yml so they need
	// only the rbe job. Unlike every other lane they do not mirror a Bazel
	// config; pr.yml opts them in with package-gates: "on".
	bazelPackageMCPJobName = "package-mcp"
	bazelPackageNPMJobName = "package-npm"
	// Best-effort, advisory pre-warm of gastownhall/gascity's rbe-west OSS
	// worker pool; see bazel.yml's own comment and
	// engdocs/CI_REQUIRED_CHECK_TOPOLOGY.md, "rbe-west Pre-warm".
	bazelRBEPrewarmJobName = "rbe-prewarm"
	setupBazelActionDir    = ".github/actions/setup-bazel"
)

// ciWorkflowEnvironment is a job's environment, written either as a bare name
// (environment: autofix) or as a mapping (environment: {name: github-pages,
// url: ...}, deploy-pages-redirect.yml).
type ciWorkflowEnvironment struct {
	Name string
	URL  string
}

func (e *ciWorkflowEnvironment) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		e.Name = value.Value
		return nil
	}
	var m struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	}
	if err := value.Decode(&m); err != nil {
		return err
	}
	e.Name, e.URL = m.Name, m.URL
	return nil
}
