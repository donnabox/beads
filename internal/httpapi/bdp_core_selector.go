package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Core's bounded Selector: an RFC 9535 filter expression over one candidate
// Resource, with singular paths, JSON scalar literals, existence tests, the six
// comparisons and &&, ||, !. Everything else is outside the profile and is
// invalid-parameter. It selects whole Resources, so evaluation is one boolean.

// errCoreSelectorLimit is a well-formed Selector past an advertised bound.
var errCoreSelectorLimit = errors.New("the selector exceeds an advertised selector limit")

type coreSelector struct{ root coreSelectorNode }

// matches reports whether candidate satisfies the Selector.
func (c *coreSelector) matches(candidate map[string]any) bool { return c.root.eval(candidate) }

type coreSelectorNode interface {
	eval(candidate map[string]any) bool
}

type (
	coreSelectorOr   struct{ left, right coreSelectorNode }
	coreSelectorAnd  struct{ left, right coreSelectorNode }
	coreSelectorNot  struct{ inner coreSelectorNode }
	coreSelectorTest struct{ path []any }
	coreSelectorCmp  struct {
		path    []any
		op      string
		literal any
		// flipped is a comparison written literal-first.
		flipped bool
	}
)

func (n coreSelectorOr) eval(c map[string]any) bool  { return n.left.eval(c) || n.right.eval(c) }
func (n coreSelectorAnd) eval(c map[string]any) bool { return n.left.eval(c) && n.right.eval(c) }
func (n coreSelectorNot) eval(c map[string]any) bool { return !n.inner.eval(c) }
func (n coreSelectorTest) eval(c map[string]any) bool {
	_, present := coreSelectorResolve(c, n.path)
	return present
}

func (n coreSelectorCmp) eval(c map[string]any) bool {
	value, present := coreSelectorResolve(c, n.path)
	left, leftPresent, right, rightPresent := value, present, n.literal, true
	if n.flipped {
		left, leftPresent, right, rightPresent = n.literal, true, value, present
	}
	switch n.op {
	case "==":
		return coreSelectorEqual(left, leftPresent, right, rightPresent)
	case "!=":
		return !coreSelectorEqual(left, leftPresent, right, rightPresent)
	case "<":
		return coreSelectorLess(left, leftPresent, right, rightPresent)
	case ">":
		return coreSelectorLess(right, rightPresent, left, leftPresent)
	case "<=":
		return coreSelectorLess(left, leftPresent, right, rightPresent) || coreSelectorEqual(left, leftPresent, right, rightPresent)
	default: // ">=", the only operator left after parsing.
		return coreSelectorLess(right, rightPresent, left, leftPresent) || coreSelectorEqual(left, leftPresent, right, rightPresent)
	}
}

func coreSelectorResolve(candidate map[string]any, path []any) (any, bool) {
	var current any = candidate
	for _, step := range path {
		switch step := step.(type) {
		case string:
			object, ok := current.(map[string]any)
			if !ok {
				return nil, false
			}
			next, present := object[step]
			if !present {
				return nil, false
			}
			current = next
		case int:
			array, ok := current.([]any)
			if !ok {
				return nil, false
			}
			index := step
			if index < 0 {
				index += len(array)
			}
			if index < 0 || index >= len(array) {
				return nil, false
			}
			current = array[index]
		}
	}
	return current, true
}

// coreSelectorEqual is RFC 9535's ==: two absent paths are equal, an absent path
// equals nothing else, and values compare by kind and content. A literal is a
// scalar, so an object or array is never equal to one.
func coreSelectorEqual(a any, aPresent bool, b any, bPresent bool) bool {
	if !aPresent || !bPresent {
		return aPresent == bPresent
	}
	switch av := a.(type) {
	case nil:
		return b == nil
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case json.Number:
		bv, ok := b.(json.Number)
		if !ok {
			return false
		}
		cmp, ok := coreSelectorCompareNumbers(av, bv)
		return ok && cmp == 0
	}
	return false
}

// coreSelectorLess is RFC 9535's <: numbers by exact value and strings by
// Unicode scalar order, which is UTF-8 byte order. Every other pair is false.
func coreSelectorLess(a any, aPresent bool, b any, bPresent bool) bool {
	if !aPresent || !bPresent {
		return false
	}
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av < bv
	case json.Number:
		bv, ok := b.(json.Number)
		if !ok {
			return false
		}
		cmp, ok := coreSelectorCompareNumbers(av, bv)
		return ok && cmp < 0
	}
	return false
}

func coreSelectorCompareNumbers(a, b json.Number) (int, bool) {
	left, ok := coreSelectorRat(a)
	if !ok {
		return 0, false
	}
	right, ok := coreSelectorRat(b)
	if !ok {
		return 0, false
	}
	return left.Cmp(right), true
}

// coreSelectorRat reads a number's exact decimal value. A stored literal outside
// binary64's range is not admissible JSON under Core, and ParseFloat refusing it
// is also what keeps an enormous exponent away from big.Rat.
func coreSelectorRat(n json.Number) (*big.Rat, bool) {
	if _, err := strconv.ParseFloat(string(n), 64); err != nil {
		return nil, false
	}
	return new(big.Rat).SetString(string(n))
}

// parseCoreSelector parses one Selector string against the advertised bounds.
func parseCoreSelector(text string) (*coreSelector, error) {
	if len(text) > coreSelectorBytes {
		return nil, errCoreSelectorLimit
	}
	if text == "" {
		return nil, errors.New("the selector is empty")
	}
	if !utf8.ValidString(text) {
		return nil, errors.New("the selector is not valid UTF-8")
	}
	p := &coreSelectorParser{text: text}
	p.space()
	if !p.literalText("$") {
		return nil, errors.New("a selector starts with $[?")
	}
	p.space()
	if !p.literalText("[") {
		return nil, errors.New("a selector starts with $[?")
	}
	p.space()
	if !p.literalText("?") {
		return nil, errors.New("a selector starts with $[?")
	}
	root, err := p.or(1)
	if err != nil {
		return nil, err
	}
	p.space()
	if !p.literalText("]") {
		return nil, p.unexpected()
	}
	p.space()
	if p.pos != len(p.text) {
		return nil, p.unexpected()
	}
	return &coreSelector{root: root}, nil
}

type coreSelectorParser struct {
	text  string
	pos   int
	nodes int
}

func (p *coreSelectorParser) node(depth int) error {
	p.nodes++
	if p.nodes > coreSelectorNodes || depth > coreSelectorDepth {
		return errCoreSelectorLimit
	}
	return nil
}

func (p *coreSelectorParser) unexpected() error {
	if p.pos >= len(p.text) {
		return errors.New("the selector ends early")
	}
	return fmt.Errorf("the selector has an unsupported construct at byte %d", p.pos)
}

func (p *coreSelectorParser) space() {
	for p.pos < len(p.text) && strings.IndexByte(" \t\r\n", p.text[p.pos]) >= 0 {
		p.pos++
	}
}

func (p *coreSelectorParser) literalText(want string) bool {
	if strings.HasPrefix(p.text[p.pos:], want) {
		p.pos += len(want)
		return true
	}
	return false
}

func (p *coreSelectorParser) or(depth int) (coreSelectorNode, error) {
	left, err := p.and(depth)
	if err != nil {
		return nil, err
	}
	for {
		p.space()
		if !p.literalText("||") {
			return left, nil
		}
		if err := p.node(depth); err != nil {
			return nil, err
		}
		right, err := p.and(depth)
		if err != nil {
			return nil, err
		}
		left = coreSelectorOr{left, right}
	}
}

func (p *coreSelectorParser) and(depth int) (coreSelectorNode, error) {
	left, err := p.basic(depth)
	if err != nil {
		return nil, err
	}
	for {
		p.space()
		if !p.literalText("&&") {
			return left, nil
		}
		if err := p.node(depth); err != nil {
			return nil, err
		}
		right, err := p.basic(depth)
		if err != nil {
			return nil, err
		}
		left = coreSelectorAnd{left, right}
	}
}

func (p *coreSelectorParser) basic(depth int) (coreSelectorNode, error) {
	if err := p.node(depth); err != nil {
		return nil, err
	}
	p.space()
	if p.pos < len(p.text) && p.text[p.pos] == '!' && !strings.HasPrefix(p.text[p.pos:], "!=") {
		p.pos++
		inner, err := p.basic(depth + 1)
		if err != nil {
			return nil, err
		}
		return coreSelectorNot{inner}, nil
	}
	if p.literalText("(") {
		inner, err := p.or(depth + 1)
		if err != nil {
			return nil, err
		}
		p.space()
		if !p.literalText(")") {
			return nil, p.unexpected()
		}
		return inner, nil
	}
	if p.pos < len(p.text) && p.text[p.pos] == '@' {
		path, err := p.path()
		if err != nil {
			return nil, err
		}
		op := p.operator()
		if op == "" {
			return coreSelectorTest{path}, nil
		}
		literal, err := p.scalar()
		if err != nil {
			return nil, err
		}
		return coreSelectorCmp{path: path, op: op, literal: literal}, nil
	}
	literal, err := p.scalar()
	if err != nil {
		return nil, err
	}
	op := p.operator()
	if op == "" {
		return nil, p.unexpected()
	}
	p.space()
	if p.pos >= len(p.text) || p.text[p.pos] != '@' {
		return nil, p.unexpected()
	}
	path, err := p.path()
	if err != nil {
		return nil, err
	}
	return coreSelectorCmp{path: path, op: op, literal: literal, flipped: true}, nil
}

func (p *coreSelectorParser) operator() string {
	p.space()
	for _, op := range []string{"==", "!=", "<=", ">=", "<", ">"} {
		if p.literalText(op) {
			return op
		}
	}
	return ""
}

// path parses a singular query relative to the candidate: @ then member names
// and array indexes, and no wildcard, slice, descent or filter.
func (p *coreSelectorParser) path() ([]any, error) {
	p.pos++ // the @
	var steps []any
	for p.pos < len(p.text) {
		switch p.text[p.pos] {
		case '.':
			p.pos++
			name := p.shorthandName()
			if name == "" {
				return nil, p.unexpected()
			}
			steps = append(steps, name)
		case '[':
			p.pos++
			p.space()
			step, err := p.bracketStep()
			if err != nil {
				return nil, err
			}
			steps = append(steps, step)
			p.space()
			if !p.literalText("]") {
				return nil, p.unexpected()
			}
		default:
			return steps, nil
		}
	}
	return steps, nil
}

func (p *coreSelectorParser) bracketStep() (any, error) {
	if p.pos < len(p.text) && (p.text[p.pos] == '"' || p.text[p.pos] == '\'') {
		return p.stringLiteral()
	}
	start := p.pos
	if p.pos < len(p.text) && p.text[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.text) && p.text[p.pos] >= '0' && p.text[p.pos] <= '9' {
		p.pos++
	}
	index, err := strconv.Atoi(p.text[start:p.pos])
	if err != nil {
		p.pos = start
		return nil, p.unexpected()
	}
	return index, nil
}

func (p *coreSelectorParser) shorthandName() string {
	start := p.pos
	for p.pos < len(p.text) {
		c, size := utf8.DecodeRuneInString(p.text[p.pos:])
		isName := c == '_' || c >= 0x80 || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(p.pos > start && c >= '0' && c <= '9')
		if !isName {
			break
		}
		p.pos += size
	}
	return p.text[start:p.pos]
}

// scalar parses a JSON scalar literal. An object or array is not an RFC 9535
// literal, so it is outside the profile here too.
func (p *coreSelectorParser) scalar() (any, error) {
	p.space()
	if p.pos >= len(p.text) {
		return nil, p.unexpected()
	}
	switch {
	case p.text[p.pos] == '"' || p.text[p.pos] == '\'':
		return p.stringLiteral()
	case p.literalText("true"):
		return true, nil
	case p.literalText("false"):
		return false, nil
	case p.literalText("null"):
		return nil, nil
	}
	start := p.pos
	for p.pos < len(p.text) && strings.IndexByte("+-.eE0123456789", p.text[p.pos]) >= 0 {
		p.pos++
	}
	number := p.text[start:p.pos]
	dec := json.NewDecoder(strings.NewReader(number))
	dec.UseNumber()
	var value any
	if number == "" || dec.Decode(&value) != nil || dec.More() {
		p.pos = start
		return nil, p.unexpected()
	}
	literal, ok := value.(json.Number)
	if !ok || !coreSelectorAdmissibleNumber(literal) {
		p.pos = start
		return nil, fmt.Errorf("the selector's number %q is not admissible under Core's number model", number)
	}
	return literal, nil
}

// coreSelectorAdmissibleNumber is Core's admission rule: the literal's exact
// value survives the round trip through the nearest binary64.
func coreSelectorAdmissibleNumber(n json.Number) bool {
	exact, ok := coreSelectorRat(n)
	if !ok {
		return false
	}
	float, err := strconv.ParseFloat(string(n), 64)
	if err != nil {
		return false
	}
	back, ok := new(big.Rat).SetString(strconv.FormatFloat(float, 'g', -1, 64))
	return ok && exact.Cmp(back) == 0
}

// stringLiteral parses an RFC 9535 string: double or single quoted, with JSON's
// escapes plus \' in the single-quoted form.
func (p *coreSelectorParser) stringLiteral() (string, error) {
	quote := p.text[p.pos]
	start := p.pos
	p.pos++
	var out strings.Builder
	for p.pos < len(p.text) {
		c := p.text[p.pos]
		switch {
		case c == quote:
			p.pos++
			return out.String(), nil
		case c < 0x20:
			p.pos = start
			return "", errors.New("the selector has a control character in a string")
		case c != '\\':
			out.WriteByte(c)
			p.pos++
		default:
			if err := p.stringEscape(&out, quote); err != nil {
				return "", err
			}
		}
	}
	return "", errors.New("the selector ends inside a string")
}

func (p *coreSelectorParser) stringEscape(out *strings.Builder, quote byte) error {
	p.pos++
	if p.pos >= len(p.text) {
		return errors.New("the selector ends inside a string")
	}
	c := p.text[p.pos]
	p.pos++
	switch c {
	case 'b':
		out.WriteByte('\b')
	case 'f':
		out.WriteByte('\f')
	case 'n':
		out.WriteByte('\n')
	case 'r':
		out.WriteByte('\r')
	case 't':
		out.WriteByte('\t')
	case '/', '\\':
		out.WriteByte(c)
	case quote:
		out.WriteByte(c)
	case 'u':
		r, err := p.hex4()
		if err != nil {
			return err
		}
		if r >= 0xDC00 && r <= 0xDFFF {
			return errors.New("the selector has an unpaired surrogate")
		}
		if r >= 0xD800 && r <= 0xDBFF {
			if !p.literalText(`\u`) {
				return errors.New("the selector has an unpaired surrogate")
			}
			low, err := p.hex4()
			if err != nil {
				return err
			}
			if low < 0xDC00 || low > 0xDFFF {
				return errors.New("the selector has an unpaired surrogate")
			}
			r = 0x10000 + (r-0xD800)<<10 + (low - 0xDC00)
		}
		out.WriteRune(r)
	default:
		return errors.New("the selector has an unknown string escape")
	}
	return nil
}

func (p *coreSelectorParser) hex4() (rune, error) {
	if p.pos+4 > len(p.text) {
		return 0, errors.New("the selector ends inside a string")
	}
	value, err := strconv.ParseUint(p.text[p.pos:p.pos+4], 16, 32)
	if err != nil {
		return 0, errors.New("the selector has a bad \\u escape")
	}
	p.pos += 4
	return rune(value), nil
}
