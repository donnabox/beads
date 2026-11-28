package translate

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ExecError is a bd invocation that started, ran and exited non-zero: bd refused
// the action. A caller records a refusal, with what bd said, and carries on; it
// is a finding about the build under test, not a failure of the harness.
//
// Only a refusal is an ExecError. A bd that cannot be started, or that a signal
// killed, fails ExecuteWith with a plain error, because neither is bd saying no.
type ExecError struct {
	// Argv is the invocation, without the binary.
	Argv []string
	// ExitCode is bd's exit status, always above zero.
	ExitCode int
	// Output is what bd printed, standard output and standard error together.
	Output string
	// Err is the *exec.ExitError underneath.
	Err error
}

func (e *ExecError) Error() string {
	return fmt.Sprintf("bd %s: %v\n%s", strings.Join(e.Argv, " "), e.Err, e.Output)
}

func (e *ExecError) Unwrap() error { return e.Err }

// execFailure is the error ExecuteWith reports for the err and output a finished
// command returned: an *ExecError for a non-zero exit, a plain error otherwise.
func execFailure(argv []string, err error, output []byte) error {
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() > 0 {
		return &ExecError{Argv: argv, ExitCode: exit.ExitCode(), Output: string(output), Err: err}
	}
	return fmt.Errorf("bd %s: %w\n%s", strings.Join(argv, " "), err, output)
}
