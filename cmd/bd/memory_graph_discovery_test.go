package main

import (
	"context"
	"errors"
	"github.com/steveyegge/beads/internal/memoryquery"
	"github.com/steveyegge/beads/internal/storage/graphstore"
	"testing"
)

func TestMemoryDiscoveryValidationIndependentOfOutput(t *testing.T) {
	for _, operation := range []string{"constraints", "subjects"} {
		t.Run(operation, func(t *testing.T) {
			for _, malformed := range []bool{false, true} {
				var read graphstore.CurrentFamilyRead
				var staged *graphstore.CurrentConstraintsRead
				if operation == "constraints" {
					r := constraintPresentationRead(t, []memoryquery.ConstraintRoot{{Path: "beads/a", RequiresDepth: 3}}, [4]string{"sab", "subject-to", "a", "b"}, [4]string{"bc", "requires", "b", "c"})
					if malformed {
						r.Stages[0].Requires.Seeds = nil
					}
					read, staged = r.CurrentFamilyRead, &r
				} else {
					reader := constraintPresentationFixture(t, [4]string{"ab", "subject-to", "a", "b"}, [4]string{"cb", "subject-to", "c", "b"}, [4]string{"ac", "subject-to", "a", "c"})
					g, err := memoryquery.CurrentSubjects(context.Background(), reader, constraintPresentationFamily, []string{"beads/b"}, memoryquery.Limits{MaxNodes: 30, MaxLinks: 30, MaxInspected: 10000})
					if err != nil {
						t.Fatal(err)
					}
					if malformed {
						g.Links = append(g.Links, reader.links[2])
						g.Views[0].Links = append(g.Views[0].Links, "links/ac")
					}
					read = constraintPresentationRecords(g)
				}
				// This validator has no format or quiet argument: the command calls it
				// before choosing tree/detail/JSON or suppressing human output.
				err := validateMemoryDiscoveryCapture(read, staged, genericTestScope, operation)
				if malformed && !errors.Is(err, graphstore.ErrInvalidStore) {
					t.Fatalf("accepted malformed %s capture: %v", operation, err)
				}
				if !malformed && err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestMemorySubjectsDiscoveryRejectsFrontier(t *testing.T) {
	reader := constraintPresentationFixture(t, [4]string{"ab", "subject-to", "a", "b"})
	g, err := memoryquery.CurrentSubjects(context.Background(), reader, constraintPresentationFamily, []string{"beads/b"}, memoryquery.Limits{MaxNodes: 30, MaxLinks: 30, MaxInspected: 10000})
	if err != nil {
		t.Fatal(err)
	}
	g.Views[0].Frontier = []string{"beads/a"}
	if err := validateMemoryDiscoveryCapture(constraintPresentationRecords(g), nil, genericTestScope, "subjects"); !errors.Is(err, graphstore.ErrInvalidStore) {
		t.Fatalf("accepted direct frontier: %v", err)
	}
}
