//go:build cgo

package main

import (
    "context"
    "encoding/json"
    "errors"
    "path/filepath"
    "reflect"
    "testing"
    "time"

    "github.com/spf13/cobra"
    "github.com/steveyegge/beads/internal/clirevision"
    "github.com/steveyegge/beads/internal/configfile"
    "github.com/steveyegge/beads/internal/httpapi/graphread"
    "github.com/steveyegge/beads/internal/storage/graphstore"
    "github.com/steveyegge/beads/internal/types"
    publicops "github.com/steveyegge/beads/issueops"
)

func TestCanonicalBeadDeleteIntentBeforeAcquisition(t *testing.T) {
    for _, tc := range []struct { name string; args []string; force bool; guard string; guardSet bool; valid bool }{
        {name:"observational preview", args:[]string{"task"}, valid:true},
        {name:"unconditional apply", args:[]string{"task"}, force:true, valid:true},
        {name:"exact guard", args:[]string{"task"}, force:true, guard:"=42", guardSet:true, valid:true},
        {name:"ordinal guard", args:[]string{"task"}, force:true, guard:"2", guardSet:true, valid:true},
        {name:"missing target"},
        {name:"multiple targets", args:[]string{"one","two"}, force:true},
        {name:"pinned write target", args:[]string{"task@=r1"}, force:true},
        {name:"guarded preview", args:[]string{"task"}, guard:"=r1", guardSet:true},
        {name:"current guard", args:[]string{"task"}, force:true, guard:"current", guardSet:true},
        {name:"bare token guard", args:[]string{"task"}, force:true, guard:"token", guardSet:true},
        {name:"empty guard", args:[]string{"task"}, force:true, guardSet:true},
    } {
        t.Run(tc.name, func(t *testing.T) {
            cmd := &cobra.Command{Use:"delete"}
            cmd.Flags().Bool("force",false,"")
            cmd.Flags().String("if-revision","","")
            if tc.force { if err := cmd.Flags().Set("force","true"); err != nil { t.Fatal(err) } }
            if tc.guardSet { if err := cmd.Flags().Set("if-revision",tc.guard); err != nil { t.Fatal(err) } }
            got, err := prepareCanonicalBeadDelete(cmd,tc.args)
            if tc.valid {
                if err != nil || got.Identity != tc.args[0] || got.Apply != tc.force || got.Guarded != tc.guardSet { t.Fatalf("intent lost: %+v %v",got,err) }
                if tc.guard == "=42" && (got.Guard.Token != "42" || got.Guard.Ordinal != 0) { t.Fatal("numeric exact token became an ordinal") }
                if tc.guard == "2" && (got.Guard.Ordinal != 2 || got.Guard.Token != "") { t.Fatal("ordinal was not preserved for checked resolution") }
            } else {
                var refusal *canonicalMutationLeafRefusal
                if !errors.As(err,&refusal) || refusal.ExitCode != 2 { t.Fatalf("invalid intent admitted: %+v %v",got,err) }
            }
        })
    }
    canceled,cancel := context.WithCancel(context.Background())
    cancel()
    if _,err := executeCanonicalBeadDelete(canceled,nil,canonicalBeadDeleteIntent{Identity:"task",Apply:true}); !errors.Is(err,context.Canceled) { t.Fatalf("cancellation did not precede store access: %v",err) }
}

// This exercises the new adapter operation against the actual checked store,
// not a fake mutation counter. CLI init is a separate completed process; the
// parent then owns one fresh embedded store for this exact regression. It is
// NOT a claim that the pending canonical profile decorator has been installed.
func TestCanonicalBeadDeleteNonCascadeEmbedded(t *testing.T) {
    isolateBeadsDirForTest(t)
    bd := buildBDUnderTest(t)
    const scope = "https://example.invalid/canonical-delete/"
    work,_ := graphListWorkspace(t,bd,"embedded",scope)
    cfg,err := configfile.LoadForDiscovery(filepath.Join(work,".beads"))
    if err != nil || cfg == nil { t.Fatalf("workspace binding: %v",err) }
    options := graphOptions(cfg)
    options.DataDir = filepath.Join(cfg.GraphWorkspace,"embeddeddolt")
    ctx,cancel := context.WithTimeout(t.Context(),90*time.Second)
    defer cancel()
    store,err := graphstore.OpenExisting(ctx,options)
    if err != nil { t.Fatal(err) }
    defer func(){ if err := store.Close(); err != nil { t.Error(err) } }()
    task,err := store.CreateIssue(ctx,"beads/native-task",publicops.CreateRequest{Actor:"author",ForceIDPrefix:true,Issue:&publicops.Issue{ID:"native-task",Title:"Native task",Status:types.StatusOpen,IssueType:types.TypeTask,Priority:2}})
    if err != nil { t.Fatal(err) }
    memory,err := store.Create(ctx,graphstore.CreateRequest{PublicID:"memory-note",Title:"Memory",Body:"Keep exact contents",Actor:"author"})
    if err != nil { t.Fatal(err) }
    memoryIdentity,err := store.ResolveCanonicalPublicResourceIdentity(ctx,memory.ID)
    if err != nil { t.Fatal(err) }
    relation,err := store.AddInformationalLink(ctx,graphstore.LinkCreateRequest{PublicID:"relation",TypeURL:graphstore.RelatedTypeURL(scope),SourcePath:"beads/native-task",TargetPath:memoryIdentity.PrivatePath,Actor:"author"})
    if err != nil { t.Fatal(err) }
    linkIdentity,err := store.ResolveCanonicalPublicResourceIdentity(ctx,relation.Link.ID)
    if err != nil { t.Fatal(err) }
    paths := []string{"beads/native-task",memoryIdentity.PrivatePath,linkIdentity.PrivatePath}
    capture := func() (graphstore.Snapshot,map[string]graphstore.VersionHistoryResult) {
        t.Helper()
        snapshot,err := store.CurrentSnapshot(ctx)
        if err != nil { t.Fatal(err) }
        history := make(map[string]graphstore.VersionHistoryResult,len(paths))
        for _,path := range paths { h,err := store.History(ctx,path); if err != nil { t.Fatal(err) }; history[path] = h }
        return snapshot,history
    }
    before,beforeHistory := capture()
    unchanged := func() {
        t.Helper()
        after,afterHistory := capture()
        if !reflect.DeepEqual(before,after) || !reflect.DeepEqual(beforeHistory,afterHistory) { t.Fatal("refused/previewed delete changed current or retained Resource state") }
    }
    for _,fixture := range []struct { id,path,revision string }{{task.ID,"beads/native-task",task.Revision},{memory.ID,memoryIdentity.PrivatePath,memory.Revision}} {
        expected,err := graphread.New(store).Resource(ctx,fixture.path)
        if err != nil { t.Fatal(err) }
        preview,err := executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:fixture.id})
        if err != nil || !preview.Preview || preview.Native != nil || !reflect.DeepEqual(preview.Resource,expected) { t.Fatalf("linked Bead preview was mutation admission/reconstruction: %+v %v",preview,err) }
        document,err := canonicalBeadDeletePreviewDocument(preview.Resource)
        if err != nil { t.Fatal(err) }
        var envelope struct { Kind string `json:"result_kind"`; Outcome string `json:"outcome"`; Data json.RawMessage `json:"data"`; Context struct { Preview string `json:"preview"` } `json:"context"` }
        if err := json.Unmarshal(document,&envelope); err != nil { t.Fatal(err) }
        var projected,original any
        expectedJSON,err := json.Marshal(expected)
        if err != nil { t.Fatal(err) }
        if json.Unmarshal(envelope.Data,&projected) != nil || json.Unmarshal(expectedJSON,&original) != nil || !reflect.DeepEqual(projected,original) || envelope.Kind != "resource" || envelope.Outcome != "success" || envelope.Context.Preview != "delete" { t.Fatalf("preview marker contaminated/lost Resource: %s",document) }
        _,err = executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:fixture.id,Apply:true,Guarded:true,Guard:clirevision.Selector{Token:fixture.revision},Actor:"deleter"})
        if !errors.Is(err,graphstore.ErrIncidentLinkConstraint) { t.Fatalf("implicit incident cascade or wrong refusal: %v",err) }
        unchanged()
    }
    wrongKind,err := executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:relation.Link.ID,Apply:true})
    var refusal *canonicalMutationLeafRefusal
    if !errors.As(err,&refusal) || refusal.ExitCode != 2 || wrongKind.Native != nil { t.Fatalf("Link target reached a Bead writer: %+v %v",wrongKind,err) }
    unchanged()
    if _,err := executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:"native",Apply:true}); !errors.Is(err,graphstore.ErrNotFound) { t.Fatalf("abbreviated ID was treated as exact: %v",err) }
    unchanged()
    canceled,stop := context.WithCancel(ctx)
    stop()
    if _,err := executeCanonicalBeadDelete(canceled,store,canonicalBeadDeleteIntent{Identity:task.ID,Apply:true}); !errors.Is(err,context.Canceled) { t.Fatalf("canceled deletion applied: %v",err) }
    unchanged()
    if _,err := executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:task.ID,Apply:true,Guarded:true,Guard:clirevision.Selector{Token:"stale"}}); !errors.Is(err,graphstore.ErrConflict) { t.Fatalf("stale guard lost: %v",err) }
    unchanged()
    if _,err := store.Unlink(ctx,graphstore.LinkDeleteRequest{Path:linkIdentity.PrivatePath,ExpectedRevision:relation.Link.Revision,DefaultInformationalSource:true,Actor:"unlinker"}); err != nil { t.Fatal(err) }
    finalTask,err := store.ShowIssue(ctx,"beads/native-task")
    if err != nil { t.Fatal(err) }
    ordinal,err := store.LocalRevisionForToken(ctx,"beads/native-task",finalTask.Revision)
    if err != nil { t.Fatal(err) }
    applied,err := executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:"native-task",Apply:true,Guarded:true,Guard:clirevision.Selector{Ordinal:ordinal.Ordinal},Actor:"deleter"})
    if err != nil { t.Fatal(err) }
    native,ok := applied.Native.(graphstore.IssueDeleteResult)
    if !ok || applied.Preview || !native.Deleted || native.Preview || native.Terminal == nil || !reflect.DeepEqual(native.Issue,finalTask) || len(native.IncidentLinkIDs) != 0 { t.Fatalf("actual checked noncascade result was replaced: %+v",applied) }
    if _,err := store.ShowIssue(ctx,"beads/native-task"); !errors.Is(err,graphstore.ErrGone) { t.Fatalf("Issue still current: %v",err) }
    retained,err := store.ReadVersion(ctx,"beads/native-task",finalTask.Revision)
    if err != nil || !reflect.DeepEqual(retained,finalTask) { t.Fatalf("final live Issue lost: %+v %v",retained,err) }
    // The same adapter keeps ordinary Memory deletion and the writer's exact
    // returned final snapshot; no native-only shortcut or batch wrapper.
    finalMemory,err := store.Show(ctx,memoryIdentity.PrivatePath)
    if err != nil { t.Fatal(err) }
    applied,err = executeCanonicalBeadDelete(ctx,store,canonicalBeadDeleteIntent{Identity:"memory-note",Apply:true,Actor:"deleter"})
    if err != nil { t.Fatal(err) }
    nativeMemory,ok := applied.Native.(graphstore.MemoryDeleteResult)
    if !ok || applied.Preview || !nativeMemory.Deleted || nativeMemory.Terminal == nil || !reflect.DeepEqual(nativeMemory.Memory,finalMemory) { t.Fatalf("Memory writer/result changed: %+v",applied) }
}
