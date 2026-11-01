package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/steveyegge/beads/internal/config"
	"github.com/steveyegge/beads/internal/httpclient"
)

const canonicalRetainedResponseLimit int64 = 262144

type canonicalRetainedBindings struct {
	Client      *http.Client
	Deadline    time.Time
	JournalRoot string
	Owner       string
}

// A submission handle couples the durable invocation to the one already
// authorized request. Neither leaves nor previews can submit a prepared value.
type canonicalRetainedSubmission struct {
	saved    canonicalRetainedSaved
	request  *http.Request
	provider string
}
type canonicalRetainedReadBinding struct{ options canonicalRetainedBindings }
type canonicalRetainedWriteBinding struct {
	options canonicalRetainedBindings
	scope   string
	targets map[string]string
	// Separate capability target; never a ninth ordinary directory member.
	batchTarget string
}

// Original response channels stay independent. Checked mutation/resource facets
// never replace NativeBody, and a Close failure cannot change Checked.
type canonicalRetainedReply struct {
	Status       int
	Headers      http.Header
	NativeBody   []byte
	OperationErr error
	ExchangeErr  error
	ReadErr      error
	LimitErr     error
	CloseErr     error
	ParseErr     error
	Sent         bool
	Checked      bool
	Resource     json.RawMessage
	Mutation     json.RawMessage
	Problem      json.RawMessage
}

func newCanonicalRetainedRootBindings(ctx context.Context, mutation bool) (canonicalRetainedBindings, error) {
	budget := 30 * time.Second
	if mutation {
		budget = 2 * time.Minute
	}
	deadline := time.Now().Add(budget)
	if parent, ok := ctx.Deadline(); ok && parent.Before(deadline) {
		deadline = parent
	}
	// A fresh transport has no alternate-protocol registration, auth retry or Host
	// rewrite. TLS uses the platform's normal verification; no redirects follow.
	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext, TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second, MaxResponseHeaderBytes: 65536, DisableKeepAlives: true}
	options := canonicalRetainedBindings{Client: &http.Client{Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, Deadline: deadline}
	if !mutation {
		return options, nil
	}
	// This strict read is mandatory even when explicit Scope selection bypassed
	// the default. Resolve the sibling only after strict parsing/error ordering.
	if _, _, err := config.UserYamlValueStrict("scope.url"); err != nil {
		return canonicalRetainedBindings{}, err
	}
	path, err := config.UserConfigYamlPath()
	if err != nil {
		return canonicalRetainedBindings{}, err
	}
	account, err := user.Current()
	if err != nil {
		return canonicalRetainedBindings{}, fmt.Errorf("cannot establish local recovery owner")
	}
	options.JournalRoot = filepath.Join(filepath.Dir(path), "retained-invocations")
	options.Owner = account.Uid + ":" + path
	return options, nil
}

func (o canonicalRetainedBindings) validateRead() error {
	if o.Client == nil || o.Deadline.IsZero() {
		return fmt.Errorf("missing bounded retained transport")
	}
	if o.Client.CheckRedirect == nil {
		return fmt.Errorf("retained transport must explicitly refuse redirects")
	}
	return nil
}
func (o canonicalRetainedBindings) validate(scope string) error {
	if err := o.validateRead(); err != nil {
		return err
	}
	if _, err := canonicalRetainedURL(scope); err != nil {
		return err
	}
	if o.Owner == "" || !filepath.IsAbs(o.JournalRoot) {
		return fmt.Errorf("missing private mutation custody context")
	}
	return nil
}
func canonicalRetainedURL(raw string) (*url.URL, error) {
	target, err := url.Parse(raw)
	if err != nil || target == nil || !target.IsAbs() || target.Host == "" || target.User != nil || target.Fragment != "" || target.Opaque != "" || (target.Scheme != "https" && target.Scheme != "http") {
		return nil, fmt.Errorf("retained target must be a whole absolute HTTP(S) URL without userinfo or fragment")
	}
	return target, nil
}

func canonicalRetainedAuthorize(ctx context.Context, req *http.Request) (string, error) {
	provider := httpclient.NewBearerProvider(req.URL)
	if err := provider.Authorize(ctx, req); err != nil {
		return "", fmt.Errorf("retained credential resolution failed")
	}
	// The existing provider warns but does not enforce this transport guard.
	// Honor the existing explicit environment opt-in, never a repository sidecar.
	host := req.URL.Hostname()
	address := net.ParseIP(host)
	loopback := strings.EqualFold(host, "localhost") || address != nil && address.IsLoopback()
	if req.Header.Get("Authorization") != "" && req.URL.Scheme == "http" && !loopback && os.Getenv("BEADS_HTTP_ALLOW_INSECURE") != "1" {
		return "", fmt.Errorf("retained credential requires TLS or explicit BEADS_HTTP_ALLOW_INSECURE=1")
	}
	return provider.Source(), nil
}

func (o canonicalRetainedBindings) exchange(ctx context.Context, request *http.Request) canonicalRetainedReply {
	// GetBody=nil prevents net/http's replay door even when the transport sees a
	// keyed request. A fresh transport is used by the normal root constructor.
	request.GetBody = nil
	response, err := o.Client.Do(request)
	reply := canonicalRetainedReply{Sent: true, ExchangeErr: err}
	if response == nil {
		return reply
	}
	reply.Status, reply.Headers = response.StatusCode, response.Header.Clone()
	if response.Body == nil {
		reply.ReadErr = fmt.Errorf("retained response lacks body custody")
		return reply
	}
	body, readErr := io.ReadAll(io.LimitReader(response.Body, canonicalRetainedResponseLimit+1))
	reply.CloseErr = response.Body.Close()
	reply.NativeBody = bytes.Clone(body)
	if int64(len(body)) > canonicalRetainedResponseLimit {
		reply.LimitErr = fmt.Errorf("retained response exceeds bound")
	}
	reply.ReadErr = readErr
	return reply
}
func (o canonicalRetainedBindings) get(ctx context.Context, target string) canonicalRetainedReply {
	if _, err := canonicalRetainedURL(target); err != nil {
		return canonicalRetainedReply{OperationErr: err}
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return canonicalRetainedReply{OperationErr: err}
	}
	request.Header.Set("Accept", canonicalRetainedMedia)
	if _, err := canonicalRetainedAuthorize(ctx, request); err != nil {
		return canonicalRetainedReply{OperationErr: err}
	}
	return o.exchange(ctx, request)
}

func canonicalRetainedJSON(reply canonicalRetainedReply, expected int) error {
	if reply.OperationErr != nil || reply.acquisitionFailed() {
		return reply.Err()
	}
	if reply.Status != expected {
		return errors.Join(fmt.Errorf("retained HTTP status %d", reply.Status), reply.CloseErr)
	}
	media, _, err := mime.ParseMediaType(reply.Headers.Get("Content-Type"))
	if err != nil || media != canonicalRetainedMedia {
		return errors.Join(fmt.Errorf("retained representation binding mismatch"), reply.CloseErr)
	}
	if err := canonicalRetainedStrictJSON(reply.NativeBody); err != nil {
		return errors.Join(err, reply.CloseErr)
	}
	return reply.CloseErr
}

// Link values are split only outside angle/quoted strings. Ambiguous duplicate
// service-desc links refuse; targets must already be absolute, never guessed.
func canonicalRetainedServiceDescription(headers http.Header) (string, error) {
	var found string
	for _, header := range headers.Values("Link") {
		start, angle, quoted, escaped := 0, false, false, false
		var parts []string
		for i, c := range header {
			if escaped {
				escaped = false
				continue
			}
			if quoted && c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				quoted = !quoted
			}
			if !quoted {
				if c == '<' {
					angle = true
				}
				if c == '>' {
					angle = false
				}
				if c == ',' && !angle {
					parts = append(parts, header[start:i])
					start = i + 1
				}
			}
		}
		if angle || quoted {
			return "", fmt.Errorf("malformed retained Link field")
		}
		parts = append(parts, header[start:])
		for _, part := range parts {
			part = strings.TrimSpace(part)
			end := strings.IndexByte(part, '>')
			if !strings.HasPrefix(part, "<") || end < 1 {
				return "", fmt.Errorf("malformed retained Link field")
			}
			_, params, err := mime.ParseMediaType("link" + part[end+1:])
			if err != nil {
				return "", fmt.Errorf("malformed retained Link parameters")
			}
			for _, rel := range strings.Fields(params["rel"]) {
				if rel != "service-desc" {
					continue
				}
				target := part[1:end]
				if found != "" || params["type"] != canonicalRetainedMedia {
					return "", fmt.Errorf("ambiguous or incompatible retained service-desc")
				}
				if _, err := canonicalRetainedURL(target); err != nil {
					return "", err
				}
				found = target
			}
		}
	}
	if found == "" {
		return "", fmt.Errorf("missing explicit owning service-desc")
	}
	return found, nil
}

type canonicalRetainedDiscovery struct {
	BDPVersion   string          `json:"bdpVersion"`
	Profile      string          `json:"profile"`
	Scope        string          `json:"scope"`
	Beads        string          `json:"beads"`
	Links        string          `json:"links"`
	Events       string          `json:"events"`
	Aliases      string          `json:"aliases"`
	Operations   string          `json:"operations"`
	Capabilities json.RawMessage `json:"capabilities,omitempty"`
}

func (o canonicalRetainedBindings) description(ctx context.Context, target string) (canonicalRetainedDiscovery, error) {
	reply := o.get(ctx, target)
	if err := canonicalRetainedJSON(reply, http.StatusOK); err != nil {
		return canonicalRetainedDiscovery{}, err
	}
	return parseCanonicalRetainedDiscovery(reply.NativeBody)
}

func discoverCanonicalRetainedWrite(ctx context.Context, scope string, options canonicalRetainedBindings) (canonicalRetainedWriteBinding, error) {
	landing := options.get(ctx, scope)
	if err := canonicalRetainedScopeLanding(landing); err != nil {
		return canonicalRetainedWriteBinding{}, err
	}
	target, err := canonicalRetainedServiceDescription(landing.Headers)
	if err != nil {
		return canonicalRetainedWriteBinding{}, err
	}
	desc, err := options.description(ctx, target)
	if err != nil {
		return canonicalRetainedWriteBinding{}, err
	}
	if desc.Scope != scope || desc.Profile != "write" {
		return canonicalRetainedWriteBinding{}, fmt.Errorf("selected Scope has no retained Write binding")
	}
	directory := options.get(ctx, desc.Operations)
	if err := canonicalRetainedJSON(directory, http.StatusOK); err != nil {
		return canonicalRetainedWriteBinding{}, err
	}
	var targets map[string]string
	if err := json.Unmarshal(directory.NativeBody, &targets); err != nil {
		return canonicalRetainedWriteBinding{}, err
	}
	names := []string{"createBead", "updateBead", "deleteBead", "createLink", "updateLink", "deleteLink", "putAlias", "deleteAlias"}
	if len(targets) != len(names) {
		return canonicalRetainedWriteBinding{}, fmt.Errorf("retained operation directory is not the selected closed shape")
	}
	seen := map[string]bool{}
	for _, name := range names {
		raw := targets[name]
		if _, err := canonicalRetainedURL(raw); err != nil {
			return canonicalRetainedWriteBinding{}, err
		}
		if seen[raw] {
			return canonicalRetainedWriteBinding{}, fmt.Errorf("retained operation URLs are not distinct")
		}
		seen[raw] = true
	}
	return canonicalRetainedWriteBinding{options: options, scope: scope, targets: targets}, nil
}
func (b canonicalRetainedReadBinding) readResource(ctx context.Context, target string) canonicalRetainedReply {
	reply := b.options.get(ctx, target)
	if reply.acquisitionFailed() {
		return reply
	}
	if reply.OperationErr == nil && reply.Status != http.StatusOK {
		media, _, err := mime.ParseMediaType(reply.Headers.Get("Content-Type"))
		if err == nil && media == "application/problem+json" && canonicalRetainedProblem(reply.NativeBody, reply.Status, canonicalRetainedReadProblem) == nil {
			reply.Problem = bytes.Clone(reply.NativeBody)
			reply.OperationErr = fmt.Errorf("retained read returned HTTP %d", reply.Status)
			reply.Checked = true
			return reply
		}
	}
	// A read/Close error is retained independently; a full valid body can remain
	// known even if its own Close failed. Discovery errors become ParseErr.
	check := reply
	check.CloseErr = nil
	if err := canonicalRetainedJSON(check, http.StatusOK); err != nil {
		reply.ParseErr = err
		return reply
	}
	description, err := canonicalRetainedServiceDescription(reply.Headers)
	if err != nil {
		reply.ParseErr = err
		return reply
	}
	if _, err := b.options.description(ctx, description); err != nil {
		reply.ParseErr = err
		return reply
	}
	kind := reply.Headers.Get("BDP-Resource-Kind")
	if kind != "bead" && kind != "link" {
		reply.ParseErr = fmt.Errorf("missing selected Resource kind")
		return reply
	}
	if err := validateCanonicalRetainedCurrentResource(reply.NativeBody, target); err != nil {
		reply.ParseErr = err
		return reply
	}
	fields, _ := canonicalRetainedObject(reply.NativeBody)
	if (fields["source"] != nil) != (kind == "link") {
		reply.ParseErr = fmt.Errorf("Resource kind contradicts selected HTTP context")
		return reply
	}
	reply.Resource = bytes.Clone(reply.NativeBody)
	reply.Checked = true
	return reply
}

func (b canonicalRetainedWriteBinding) apply(ctx context.Context, prepared canonicalRetainedPrepared) (canonicalRetainedReply, string) {
	target := b.targets[string(prepared.Operation)]
	if prepared.Operation == retainedAtomicBatch {
		target = b.batchTarget
	}
	if target == "" || prepared.Scope != b.scope || prepared.Contract != canonicalRetainedContract || prepared.Method != http.MethodPost || prepared.Target != target {
		return canonicalRetainedReply{OperationErr: fmt.Errorf("retained request binding changed")}, ""
	}
	if prepared.Key == "" {
		var key [32]byte
		if _, err := rand.Read(key[:]); err != nil {
			return canonicalRetainedReply{OperationErr: err}, ""
		}
		prepared.Key = hex.EncodeToString(key[:])
	}
	request, provider, err := b.request(ctx, prepared)
	if err != nil {
		return canonicalRetainedReply{OperationErr: err}, ""
	}
	saved, err := saveCanonicalRetainedInvocation(b.options.JournalRoot, canonicalRetainedJournalRecord{Version: 1, Owner: b.options.Owner, Provider: provider, Request: prepared})
	if err != nil {
		return canonicalRetainedReply{OperationErr: err}, ""
	}
	return b.submitSaved(ctx, canonicalRetainedSubmission{saved: saved, request: request, provider: provider}), saved.path
}
func (b canonicalRetainedWriteBinding) request(ctx context.Context, prepared canonicalRetainedPrepared) (*http.Request, string, error) {
	if len(prepared.Body) > 65536 || !canonicalRetainedKey(prepared.Key) {
		return nil, "", fmt.Errorf("retained request exceeds admission bounds")
	}
	if _, err := canonicalRetainedURL(prepared.Target); err != nil {
		return nil, "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, prepared.Target, bytes.NewReader(prepared.Body))
	if err != nil {
		return nil, "", err
	}
	request.GetBody = nil
	request.Header.Set("Accept", canonicalRetainedMedia)
	request.Header.Set("Content-Type", canonicalRetainedMedia)
	request.Header.Set("Idempotency-Key", prepared.Key)
	provider, err := canonicalRetainedAuthorize(ctx, request)
	return request, provider, err
}
func (b canonicalRetainedWriteBinding) submitSaved(ctx context.Context, submission canonicalRetainedSubmission) canonicalRetainedReply {
	saved, request, provider := submission.saved, submission.request, submission.provider
	if saved.path == "" || request == nil || saved.record.Owner != b.options.Owner || saved.record.Provider != provider {
		return canonicalRetainedReply{OperationErr: fmt.Errorf("retained recovery credential owner/provider context changed")}
	}
	// Request was authorized before durable save. Its body/key/target are created
	// from that very prepared value, and no provider is refreshed after save.
	reply := b.options.exchange(ctx, request)
	if saved.record.Request.Operation == retainedAtomicBatch {
		reply.validateBatch(saved.record.Request.Body, saved.record.Request.Scope)
	} else {
		reply.validateMutation(saved.record.Request.Operation)
	}
	return reply
}
func (b canonicalRetainedWriteBinding) retry(ctx context.Context, path string, op canonicalRetainedOperation) (canonicalRetainedReply, string) {
	saved, err := loadCanonicalRetainedInvocation(b.options.JournalRoot, path)
	if err != nil {
		return canonicalRetainedReply{OperationErr: err}, ""
	}
	// Re-establish local durability barriers after a crash/reopen before POST.
	saved, err = saveCanonicalRetainedInvocation(b.options.JournalRoot, saved.record)
	if err != nil {
		return canonicalRetainedReply{OperationErr: err}, path
	}
	prepared := saved.record.Request
	if prepared.Operation != op || prepared.Scope != b.scope || prepared.Target != b.targets[string(prepared.Operation)] {
		return canonicalRetainedReply{OperationErr: fmt.Errorf("retained recovery navigation changed")}, path
	}
	request, provider, err := b.request(ctx, prepared)
	if err != nil {
		return canonicalRetainedReply{OperationErr: err}, path
	}
	// Retry is explicit and sends the original key and meaning once. It never
	// substitutes fresh operands, current defaults or a newly generated key.
	return b.submitSaved(ctx, canonicalRetainedSubmission{saved: saved, request: request, provider: provider}), path
}

// Token admission preserves numeric spelling in the original RawMessages. The
// decoder never converts JSON numbers to float64. Duplicate decoded keys and a
// trailing second value refuse before any document is interpreted.
func canonicalRetainedStrictJSON(raw []byte) error {
	if !utf8.Valid(raw) {
		return fmt.Errorf("retained JSON is not valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value func(int) error
	value = func(depth int) error {
		if depth > 128 {
			return fmt.Errorf("retained JSON nesting exceeds bound")
		}
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				token, err := decoder.Token()
				if err != nil {
					return err
				}
				key, ok := token.(string)
				if !ok || seen[key] {
					return fmt.Errorf("duplicate or invalid retained JSON key")
				}
				seen[key] = true
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			token, err = decoder.Token()
			if err != nil || token != json.Delim('}') {
				return fmt.Errorf("unterminated retained object")
			}
		case '[':
			for decoder.More() {
				if err := value(depth + 1); err != nil {
					return err
				}
			}
			token, err = decoder.Token()
			if err != nil || token != json.Delim(']') {
				return fmt.Errorf("unterminated retained array")
			}
		default:
			return fmt.Errorf("unexpected retained JSON delimiter")
		}
		return nil
	}
	if err := value(0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("retained JSON has trailing data")
	}
	return nil
}
func canonicalRetainedObject(raw []byte, allowed ...string) (map[string]json.RawMessage, error) {
	if err := canonicalRetainedStrictJSON(raw); err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, fmt.Errorf("retained value must be an object")
	}
	if len(allowed) != 0 {
		names := map[string]bool{}
		for _, name := range allowed {
			names[name] = true
		}
		for name := range fields {
			if !names[name] {
				return nil, fmt.Errorf("unknown retained member %s", name)
			}
		}
	}
	return fields, nil
}
func canonicalRetainedString(raw json.RawMessage) (string, error) {
	var value string
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, &value) != nil || value == "" {
		return "", fmt.Errorf("retained member must be a nonempty string")
	}
	return value, nil
}
func canonicalRetainedReference(raw json.RawMessage, httpOnly bool) error {
	uri, err := canonicalRetainedString(raw)
	if err != nil {
		fields, e := canonicalRetainedObject(raw, "uri", "revision")
		if e != nil {
			return e
		}
		uri, e = canonicalRetainedString(fields["uri"])
		if e != nil {
			return e
		}
		if _, e := canonicalRetainedString(fields["revision"]); e != nil {
			return e
		}
	}
	if httpOnly {
		_, err = canonicalRetainedURL(uri)
		return err
	}
	parsed, e := url.Parse(uri)
	if e != nil || !parsed.IsAbs() {
		return fmt.Errorf("retained Reference must be absolute")
	}
	return nil
}
func validateCanonicalRetainedResource(raw []byte, target string) error {
	return validateCanonicalRetainedResourceContext(raw, target, false)
}
func validateCanonicalRetainedCurrentResource(raw []byte, target string) error {
	return validateCanonicalRetainedResourceContext(raw, target, true)
}
func validateCanonicalRetainedResourceContext(raw []byte, target string, current bool) error {
	fields, err := canonicalRetainedObject(raw, "id", "type", "revision", "properties", "metadata", "links", "ownedLinks", "revisionProvenance", "source", "target")
	if err != nil {
		return err
	}
	id, err := canonicalRetainedString(fields["id"])
	if err != nil {
		return err
	}
	if _, err := canonicalRetainedURL(id); err != nil {
		return err
	}
	if target != "" && id != target {
		return fmt.Errorf("retained Resource identity differs from the requested whole URI")
	}
	if _, err := canonicalRetainedString(fields["revision"]); err != nil {
		return err
	}
	if err := canonicalRetainedReference(fields["type"], true); err != nil {
		return err
	}
	if _, err := canonicalRetainedObject(fields["properties"]); err != nil {
		return err
	}
	if value, ok := fields["metadata"]; ok {
		if _, err := canonicalRetainedObject(value); err != nil {
			return err
		}
	} else if current {
		return fmt.Errorf("current full Resource requires metadata")
	}
	if (fields["source"] == nil) != (fields["target"] == nil) {
		return fmt.Errorf("incomplete retained Link endpoints")
	}
	if fields["source"] != nil {
		if fields["links"] != nil || fields["ownedLinks"] != nil {
			return fmt.Errorf("Link has Bead-only members")
		}
		for _, name := range []string{"source", "target"} {
			if err := canonicalRetainedReference(fields[name], false); err != nil {
				return err
			}
		}
	}
	if value, ok := fields["revisionProvenance"]; ok {
		if err := canonicalRetainedProvenance(value); err != nil {
			return err
		}
	}
	if value, ok := fields["links"]; ok {
		collection, e := canonicalRetainedObject(value, "items", "next")
		if e != nil {
			return e
		}
		if collection["items"] == nil || collection["next"] == nil {
			return fmt.Errorf("incomplete Link collection")
		}
		var links []json.RawMessage
		if bytes.Equal(collection["items"], []byte("null")) || json.Unmarshal(collection["items"], &links) != nil {
			return fmt.Errorf("invalid Link collection items")
		}
		for _, link := range links {
			if e := validateCanonicalRetainedResourceContext(link, "", current); e != nil {
				return e
			}
			member, _ := canonicalRetainedObject(link)
			if member["source"] == nil {
				return fmt.Errorf("Link collection contains a Bead")
			}
		}
		if !bytes.Equal(collection["next"], []byte("null")) {
			next, e := canonicalRetainedString(collection["next"])
			if e != nil {
				return e
			}
			if _, e := canonicalRetainedURL(next); e != nil {
				return e
			}
		}
	}
	if value, ok := fields["ownedLinks"]; ok {
		groups, e := canonicalRetainedObject(value)
		if e != nil {
			return e
		}
		for typ, raw := range groups {
			if _, e := canonicalRetainedURL(typ); e != nil {
				return e
			}
			var links []json.RawMessage
			if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, &links) != nil {
				return fmt.Errorf("invalid owned Links")
			}
			for _, link := range links {
				if e := validateCanonicalRetainedResourceContext(link, "", current); e != nil {
					return e
				}
				member, _ := canonicalRetainedObject(link)
				if member["source"] == nil {
					return fmt.Errorf("owned Link group contains a Bead")
				}
			}
		}
	}
	return nil
}
func (r *canonicalRetainedReply) validateMutation(op canonicalRetainedOperation) {
	if r.Checked || r.ParseErr != nil || r.OperationErr != nil || r.acquisitionFailed() {
		return
	}
	if r.Status != http.StatusOK {
		media, _, err := mime.ParseMediaType(r.Headers.Get("Content-Type"))
		if err != nil || media != "application/problem+json" || canonicalRetainedProblem(r.NativeBody, r.Status, canonicalRetainedWriteProblem) != nil {
			r.ParseErr = fmt.Errorf("unresolved retained operation response")
			return
		}
		// Store original Problem; transport status alone makes no statement about a
		// durable key claim. Admission/transience/disclosure remain native semantics.
		r.Problem = bytes.Clone(r.NativeBody)
		r.OperationErr = fmt.Errorf("retained operation returned HTTP %d", r.Status)
		r.Checked = true
		return
	}
	check := *r
	check.CloseErr = nil
	if err := canonicalRetainedJSON(check, http.StatusOK); err != nil {
		r.ParseErr = err
		return
	}
	fields, err := canonicalRetainedObject(r.NativeBody, "outcome", "changed", "resource", "deleted", "ownerRevisions")
	if err != nil {
		r.ParseErr = err
		return
	}
	outcome, err := canonicalRetainedString(fields["outcome"])
	if err != nil {
		r.ParseErr = err
		return
	}
	expected := "updated"
	if op == retainedCreateBead || op == retainedCreateLink {
		expected = "created"
	}
	if op == retainedDeleteBead || op == retainedDeleteLink {
		expected = "deleted"
	}
	if outcome != expected {
		r.ParseErr = fmt.Errorf("retained result outcome contradicts operation")
		return
	}
	var changed bool
	if fields["changed"] == nil || bytes.Equal(fields["changed"], []byte("null")) || json.Unmarshal(fields["changed"], &changed) != nil || outcome != "updated" && !changed {
		r.ParseErr = fmt.Errorf("invalid retained changed disposition")
		return
	}
	link := op == retainedCreateLink || op == retainedUpdateLink || op == retainedDeleteLink
	if outcome == "deleted" {
		if fields["resource"] != nil {
			r.ParseErr = fmt.Errorf("deleted result contains a live Resource")
			return
		}
		deleted, e := canonicalRetainedObject(fields["deleted"], "resourceKind", "resource")
		if e != nil {
			r.ParseErr = e
			return
		}
		kind, e := canonicalRetainedString(deleted["resourceKind"])
		if e != nil || (kind == "link") != link || kind != "bead" && kind != "link" {
			r.ParseErr = fmt.Errorf("deleted Resource kind mismatch")
			return
		}
		identity, e := canonicalRetainedObject(deleted["resource"], "id", "type", "revision")
		if e != nil {
			r.ParseErr = e
			return
		}
		id, e := canonicalRetainedString(identity["id"])
		if e != nil {
			r.ParseErr = e
			return
		}
		if _, e = canonicalRetainedURL(id); e != nil {
			r.ParseErr = e
			return
		}
		if _, e = canonicalRetainedString(identity["revision"]); e != nil {
			r.ParseErr = e
			return
		}
		if e = canonicalRetainedReference(identity["type"], true); e != nil {
			r.ParseErr = e
			return
		}
	} else {
		if fields["deleted"] != nil {
			r.ParseErr = fmt.Errorf("live result contains deleted identity")
			return
		}
		if err := validateCanonicalRetainedResource(fields["resource"], ""); err != nil {
			r.ParseErr = err
			return
		}
		resource, _ := canonicalRetainedObject(fields["resource"])
		if (resource["source"] != nil) != link {
			r.ParseErr = fmt.Errorf("retained mutation Resource kind mismatch")
			return
		}
	}
	if !link && fields["ownerRevisions"] != nil {
		r.ParseErr = fmt.Errorf("Bead result has Link owner revisions")
		return
	}
	if value, ok := fields["ownerRevisions"]; ok {
		var owners []json.RawMessage
		if json.Unmarshal(value, &owners) != nil || len(owners) < 1 || len(owners) > 2 {
			r.ParseErr = fmt.Errorf("invalid owner revisions")
			return
		}
		seen := map[string]bool{}
		for _, owner := range owners {
			identity, e := canonicalRetainedObject(owner, "id", "revision")
			if e != nil {
				r.ParseErr = e
				return
			}
			id, e := canonicalRetainedString(identity["id"])
			if e != nil {
				r.ParseErr = e
				return
			}
			if _, e := canonicalRetainedURL(id); e != nil {
				r.ParseErr = e
				return
			}
			if _, e := canonicalRetainedString(identity["revision"]); e != nil {
				r.ParseErr = e
				return
			}
			if seen[id] {
				r.ParseErr = fmt.Errorf("duplicate owner identity")
				return
			}
			seen[id] = true
		}
	}
	r.Mutation = bytes.Clone(r.NativeBody)
	r.Checked = true
}

func canonicalRetainedActor(raw []byte) error {
	fields, err := canonicalRetainedObject(raw, "id", "basis")
	if err != nil {
		return err
	}
	if _, err := canonicalRetainedString(fields["id"]); err != nil {
		return err
	}
	basis, err := canonicalRetainedString(fields["basis"])
	if err != nil || basis != "writer-supplied" && basis != "unknown" {
		return fmt.Errorf("invalid revision actor basis")
	}
	return nil
}
func canonicalRetainedProvenance(raw []byte) error {
	fields, err := canonicalRetainedObject(raw, "actor", "committedAt", "agent", "message")
	if err != nil {
		return err
	}
	if len(fields) == 0 {
		return fmt.Errorf("empty revision provenance")
	}
	if fields["actor"] != nil {
		if err := canonicalRetainedActor(fields["actor"]); err != nil {
			return err
		}
	}
	contextual := fields["committedAt"] != nil || fields["agent"] != nil || fields["message"] != nil
	if contextual {
		for _, name := range []string{"committedAt", "agent", "message"} {
			context, err := canonicalRetainedObject(fields[name], "state", "value")
			if err != nil {
				return err
			}
			state, err := canonicalRetainedString(context["state"])
			if err != nil {
				return err
			}
			switch state {
			case "present":
				var value string
				if context["value"] == nil || bytes.Equal(context["value"], []byte("null")) || json.Unmarshal(context["value"], &value) != nil || name != "message" && value == "" {
					return fmt.Errorf("invalid provenance context value")
				}
				if name == "committedAt" {
					if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
						return err
					}
				}
			case "undetermined", "absent":
				if name == "committedAt" && state == "absent" || context["value"] != nil {
					return fmt.Errorf("invalid absent provenance context")
				}
			default:
				return fmt.Errorf("invalid provenance context state")
			}
		}
	}
	return nil
}
func canonicalRetainedProvenanceInput(raw []byte) error {
	fields, err := canonicalRetainedObject(raw, "actor", "agent", "message")
	if err != nil {
		return err
	}
	if value := fields["actor"]; value != nil {
		if err := canonicalRetainedActor(value); err != nil {
			return err
		}
	}
	for _, name := range []string{"agent", "message"} {
		if value := fields[name]; value != nil && !bytes.Equal(value, []byte("null")) {
			var text string
			if json.Unmarshal(value, &text) != nil || name == "agent" && text == "" {
				return fmt.Errorf("invalid provenance input")
			}
		}
	}
	return nil
}

type canonicalRetainedProblemFamily uint8

const (
	canonicalRetainedReadProblem canonicalRetainedProblemFamily = iota
	canonicalRetainedWriteProblem
)

func canonicalRetainedReadProblemCode(code string) bool {
	switch code {
	case "malformed-request", "invalid-parameter", "unauthenticated", "forbidden", "resource-not-found", "resource-pruned", "resource-erased", "foreign-view", "cursor-expired", "request-too-large", "limit-exceeded", "rate-limited", "temporarily-unavailable", "revision-unknown", "revision-unretained", "revision-reorganized", "revision-not-tracked", "revision-unrepresentable":
		return true
	default:
		return false
	}
}
func canonicalRetainedProblem(raw []byte, status int, family canonicalRetainedProblemFamily) error {
	// The selected Problem schemas permit extension members. Validate declared
	// fields and explicit conditional non-disclosure prohibitions, preserving
	// opaque permitted extensions and the complete original body.
	fields, err := canonicalRetainedObject(raw)
	if err != nil {
		return err
	}
	typ, err := canonicalRetainedString(fields["type"])
	if err != nil {
		return err
	}
	if _, err := canonicalRetainedURL(typ); err != nil {
		return err
	}
	code, err := canonicalRetainedString(fields["code"])
	if err != nil {
		return err
	}
	switch family {
	case canonicalRetainedReadProblem:
		if !canonicalRetainedReadProblemCode(code) {
			return fmt.Errorf("native Problem code is outside the selected Read family")
		}
	case canonicalRetainedWriteProblem:
	default:
		return fmt.Errorf("unknown selected Problem family")
	}
	retry, err := canonicalRetainedString(fields["retry"])
	if err != nil {
		return err
	}
	wantType, wantStatus, wantRetry, ok := canonicalRetainedProblemBinding(code)
	if !ok || typ != wantType || status != wantStatus || retry != wantRetry {
		return fmt.Errorf("native Problem contradicts selected diagnosis binding")
	}
	for _, name := range []string{"title", "detail"} {
		if raw, ok := fields[name]; ok {
			if _, err := canonicalRetainedJSONString(raw); err != nil {
				return err
			}
		}
	}
	if raw, ok := fields["instance"]; ok {
		if err := canonicalRetainedAbsoluteURI(raw); err != nil {
			return err
		}
	}
	if raw, ok := fields["status"]; ok {
		if err := canonicalRetainedInteger(raw, 100); err != nil {
			return err
		}
		stated, _ := new(big.Rat).SetString(strings.TrimSpace(string(raw)))
		if stated.Cmp(big.NewRat(int64(status), 1)) != 0 {
			return fmt.Errorf("Problem HTTP status mismatch")
		}
	}
	if family == canonicalRetainedWriteProblem {
		if raw, ok := fields["retryAfter"]; ok {
			if retry != "after-delay" {
				return fmt.Errorf("retryAfter is prohibited for this diagnosis")
			}
			if err := canonicalRetainedInteger(raw, 0); err != nil {
				return err
			}
		}
	}
	if raw, ok := fields["archivedAt"]; ok {
		if code != "resource-pruned" {
			return fmt.Errorf("archive reference is prohibited for this diagnosis")
		}
		if err := canonicalRetainedReference(raw, false); err != nil {
			return err
		}
	}
	if code == "revision-unretained" {
		if err := canonicalRetainedHistoryMissing(fields["missing"]); err != nil {
			return err
		}
	} else if _, ok := fields["missing"]; ok {
		return fmt.Errorf("missing history is prohibited for this diagnosis")
	}
	// Read permits these names as opaque extensions; only the Write family
	// declares diagnostics/truncation and their diagnosis-specific restrictions.
	if family == canonicalRetainedWriteProblem {
		if code == "validation-failed" {
			if err := canonicalRetainedDiagnostics(fields["diagnostics"]); err != nil {
				return err
			}
			if raw, ok := fields["diagnosticsTruncated"]; ok {
				if _, err := canonicalRetainedBoolean(raw); err != nil {
					return err
				}
			}
		} else if fields["diagnostics"] != nil || fields["diagnosticsTruncated"] != nil {
			return fmt.Errorf("validation diagnostics are prohibited for this diagnosis")
		}
	}
	var forbidden []string
	switch code {
	case "resource-erased":
		forbidden = []string{"pointer"}
	case "revision-unknown", "revision-unretained", "revision-reorganized", "revision-not-tracked", "revision-unrepresentable":
		forbidden = []string{"window", "mayChangeAfterSync", "resource", "properties", "ownedLinks", "archivedAt", "pointer", "revisionProvenance"}
	case "revision-allocation-unsafe":
		forbidden = []string{"missing", "window", "resource", "properties", "ownedLinks", "pointer", "identity", "revision", "revisionProvenance"}
	}
	for _, name := range forbidden {
		if _, ok := fields[name]; ok {
			return fmt.Errorf("Problem discloses prohibited member %s", name)
		}
	}
	return nil
}

func canonicalRetainedDiagnostics(raw []byte) error {
	diagnostics, err := canonicalRetainedArray(raw)
	if err != nil || len(diagnostics) == 0 {
		return fmt.Errorf("validation-failed requires nonempty diagnostics")
	}
	for _, raw := range diagnostics {
		fields, err := canonicalRetainedObject(raw, "type", "schemaLocation", "document", "instanceLocation", "message")
		if err != nil {
			return err
		}
		if _, err := canonicalRetainedString(fields["message"]); err != nil {
			return err
		}
		if (fields["type"] != nil) != (fields["schemaLocation"] != nil) {
			return fmt.Errorf("diagnostic Type and schemaLocation are paired")
		}
		if raw, ok := fields["type"]; ok {
			text, err := canonicalRetainedString(raw)
			if err != nil {
				return err
			}
			if _, err := canonicalRetainedURL(text); err != nil {
				return err
			}
			if err := canonicalRetainedAbsoluteURI(fields["schemaLocation"]); err != nil {
				return err
			}
		}
		if raw, ok := fields["document"]; ok {
			document, err := canonicalRetainedString(raw)
			if err != nil || document != "properties" && document != "metadata" {
				return fmt.Errorf("invalid diagnostic document")
			}
		}
		if raw, ok := fields["instanceLocation"]; ok {
			if _, err := canonicalRetainedJSONPointer(raw); err != nil {
				return err
			}
		}
	}
	return nil
}
func canonicalRetainedHistoryMissing(raw []byte) error {
	fields, err := canonicalRetainedObject(raw, "complete", "items")
	if err != nil {
		return err
	}
	complete, err := canonicalRetainedBoolean(fields["complete"])
	if err != nil {
		return err
	}
	items, err := canonicalRetainedArray(fields["items"])
	if err != nil || len(items) > 64 || complete && len(items) == 0 {
		return fmt.Errorf("invalid missing-history item bounds")
	}
	seen := map[string]bool{}
	for _, raw := range items {
		item, err := canonicalRetainedObject(raw, "kind", "pointer", "type")
		if err != nil {
			return err
		}
		kind, err := canonicalRetainedString(item["kind"])
		if err != nil {
			return err
		}
		pointer, typ := "", ""
		switch kind {
		case "record":
			if len(item) != 1 {
				return fmt.Errorf("record missing item has extra members")
			}
		case "property":
			if item["type"] != nil || item["pointer"] == nil {
				return fmt.Errorf("invalid property missing item")
			}
			pointer, err = canonicalRetainedJSONPointer(item["pointer"])
			if err != nil {
				return err
			}
			matches, _ := regexp.MatchString(`^/(properties(?:/|$)|ownedLinks/([^/~]|~[01])+/(0|[1-9][0-9]*)/properties(?:/|$))`, pointer)
			if !matches {
				return fmt.Errorf("missing property pointer is outside selected paths")
			}
		case "owned-links":
			if item["pointer"] != nil {
				return fmt.Errorf("owned-links missing item has pointer")
			}
			if raw, ok := item["type"]; ok {
				typ, err = canonicalRetainedString(raw)
				if err != nil {
					return err
				}
				if _, err := canonicalRetainedURL(typ); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown missing-history item kind")
		}
		// Decoded strings, rather than raw JSON spelling, establish uniqueItems.
		identity, _ := json.Marshal([]string{kind, pointer, typ})
		key := string(identity)
		if seen[key] {
			return fmt.Errorf("duplicate missing-history item")
		}
		seen[key] = true
	}
	return nil
}

// Literal correspondence to the fixed selected schema Problem table. This
// describes returned diagnosis only, never whether an authority retained a key.
func canonicalRetainedProblemBinding(code string) (string, int, string, bool) {
	switch code {
	case "malformed-request":
		return "https://github.com/gastownhall/bdp/problems/request", 400, "never", true
	case "invalid-parameter":
		return "https://github.com/gastownhall/bdp/problems/request", 400, "never", true
	case "unauthenticated":
		return "https://github.com/gastownhall/bdp/problems/authentication", 401, "after-state-change", true
	case "forbidden":
		return "https://github.com/gastownhall/bdp/problems/authorization", 403, "after-state-change", true
	case "resource-not-found":
		return "https://github.com/gastownhall/bdp/problems/not-found", 404, "after-state-change", true
	case "foreign-view":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "cursor-expired":
		return "https://github.com/gastownhall/bdp/problems/gone", 410, "after-state-change", true
	case "request-too-large":
		return "https://github.com/gastownhall/bdp/problems/size", 413, "never", true
	case "limit-exceeded":
		return "https://github.com/gastownhall/bdp/problems/size", 413, "never", true
	case "rate-limited":
		return "https://github.com/gastownhall/bdp/problems/rate-limit", 429, "after-delay", true
	case "temporarily-unavailable":
		return "https://github.com/gastownhall/bdp/problems/unavailable", 503, "after-delay", true
	case "resource-pruned":
		return "https://github.com/gastownhall/bdp/problems/gone", 410, "never", true
	case "resource-erased":
		return "https://github.com/gastownhall/bdp/problems/gone", 410, "never", true
	case "revision-unknown":
		return "https://github.com/gastownhall/bdp/problems/not-found", 404, "after-state-change", true
	case "revision-unretained":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "revision-reorganized":
		return "https://github.com/gastownhall/bdp/problems/gone", 410, "after-state-change", true
	case "revision-not-tracked":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "revision-unrepresentable":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "unsupported-media-type":
		return "https://github.com/gastownhall/bdp/problems/request", 415, "never", true
	case "binding-unavailable":
		return "https://github.com/gastownhall/bdp/problems/request", 400, "never", true
	case "validation-failed":
		return "https://github.com/gastownhall/bdp/problems/validation", 422, "never", true
	case "type-not-installed":
		return "https://github.com/gastownhall/bdp/problems/validation", 422, "after-state-change", true
	case "identity-taken":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "never", true
	case "alias-path-taken":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "revision-mismatch":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "incident-links-exist":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "aggregate-constraint-violation":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	case "idempotency-conflict":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "never", true
	case "idempotency-in-progress":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-delay", true
	case "idempotency-expired":
		return "https://github.com/gastownhall/bdp/problems/gone", 410, "never", true
	case "revision-allocation-unsafe":
		return "https://github.com/gastownhall/bdp/problems/conflict", 409, "after-state-change", true
	default:
		return "", 0, "", false
	}
}

func (r canonicalRetainedReply) acquisitionFailed() bool {
	return r.ExchangeErr != nil || r.ReadErr != nil || r.LimitErr != nil
}
func (r canonicalRetainedReply) Err() error {
	return errors.Join(r.OperationErr, r.ExchangeErr, r.ReadErr, r.LimitErr, r.ParseErr, r.CloseErr)
}

// The Scope landing is navigation, not a native Resource representation. Its
// human-readable 200 body is bounded but never interpreted as generic BDP JSON.
func canonicalRetainedScopeLanding(reply canonicalRetainedReply) error {
	if reply.Err() != nil {
		return reply.Err()
	}
	switch reply.Status {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		if len(reply.NativeBody) == 0 {
			return nil
		}
	}
	return fmt.Errorf("retained Scope landing refused")
}

func canonicalRetainedBoolean(raw []byte) (bool, error) {
	switch string(bytes.TrimSpace(raw)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("retained member must be boolean")
	}
}
func canonicalRetainedInteger(raw []byte, minimum int64) error {
	text := strings.TrimSpace(string(raw))
	if text == "" || strings.HasPrefix(text, "\"") {
		return fmt.Errorf("retained member must be an integer")
	}
	value, ok := new(big.Rat).SetString(text)
	if !ok || !value.IsInt() || value.Cmp(big.NewRat(minimum, 1)) < 0 {
		return fmt.Errorf("retained integer is outside its admitted range")
	}
	return nil
}
func canonicalRetainedJSONString(raw []byte) (string, error) {
	trimmed := bytes.TrimSpace(raw)
	var value string
	if len(trimmed) == 0 || trimmed[0] != '"' || json.Unmarshal(raw, &value) != nil {
		return "", fmt.Errorf("retained member must be a string")
	}
	return value, nil
}
func canonicalRetainedAbsoluteURI(raw []byte) error {
	value, err := canonicalRetainedString(raw)
	if err != nil {
		return err
	}
	parsed, err := url.Parse(value)
	if err != nil || !parsed.IsAbs() {
		return fmt.Errorf("retained URI must be absolute")
	}
	return nil
}
func canonicalRetainedJSONPointer(raw []byte) (string, error) {
	pointer, err := canonicalRetainedJSONString(raw)
	if err != nil {
		return "", err
	}
	if pointer != "" && !strings.HasPrefix(pointer, "/") {
		return "", fmt.Errorf("invalid retained JSON Pointer")
	}
	for i := 0; i < len(pointer); i++ {
		if pointer[i] == '~' {
			if i+1 >= len(pointer) || pointer[i+1] != '0' && pointer[i+1] != '1' {
				return "", fmt.Errorf("invalid retained JSON Pointer escape")
			}
			i++
		}
	}
	return pointer, nil
}
func canonicalRetainedArray(raw []byte) ([]json.RawMessage, error) {
	var items []json.RawMessage
	if len(bytes.TrimSpace(raw)) == 0 || bytes.TrimSpace(raw)[0] != '[' || json.Unmarshal(raw, &items) != nil {
		return nil, fmt.Errorf("retained member must be an array")
	}
	return items, nil
}

func parseCanonicalRetainedDiscovery(raw []byte) (canonicalRetainedDiscovery, error) {
	fields, err := canonicalRetainedObject(raw, "bdpVersion", "profile", "scope", "beads", "links", "events", "aliases", "operations", "order", "limits", "maximumEndpointMultiplicity", "historicalResolution", "capabilities")
	if err != nil {
		return canonicalRetainedDiscovery{}, err
	}
	version, err := canonicalRetainedString(fields["bdpVersion"])
	if err != nil || version != "0.1" {
		return canonicalRetainedDiscovery{}, fmt.Errorf("unsupported retained contract")
	}
	profile, err := canonicalRetainedString(fields["profile"])
	if err != nil || profile != "core" && profile != "write" {
		return canonicalRetainedDiscovery{}, fmt.Errorf("unsupported retained profile")
	}
	for _, name := range []string{"scope", "beads", "links", "events"} {
		value, err := canonicalRetainedString(fields[name])
		if err != nil {
			return canonicalRetainedDiscovery{}, err
		}
		if _, err := canonicalRetainedURL(value); err != nil {
			return canonicalRetainedDiscovery{}, err
		}
	}
	for _, name := range []string{"aliases", "operations"} {
		value, exists := fields[name]
		if name == "operations" && profile == "core" && exists {
			return canonicalRetainedDiscovery{}, fmt.Errorf("Core discovery has Write operations")
		}
		if profile == "write" || exists {
			text, err := canonicalRetainedString(value)
			if err != nil {
				return canonicalRetainedDiscovery{}, err
			}
			if _, err := canonicalRetainedURL(text); err != nil {
				return canonicalRetainedDiscovery{}, err
			}
		}
	}
	if value, ok := fields["order"]; ok {
		order, err := canonicalRetainedString(value)
		if err != nil || order != "canonical-uri" {
			return canonicalRetainedDiscovery{}, fmt.Errorf("unsupported discovery order")
		}
	}
	if value, ok := fields["limits"]; ok {
		if err := canonicalRetainedDiscoveryLimits(value, profile); err != nil {
			return canonicalRetainedDiscovery{}, err
		}
	}
	if value, ok := fields["maximumEndpointMultiplicity"]; ok {
		policies, err := canonicalRetainedArray(value)
		if err != nil {
			return canonicalRetainedDiscovery{}, err
		}
		for _, raw := range policies {
			policy, err := canonicalRetainedObject(raw, "linkConformsTo", "endpoint", "max")
			if err != nil {
				return canonicalRetainedDiscovery{}, err
			}
			typ, err := canonicalRetainedString(policy["linkConformsTo"])
			if err != nil {
				return canonicalRetainedDiscovery{}, err
			}
			if _, err := canonicalRetainedURL(typ); err != nil {
				return canonicalRetainedDiscovery{}, err
			}
			endpoint, err := canonicalRetainedString(policy["endpoint"])
			if err != nil || endpoint != "source" && endpoint != "target" {
				return canonicalRetainedDiscovery{}, fmt.Errorf("invalid endpoint multiplicity policy")
			}
			if err := canonicalRetainedInteger(policy["max"], 1); err != nil {
				return canonicalRetainedDiscovery{}, err
			}
		}
	}
	if value, ok := fields["historicalResolution"]; ok {
		history, err := canonicalRetainedObject(value, "version")
		if err != nil {
			return canonicalRetainedDiscovery{}, err
		}
		if err := canonicalRetainedInteger(history["version"], 1); err != nil {
			return canonicalRetainedDiscovery{}, err
		}
		version, _ := new(big.Rat).SetString(strings.TrimSpace(string(history["version"])))
		if version.Cmp(big.NewRat(1, 1)) != 0 {
			return canonicalRetainedDiscovery{}, fmt.Errorf("unsupported historical-resolution grammar")
		}
	}
	if value, ok := fields["capabilities"]; ok {
		if err := canonicalRetainedDiscoveryCapabilities(value, profile); err != nil {
			return canonicalRetainedDiscovery{}, err
		}
	}
	// Typed decoding happens only after exact case-sensitive member admission.
	var desc canonicalRetainedDiscovery
	if err := json.Unmarshal(raw, &desc); err != nil {
		return desc, err
	}
	return desc, nil
}
func canonicalRetainedDiscoveryLimits(raw []byte, profile string) error {
	allowed := []string{"page", "request", "resource", "selector", "retention"}
	if profile == "write" {
		allowed = append(allowed, "patch", "validation")
	}
	fields, err := canonicalRetainedObject(raw, allowed...)
	if err != nil {
		return err
	}
	groups := map[string][]string{"page": {"defaultItems", "maximumItems"}, "request": {"targetBytes", "bodyBytes"}, "resource": {"representationBytes", "propertiesBytes"}, "selector": {"bytes", "depth", "nodes"}, "patch": {"operations", "pathBytes", "pathDepth"}, "validation": {"diagnostics", "diagnosticBytes"}}
	for name, value := range fields {
		if name == "retention" {
			names := []string{"maximumSnapshotLifetime"}
			if profile == "write" {
				names = append(names, "idempotency")
			}
			retention, err := canonicalRetainedObject(value, names...)
			if err != nil {
				return err
			}
			for _, duration := range retention {
				text, err := canonicalRetainedString(duration)
				if err != nil || !strings.HasPrefix(text, "P") || len(text) < 2 {
					return fmt.Errorf("invalid retained duration")
				}
			}
		} else {
			limits, err := canonicalRetainedObject(value, groups[name]...)
			if err != nil {
				return err
			}
			for _, number := range limits {
				if err := canonicalRetainedInteger(number, 1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func canonicalRetainedDiscoveryCapabilities(raw []byte, profile string) error {
	capabilities, err := canonicalRetainedObject(raw)
	if err != nil {
		return err
	}
	for identity, rawVersions := range capabilities {
		encoded, _ := json.Marshal(identity)
		if err := canonicalRetainedAbsoluteURI(encoded); err != nil {
			return err
		}
		versions, err := canonicalRetainedObject(rawVersions)
		if err != nil || len(versions) == 0 {
			return fmt.Errorf("capability needs a nonempty version map")
		}
		for version, details := range versions {
			if version == "" {
				return fmt.Errorf("empty capability version")
			}
			if _, err := canonicalRetainedObject(details); err != nil {
				return err
			}
			// Unknown capability identities/versions retain their opaque detail object.
			if version != "1" {
				continue
			}
			switch identity {
			case "https://github.com/gastownhall/bdp/capabilities/atomic-batch":
				if profile == "core" {
					return fmt.Errorf("Core discovery cannot advertise atomic batch1")
				}
				fields, err := canonicalRetainedObject(details, "operation", "operations", "maximumOperations", "maximumResultBytes")
				if err != nil {
					return err
				}
				target, err := canonicalRetainedString(fields["operation"])
				if err != nil {
					return err
				}
				if _, err := canonicalRetainedURL(target); err != nil {
					return err
				}
				operations, err := canonicalRetainedArray(fields["operations"])
				if err != nil || len(operations) == 0 {
					return fmt.Errorf("atomic batch advertisement needs operations")
				}
				seen := map[string]bool{}
				for _, raw := range operations {
					name, err := canonicalRetainedString(raw)
					if err != nil || seen[name] {
						return fmt.Errorf("invalid advertised batch operation")
					}
					switch name {
					case "createBead", "updateBead", "deleteBead", "createLink", "updateLink", "deleteLink":
					default:
						return fmt.Errorf("unsupported advertised batch operation")
					}
					seen[name] = true
				}
				for _, name := range []string{"maximumOperations", "maximumResultBytes"} {
					if raw, ok := fields[name]; ok {
						if err := canonicalRetainedInteger(raw, 1); err != nil {
							return err
						}
					}
				}
			case "https://github.com/gastownhall/bdp/capabilities/type-definition-read":
				fields, err := canonicalRetainedObject(details, "operation")
				if err != nil {
					return err
				}
				target, err := canonicalRetainedString(fields["operation"])
				if err != nil {
					return err
				}
				if _, err := canonicalRetainedURL(target); err != nil {
					return err
				}
			case "https://github.com/gastownhall/bdp/capabilities/resource-projection":
				fields, err := canonicalRetainedObject(details, "current", "exact", "list")
				if err != nil {
					return err
				}
				any := false
				for _, name := range []string{"current", "exact", "list"} {
					value, err := canonicalRetainedBoolean(fields[name])
					if err != nil {
						return err
					}
					any = any || value
				}
				if !any {
					return fmt.Errorf("empty projection capability")
				}
			}
		}
	}
	return nil
}
