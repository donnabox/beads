package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Graph JSON uses the native batch role and the canonical graph projection in
// one transaction. This first route accepts only new Issues and local edges;
// omitted forms refuse before the store is opened rather than losing intent.
func runGraphPreviewCreateGraph(cmd *cobra.Command, args []string) error {
	if err := graphPreviewFlags(cmd, "graph", "force"); err != nil {
		return err
	}
	if len(args) != 0 {
		return graphFailure("invalid_selector", "--graph cannot be combined with a title", 2)
	}
	path, _ := cmd.Flags().GetString("graph")
	if path == "" || path == "-" {
		return graphFailure("invalid_selector", "--graph requires a JSON file path", 2)
	}
	data, err := os.ReadFile(path) // #nosec G304 -- explicit user-selected plan file.
	if err != nil {
		return graphFailure("invalid_properties", fmt.Sprintf("reading graph plan: %v", err), 2)
	}
	if unknown := detectUnknownGraphFields(data); len(unknown) != 0 {
		return graphFailure("invalid_properties", fmt.Sprintf("unknown graph plan fields: %v", unknown), 2)
	}
	var plan GraphApplyPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return graphFailure("invalid_properties", fmt.Sprintf("parsing graph plan: %v", err), 2)
	}
	if len(plan.Nodes) == 0 {
		return graphFailure("invalid_properties", "graph plan requires at least one node", 2)
	}
	for _, node := range plan.Nodes {
		// These initial values already have lossless native Issue create and
		// graph-read representations. Keep relationship, lifecycle, metadata,
		// and storage-class options behind their separate capability proofs.
		allowed := GraphApplyNode{
			Key: node.Key, ID: node.ID, Title: node.Title, Type: node.Type,
			Description: node.Description, Design: node.Design,
			AcceptanceCriteria: node.AcceptanceCriteria, Notes: node.Notes,
			SpecID: node.SpecID, ExternalRef: node.ExternalRef,
			Assignee: node.Assignee, Owner: node.Owner, Priority: node.Priority,
			EstimatedMinutes: node.EstimatedMinutes, DueAt: node.DueAt,
			Labels: node.Labels,
		}
		if node.Status == "open" {
			allowed.Status = node.Status
		}
		if !reflect.DeepEqual(node, allowed) {
			return graphFailure("capability_unavailable", fmt.Sprintf("graph node %q uses a field not yet supported by atomic canonical Issue creation", node.Key), 5)
		}
	}
	for _, edge := range plan.Edges {
		allowed := GraphApplyEdge{FromKey: edge.FromKey, FromID: edge.FromID, ToKey: edge.ToKey, ToID: edge.ToID, Type: edge.Type}
		if !reflect.DeepEqual(edge, allowed) || (edge.FromKey == "") == (edge.FromID == "") ||
			(edge.ToKey == "") == (edge.ToID == "") {
			return graphFailure("capability_unavailable", "graph edges require new keyed or exact existing local Issue endpoints; metadata is unavailable", 5)
		}
	}
	force, _ := cmd.Flags().GetBool("force")
	request, err := buildGraphApplyBatchRequest(&plan, GraphApplyOptions{Force: force}, getActorWithGit(), getOwner())
	if err != nil {
		return graphFailure("invalid_properties", err.Error(), 2)
	}
	return withGraphFamiliarStore(func(ctx context.Context, store *graphstore.Store) (any, string, error) {
		records, err := store.ApplyNewIssueGraph(ctx, request)
		if err != nil {
			return nil, "", err
		}
		result := GraphApplyResult{IDs: make(map[string]string, len(records))}
		keys := make([]string, 0, len(records))
		for key, record := range records {
			result.IDs[key] = record.Properties.ID
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var human strings.Builder
		fmt.Fprintf(&human, "Created %d issues", len(keys))
		for _, key := range keys {
			fmt.Fprintf(&human, "\n  %s -> %s", key, result.IDs[key])
		}
		return result, human.String(), nil
	})
}
