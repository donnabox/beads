package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/beads/internal/storage/graphstore"
)

// Selection metadata is separate from complete canonical records. This is a
// CLI result, not a new BDP record or historical graph snapshot contract.
type graphSelectedOutput struct {
	Scope     string                  `json:"scope"`
	Beads     []any                   `json:"beads"`
	Links     []graphstore.LinkRecord `json:"links"`
	Traversal graphSelectedTraversal  `json:"traversal"`
}
type graphSelectedTraversal struct {
	Root         string   `json:"root,omitempty"`
	All          bool     `json:"all,omitempty"`
	Direction    string   `json:"direction"`
	Depth        int      `json:"depth"`
	MaxNodes     int      `json:"maxNodes"`
	MaxLinks     int      `json:"maxLinks"`
	LinkTypes    []string `json:"linkTypes"`
	AllLinkTypes bool     `json:"allLinkTypes,omitempty"`
	Complete     bool     `json:"complete"`
	Frontier     []string `json:"frontier"`
}

func graphSelectedRecords(result graphGenericResult, snapshot graphstore.Snapshot) (graphSelectedOutput, error) {
	out := graphSelectedOutput{Scope: result.Scope, Beads: []any{}, Links: []graphstore.LinkRecord{}, Traversal: graphSelectedTraversal{Root: result.Root, All: result.All, Direction: result.Direction, Depth: result.Depth, MaxNodes: result.MaxNodes, MaxLinks: result.MaxLinks, LinkTypes: result.LinkTypes, AllLinkTypes: result.AllLinkTypes, Complete: result.Complete, Frontier: result.Frontier}}
	records := map[string]any{}
	for _, raw := range snapshot.Records {
		switch r := raw.(type) {
		case graphstore.Record:
			records[r.ID] = r
		case graphstore.IssueRecord:
			records[r.ID] = r
		case graphstore.LinkRecord:
			records[r.ID] = r
		}
	}
	for _, n := range result.Nodes {
		record, ok := records[n.ID]
		if !ok {
			return graphSelectedOutput{}, fmt.Errorf("%w: selected Bead is absent", graphstore.ErrInvalidStore)
		}
		switch record.(type) {
		case graphstore.Record, graphstore.IssueRecord:
		default:
			return graphSelectedOutput{}, fmt.Errorf("%w: selected Bead kind mismatch", graphstore.ErrInvalidStore)
		}
		out.Beads = append(out.Beads, record)
	}
	for _, l := range result.Links {
		record, ok := records[l.ID].(graphstore.LinkRecord)
		if !ok {
			return graphSelectedOutput{}, fmt.Errorf("%w: selected Link is absent or has wrong kind", graphstore.ErrInvalidStore)
		}
		out.Links = append(out.Links, record)
	}
	return out, nil
}

func graphSelectedOpenIssues(snapshot graphstore.Snapshot) graphstore.Snapshot {
	selected := map[string]bool{}
	out := snapshot
	out.Records = []any{}
	for _, raw := range snapshot.Records {
		if r, ok := raw.(graphstore.IssueRecord); ok && r.Properties != nil {
			switch r.Properties.Status {
			case "open", "in_progress", "blocked":
				selected[r.ID] = true
				out.Records = append(out.Records, r)
			}
		}
	}
	for _, raw := range snapshot.Records {
		if r, ok := raw.(graphstore.LinkRecord); ok && selected[r.Source] && selected[r.Target] {
			out.Records = append(out.Records, r)
		}
	}
	return out
}

// Reuse the command writer and JSON record projection. Native dependency
// renderers are left untouched: their readiness, priority and reversed blocker
// arrows do not describe a generic Memory Link.
func renderGraphSelected(cmd *cobra.Command, result graphGenericResult, snapshot graphstore.Snapshot) (string, error) {
	var output bytes.Buffer
	if jsonOutput {
		records, err := graphSelectedRecords(result, snapshot)
		if err != nil {
			return "", err
		}
		if err := graphPrintTo(&output, records, "", quietFlag, true); err != nil {
			return "", err
		}
	} else if !quietFlag {
		w := &graphExportWriter{out: &output}
		dot, _ := cmd.Flags().GetBool("dot")
		web, _ := cmd.Flags().GetBool("html")
		box, _ := cmd.Flags().GetBool("box")
		switch {
		case dot:
			renderGraphSelectedDOT(w, result)
		case web:
			if err := renderGraphSelectedHTML(w, result); err != nil {
				return "", err
			}
		default:
			if box {
				renderGraphSelectedText(w, result, true, false)
			} else if err := renderGraphRelationshipTree(w, result); err != nil {
				return "", err
			}
		}
		if err := w.wrapError("graph"); err != nil {
			return "", err
		}
	}
	if output.Len() > graphstore.PreviewCurrentReadByteLimit {
		return "", fmt.Errorf("%w: graph output exceeds %d bytes", graphstore.ErrLimitExceeded, graphstore.PreviewCurrentReadByteLimit)
	}
	return output.String(), nil
}
func renderGraphSelectedText(w *graphExportWriter, r graphGenericResult, box, compact bool) {
	local := func(value string) string { return strings.TrimPrefix(value, r.Scope) }
	w.printf("Graph (%d Beads, %d Links; direction %s; depth %d)\n", len(r.Nodes), len(r.Links), r.Direction, r.Depth)
	// A hop-selected set is not an execution schedule. Never label a Memory ready.
	for _, n := range r.Nodes {
		switch {
		case box:
			w.printf("┌ %s [%s] %s\n│ %q\n└\n", local(n.ID), local(n.Type), n.Status, n.Title)
		case compact:
			w.printf("• %s [%s] %s %q\n", local(n.ID), local(n.Type), n.Status, n.Title)
		default:
			w.printf("%s [%s] %s %q\n", local(n.ID), local(n.Type), n.Status, n.Title)
		}
	}
	for _, l := range r.Links {
		w.printf("%s --%s (%s)--> %s", local(l.Source), local(l.Type), local(l.ID), local(l.Target))
		if l.SourceRevision != "" || l.TargetRevision != "" {
			w.printf(" [source revision=%q target revision=%q]", l.SourceRevision, l.TargetRevision)
		}
		w.println()
	}
	w.printf("Complete: %t; frontier: %d\n", r.Complete, len(r.Frontier))
	for _, id := range r.Frontier {
		w.printf("Frontier %s\n", id)
	}
}
func renderGraphSelectedDOT(w *graphExportWriter, r graphGenericResult) {
	w.println("digraph beads {")
	w.println("  rankdir=LR;")
	for _, n := range r.Nodes {
		w.printf("  %q [label=%q, shape=box];\n", n.ID, n.Title+"\n"+n.Type)
	}
	for _, l := range r.Links {
		w.printf("  %q -> %q [id=%q, label=%q, tooltip=%q];\n", l.Source, l.Target, l.ID, l.Type, l.ID+" source revision="+l.SourceRevision+" target revision="+l.TargetRevision)
	}
	w.printf("  graph [label=%q];\n", fmt.Sprintf("direction %s; depth %d; complete %t; frontier %d", r.Direction, r.Depth, r.Complete, len(r.Frontier)))
	w.println("}")
}
func renderGraphSelectedHTML(w *graphExportWriter, r graphGenericResult) error {
	// Existing HTML/D3 template, with neutral Type labels rather than synthetic
	// Issue status/priority. The Link array retains parallel identities and pins.
	type neutralNode struct {
		ID    string `json:"id"`
		Label string `json:"label"`
		Title string `json:"title"`
		Type  string `json:"type"`
		Layer int    `json:"layer"`
	}
	nodes := make([]neutralNode, 0, len(r.Nodes))
	for _, n := range r.Nodes {
		nodes = append(nodes, neutralNode{ID: n.ID, Label: strings.TrimPrefix(n.ID, r.Scope), Title: n.Title, Type: n.Type})
	}
	rawNodes, err := json.Marshal(nodes)
	if err != nil {
		return err
	}
	rawLinks, err := json.Marshal(r.Links)
	if err != nil {
		return err
	}
	template := strings.NewReplacer(
		`.text(d => d.id + " P" + d.priority)`, `.text(d => d.label)`,
		`tooltip.append("span").text("Priority: P" + d.priority + " | Type: " + d.type);`, `tooltip.append("span").text("Type: " + d.type);`,
		`tooltip.append("span").text("Layer: " + d.layer);`, `tooltip.append("span").text("Current Bead record; inspect Link pins separately");`,
		`tooltip.append("span").attr("class","tt-status").style("background", statusColors[d.status]||"#555").text(d.status);`, ``,
		`.attr("refX",20)`, `.attr("refX",10)`,
		`.node text { font-size:`, `.node text { fill: #eee; font-size:`,
		`.node .id-text { font-size: 9px; fill: #999; }`, `.node .id-text { font-size: 9px; fill: #ccc; }`,
		`.link { fill: none;`, `.link { stroke: #888; fill: none;`,
		`selectAll("line").data(links).join("line")`, `selectAll("path").data(links).join("path")`,
		`link.attr("x1",d=>d.source.x).attr("y1",d=>d.source.y).attr("x2",d=>d.target.x).attr("y2",d=>d.target.y);`, `link.attr("d", graphLinkPath);`,
	).Replace(htmlTemplate)
	// The native legend describes blocking Issue workflows. Replace only this
	// private template copy; no dependency-mode HTML bytes change.
	legendStart := strings.Index(template, `<div id="legend">`)
	legendEnd := strings.Index(template[legendStart:], `<div id="controls">`) + legendStart
	template = template[:legendStart] + `<div id="legend"><h3>Typed Bead graph</h3><div>Arrows follow authored source → target.</div><div>Hover a Link for its identity, Type and pins.</div></div>` + template[legendEnd:]
	// Give parallel and reverse Links distinct arcs, retaining the existing D3
	// zoom, drag and layout. Stable input order makes offsets deterministic.
	const linkGeometry = `
const graphPairs = new Map();
for (const edge of links) {
  const key = JSON.stringify([edge.source, edge.target].sort());
  const group = graphPairs.get(key) || [];
  edge.graphArc = group.length + 1;
  group.push(edge); graphPairs.set(key, group);
}
function graphLinkPath(d) {
  const sx=d.source.x, sy=d.source.y, tx=d.target.x, ty=d.target.y;
  if (d.source.id === d.target.id) {
    const radius=35+d.graphArc*12;
    return "M"+(sx-25)+","+(sy-20)+" C"+(sx-radius)+","+(sy-radius)+" "+(sx+radius)+","+(sy-radius)+" "+(sx+25)+","+(sy-20);
  }
  const dx=tx-sx, dy=ty-sy, length=Math.sqrt(dx*dx+dy*dy) || 1;
  const orientation=d.source.id<d.target.id ? 1 : -1;
  const offset=d.graphArc*18*orientation;
  const cx=(sx+tx)/2-dy/length*offset, cy=(sy+ty)/2+dx/length*offset;
  function boundary(x,y,vx,vy) {
    const scale=Math.min(65/(Math.abs(vx)||1e-9),20/(Math.abs(vy)||1e-9));
    return [x+vx*scale,y+vy*scale];
  }
  const start=boundary(sx,sy,cx-sx,cy-sy), end=boundary(tx,ty,cx-tx,cy-ty);
  return "M"+start[0]+","+start[1]+" Q"+cx+","+cy+" "+end[0]+","+end[1];
}
`
	// D3 forceLink resolves endpoint strings to objects at construction. Assign
	// arc groups before constructing that force so grouping uses authored IDs.
	template = strings.Replace(template, `const simulation = d3.forceSimulation(nodes)`, linkGeometry+`
const simulation = d3.forceSimulation(nodes)`, 1)
	const linkTitle = `link.append("title").text(d => d.id + "\n" + d.type + "\nsource revision: " + (d.sourceRevision || "unpinned") + "\ntarget revision: " + (d.targetRevision || "unpinned"));
const node = g.append("g")`
	template = strings.Replace(template, `const node = g.append("g")`, linkTitle, 1)
	title := fmt.Sprintf("Beads graph — complete %t; depth %d; frontier %d", r.Complete, r.Depth, len(r.Frontier))
	w.printf(template, html.EscapeString(title), string(rawNodes), string(rawLinks))
	return nil
}
