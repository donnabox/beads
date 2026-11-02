package graphread

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/steveyegge/beads/internal/httpapi/bdpwire"
	"github.com/steveyegge/beads/internal/resourcequery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

func TestQueryRepresentationInputDateTimeRFCGuards(t *testing.T) {
	// First eleven cases are the exact date-time cases from the selected BDP
	// packages/protocol/src/schema-formats.test.ts. Remaining cases exercise the
	// inspected AJV calendar/clock/timezone/leap/fraction branches, with the
	// selected RFC guards rejecting malformed offsets and local clocks.
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"2026-99-99T99:99:99+99:99", false},
		{"2026-02-30T00:00:00Z", false},
		{"2026-09-07T18:04:12", false},
		{"2026-09-07 18:04:12Z", false},
		{"2026-09-07\t18:04:12Z", false},
		{"2026-09-07\n18:04:12Z", false},
		{"2026-09-07", false},
		{"not an instant", false},
		{"2026-09-07T18:04:12Z", true},
		{"2026-09-07T18:04:12.5+02:00", true},
		{"2026-09-07t18:04:12z", true},
		{"2024-02-29T00:00:00Z", true},
		{"1900-02-29T00:00:00Z", false},
		{"2000-02-29T00:00:00Z", true},
		{"0000-02-29T00:00:00Z", true},
		{"2026-04-31T00:00:00Z", false},
		{"2026-00-01T00:00:00Z", false},
		{"2026-01-00T00:00:00Z", false},
		{"2026-09-07T24:00:00Z", false},
		{"2026-09-07T12:60:00Z", false},
		{"2016-12-31T23:59:60Z", true},
		{"2016-12-31T23:59:60.5z", true},
		{"2016-12-31T23:59:61Z", false},
		{"2016-12-31T12:59:60Z", false},
		{"2017-01-01T00:59:60+01:00", true},
		{"2017-01-01T00:29:60+00:30", true},
		{"2016-12-31T18:59:60-05:00", true},
		{"2026-09-07T18:04:12+23:59", true},
		{"2026-09-07T18:04:12+24:00", false},
		{"2026-09-07T18:04:12+00:60", false},
		{"2026-09-07T18:04:12+02", false},
		{"2026-09-07T18:04:12+0200", false},
		{"2026-09-07T18:04:12-00:00", true},
		{"2026-09-07T18:04:12+2:00", false},
		{"2026-09-07T18:04:12.Z", false},
		{"2026-09-07T18:04:12.00000000000000000001Z", true},
		{"2026-09-07T18:04:12Z\n", false},
		// Diagnostic negatives: AJV accepts these malformed local clocks in its
		// leap branch. They are validator defects, not BDP conformance cases.
		{"2016-12-31T24:59:60+01:00", false},
		{"2016-12-31T23:99:60+00:40", false},
	} {
		if got := queryRepresentationInputDateTime(tc.value); got != tc.valid {
			t.Errorf("%q: got %t want %t", tc.value, got, tc.valid)
		}
	}
}

func TestQueryRepresentationRecordedTimeValidationAndPreservation(t *testing.T) {
	for _, value := range []string{"2026-09-07T18:04:12Z", "2016-12-31T23:59:60.5Z", "2026-09-07T18:04:12.5+02:00", "not an instant"} {
		bead := queryRepresentationBead()
		bead.ChangeContext = &bdpwire.ChangeContext{
			CommittedAt: bdpwire.ContextTime{State: bdpwire.ContextTimePresent, Value: &value},
			Agent:       bdpwire.ContextString{State: bdpwire.ContextAbsent},
			Message:     bdpwire.ContextMessage{State: bdpwire.ContextUndetermined},
		}
		raw, err := queryResourceJSON(bead)
		if value == "not an instant" {
			if !errors.Is(err, graphstore.ErrInvalidStore) {
				t.Fatalf("invalid recorded time: %v", err)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		var selected struct {
			RevisionProvenance struct {
				CommittedAt struct {
					Value string `json:"value"`
				} `json:"committedAt"`
			} `json:"revisionProvenance"`
		}
		if err := json.Unmarshal(raw, &selected); err != nil {
			t.Fatal(err)
		}
		if selected.RevisionProvenance.CommittedAt.Value != value {
			t.Fatalf("time normalized: %s", raw)
		}
	}
}

func TestQueryRepresentationLowercaseInputCannotBeEmittedUnchanged(t *testing.T) {
	for _, value := range []string{"2026-09-07t18:04:12Z", "2026-09-07T18:04:12z", "2026-09-07t18:04:12z"} {
		if !queryRepresentationInputDateTime(value) {
			t.Fatalf("valid lowercase input refused: %q", value)
		}
		bead := queryRepresentationBead()
		bead.ChangeContext = &bdpwire.ChangeContext{
			CommittedAt: bdpwire.ContextTime{State: bdpwire.ContextTimePresent, Value: &value},
			Agent:       bdpwire.ContextString{State: bdpwire.ContextAbsent},
			Message:     bdpwire.ContextMessage{State: bdpwire.ContextUndetermined},
		}
		raw, err := queryResourceJSON(bead)
		if !errors.Is(err, resourcequery.ErrUnsupported) || errors.Is(err, graphstore.ErrInvalidStore) || raw != nil {
			t.Fatalf("lowercase recorded context must refuse without output: raw=%s err=%v", raw, err)
		}
	}
}
