//go:build cgo

package graphstore

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/steveyegge/beads/internal/types"
)

func TestIssueLabelsConcurrentPriorityWriter(t *testing.T) {
	for _, backend := range []string{"embedded", "server"} {
		t.Run(backend, func(t *testing.T) {
			base, options, first, source, target, dependency := reopenFixture(t, backend)
			ctx, cancel := context.WithTimeout(base, time.Minute)
			defer cancel()
			type auditRow struct {
				eventType types.EventType
				actor     string
			}
			readAudit := func() map[string]auditRow {
				t.Helper()
				rows, err := first.db.QueryContext(ctx, "SELECT id, event_type, actor FROM events WHERE issue_id = ?", source.Properties.ID)
				if err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := rows.Close(); err != nil {
						t.Error(err)
					}
				}()
				out := make(map[string]auditRow)
				for rows.Next() {
					var id string
					var row auditRow
					if err := rows.Scan(&id, &row.eventType, &row.actor); err != nil {
						t.Fatal(err)
					}
					out[id] = row
				}
				if err := rows.Err(); err != nil {
					t.Fatal(err)
				}
				return out
			}
			// Preserve the fixture's create/Dependency events; assess only IDs
			// added by these two transactions, never a timestamp cutoff.
			beforeAudit := readAudit()
			// Keep embedded's production single-session pool. Ordinary server uses
			// independent sessions and must reach the barrier in both transactions.
			second := &Store{db: first.db, options: options}
			if backend == "server" {
				var err error
				second, err = OpenExisting(ctx, options)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if err := second.Close(); err != nil {
						t.Error(err)
					}
				})
			}
			reached, release := make(chan struct{}, 2), make(chan struct{})
			pause := func(stage string) error {
				if backend != "server" || stage != "source-retained" {
					return nil
				}
				reached <- struct{}{}
				select {
				case <-release:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			first.afterWrite, second.afterWrite = pause, pause
			type outcome struct {
				kind  string
				value IssueMutationResult
				err   error
			}
			results := make(chan outcome, 2)
			var writers sync.WaitGroup
			writers.Add(2)
			defer func() { cancel(); writers.Wait(); first.afterWrite, second.afterWrite = nil, nil }()
			labels := []string{"Alpha", "alpha", "café"}
			go func() {
				defer writers.Done()
				value, err := first.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "labels-writer", ExpectedRevision: source.Revision, Labels: &labels})
				results <- outcome{kind: "labels", value: value, err: err}
			}()
			go func() {
				defer writers.Done()
				value, err := second.UpdateIssue(ctx, UpdateIssueRequest{Path: "beads/work", Actor: "priority-writer", ExpectedRevision: source.Revision, Priority: issuePriority(0)})
				results <- outcome{kind: "priority", value: value, err: err}
			}()
			if backend == "server" {
				for range 2 {
					select {
					case <-reached:
					case early := <-results:
						t.Fatalf("writer exited before forced overlap: %+v", early)
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					}
				}
				close(release)
			}
			var winner outcome
			conflicts := 0
			for range 2 {
				select {
				case result := <-results:
					if result.err == nil {
						if winner.kind != "" || !result.value.Changed {
							t.Fatalf("expected one changed winner: %+v", result)
						}
						winner = result
					} else {
						if !errors.Is(result.err, ErrConflict) || errors.Is(result.err, ErrOutcomeUnknown) || !reflect.DeepEqual(result.value, IssueMutationResult{}) {
							t.Fatalf("non-atomic conflict: %+v", result)
						}
						conflicts++
					}
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			writers.Wait()
			first.afterWrite, second.afterWrite = nil, nil
			if winner.kind == "" || conflicts != 1 {
				t.Fatalf("winner=%s conflicts=%d", winner.kind, conflicts)
			}
			current, err := first.ShowIssue(ctx, "beads/work")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(current, winner.value.Issue) || current.Revision == source.Revision || current.Attribution.Actor != winner.kind+"-writer" {
				t.Fatalf("current state is not complete accepted result: %+v", current)
			}
			wantedLabels, wantedPriority := source.Properties.Labels, source.Properties.Priority
			if winner.kind == "labels" {
				wantedLabels = labels
			} else {
				wantedPriority = 0
			}
			assertIssueLabels(t, current.Properties.Labels, wantedLabels)
			if current.Properties.Priority != wantedPriority {
				t.Fatal("losing priority mutation leaked")
			}
			properties := *current.Properties
			properties.Labels = source.Properties.Labels
			properties.Priority = source.Properties.Priority
			properties.UpdatedAt = source.Properties.UpdatedAt
			properties.ContentHash = source.Properties.ContentHash
			properties.RowVersion = source.Properties.RowVersion
			if !reflect.DeepEqual(properties, *source.Properties) || !reflect.DeepEqual(current.Owned, source.Owned) {
				t.Fatal("unrelated properties or complete owned state changed")
			}
			if got, err := first.ShowIssue(ctx, "beads/prereq"); err != nil || !reflect.DeepEqual(got, target) {
				t.Fatalf("target changed: %+v %v", got, err)
			}
			if got, err := first.ShowLink(ctx, "links/block"); err != nil || !reflect.DeepEqual(got, dependency) {
				t.Fatalf("owned Dependency changed: %+v %v", got, err)
			}
			afterAudit := readAudit()
			for id, before := range beforeAudit {
				if after, ok := afterAudit[id]; !ok || after != before {
					t.Fatalf("prior audit row changed or disappeared: %s", id)
				}
			}
			newEvents := make(map[types.EventType]int)
			for id, row := range afterAudit {
				if _, existed := beforeAudit[id]; existed {
					continue
				}
				if row.actor != winner.kind+"-writer" {
					t.Fatalf("losing or unexpected writer leaked audit row: %s %+v", id, row)
				}
				newEvents[row.eventType]++
			}
			wantEvents := map[types.EventType]int{types.EventUpdated: 1}
			if winner.kind == "labels" {
				wantEvents = make(map[types.EventType]int)
				oldSet, newSet := make(map[string]bool), make(map[string]bool)
				for _, label := range source.Properties.Labels {
					oldSet[label] = true
				}
				for _, label := range labels {
					newSet[label] = true
				}
				for label := range newSet {
					if !oldSet[label] {
						wantEvents[types.EventLabelAdded]++
					}
				}
				for label := range oldSet {
					if !newSet[label] {
						wantEvents[types.EventLabelRemoved]++
					}
				}
			}
			if !reflect.DeepEqual(newEvents, wantEvents) {
				t.Fatalf("winning %s audit events = %v, want %v", winner.kind, newEvents, wantEvents)
			}
			assertIssueEditCounts(t, ctx, first, source.Properties.ID, 3)
			assertIssueListRetained(t, ctx, first, "beads/work", source)
			assertIssueListRetained(t, ctx, first, "beads/work", current)
		})
	}
}
