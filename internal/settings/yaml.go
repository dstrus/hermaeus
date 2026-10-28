package settings

import (
	"bytes"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// A settings file is changed in place: each changed value is swapped for its
// new text on the line where it sits, in the style it had (plain, quoted,
// literal, or folded), so comments, blank lines, and line endings stay
// exactly as they were. That handles every change that only replaces values,
// which is what the settings pages do. When a change adds a setting, the
// file is written as YAML the way SkyrimNet writes it, and only the lines
// that change replace the matching lines of the original (a three-way
// merge, as version control does).

// document is a parsed settings file, with what the parser can't keep but
// the file as written needs back: its line endings, whether it ends in a
// newline, and which empty lists or maps sat on their own line.
type document struct {
	root, doc     *yaml.Node
	ownLine       map[*yaml.Node]bool
	crlf, finalNL bool
}

func parse(b []byte) (*document, error) {
	d := &document{root: &yaml.Node{}, ownLine: map[*yaml.Node]bool{}}
	if err := yaml.Unmarshal(b, d.root); err != nil {
		return nil, err
	}
	if len(d.root.Content) == 0 {
		d.root = &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	d.doc = d.root.Content[0]
	d.crlf = bytes.Contains(b, []byte("\r\n"))
	d.finalNL = bytes.HasSuffix(b, []byte("\n"))
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.MappingNode {
			for i := 0; i+1 < len(n.Content); i += 2 {
				k, v := n.Content[i], n.Content[i+1]
				if isEmptyCollection(v) {
					d.ownLine[v] = v.Line > k.Line
				}
			}
		}
		if n.Style&yaml.FlowStyle == 0 {
			for _, c := range n.Content {
				walk(c)
			}
		}
	}
	walk(d.doc)
	return d, nil
}

func isEmptyCollection(n *yaml.Node) bool {
	return (n.Kind == yaml.SequenceNode || n.Kind == yaml.MappingNode) && len(n.Content) == 0
}

// emptyValueRe matches "key: []" and "key: {}" lines as the encoder writes them.
var emptyValueRe = regexp.MustCompile(`(?m)^( *)([^\s\-#][^\n]*?): (\[\]|\{\})$`)

// encode writes the document the way SkyrimNet writes its settings: 2-space
// indents, the file's own line endings, no final newline unless it had one,
// and each empty list or map ([] or {}) on the key's line or its own, as it
// was. With those, a file read and written unchanged comes back byte for
// byte.
func (d *document) encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(d.root); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	// The encoder writes every empty collection on its key's line, in document
	// order; move back the ones that sat on their own line.
	var own []bool
	var walk func(n *yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.MappingNode {
			for i := 1; i < len(n.Content); i += 2 {
				if isEmptyCollection(n.Content[i]) {
					own = append(own, d.ownLine[n.Content[i]])
				}
			}
		}
		if n.Style&yaml.FlowStyle == 0 {
			for i, c := range n.Content {
				if n.Kind == yaml.MappingNode && i%2 == 0 {
					continue
				}
				walk(c)
			}
		}
	}
	walk(d.doc)
	out := buf.Bytes()
	if matches := emptyValueRe.FindAllSubmatchIndex(out, -1); len(matches) == len(own) {
		var b bytes.Buffer
		last := 0
		for i, m := range matches {
			if !own[i] {
				continue
			}
			indent := out[m[2]:m[3]]
			b.Write(out[last:m[0]])
			b.Write(indent)
			b.Write(out[m[4]:m[5]])
			b.WriteString(":\n")
			b.Write(indent)
			b.WriteString("  ")
			b.Write(out[m[6]:m[7]])
			last = m[1]
		}
		b.Write(out[last:])
		out = b.Bytes()
	}
	if d.crlf {
		out = bytes.ReplaceAll(out, []byte("\n"), []byte("\r\n"))
	}
	if !d.finalNL {
		out = bytes.TrimRight(out, "\r\n")
	}
	return out, nil
}

// findNode looks up a dotted key such as narration.enabled.
func findNode(n *yaml.Node, key string) *yaml.Node {
	for _, part := range strings.Split(key, ".") {
		if n == nil || n.Kind != yaml.MappingNode {
			return nil
		}
		var next *yaml.Node
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Value == part {
				next = n.Content[i+1]
			}
		}
		n = next
	}
	return n
}

func scalar(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

// ---------- changed values only ----------

// patchValues returns original with each value that differs between
// untouched and edited (the file as parsed, and as parsed and changed)
// replaced in place. ok is false when the change adds, removes, or moves a
// setting, or a value can't be found where the parser says it is; the
// caller then falls back to keepFormatting.
func patchValues(original string, untouched, edited *yaml.Node) (string, bool) {
	type change struct {
		old, new *yaml.Node
		indent   int // the indentation that content of a block value must exceed
	}
	var changes []change
	var walk func(a, b *yaml.Node, indent int) bool
	walk = func(a, b *yaml.Node, indent int) bool {
		if a.Kind != b.Kind || len(a.Content) != len(b.Content) {
			return false
		}
		switch a.Kind {
		case yaml.MappingNode:
			for i := 0; i+1 < len(a.Content); i += 2 {
				if a.Content[i].Value != b.Content[i].Value || !walk(a.Content[i+1], b.Content[i+1], a.Content[i].Column-1) {
					return false
				}
			}
		case yaml.SequenceNode:
			for i := range a.Content {
				if !walk(a.Content[i], b.Content[i], a.Content[i].Column-3) {
					return false
				}
			}
		case yaml.ScalarNode:
			if a.Value != b.Value || a.ShortTag() != b.ShortTag() {
				changes = append(changes, change{a, b, indent})
			}
		default:
			return a.Value == b.Value
		}
		return true
	}
	if !walk(untouched, edited, -1) {
		return "", false
	}
	ls := lines(original)
	// From the last change to the first, so earlier positions stay valid.
	sort.Slice(changes, func(i, j int) bool {
		a, b := changes[i].old, changes[j].old
		return a.Line > b.Line || (a.Line == b.Line && a.Column > b.Column)
	})
	for _, c := range changes {
		var ok bool
		if ls, ok = replaceValue(ls, c.old, c.new, c.indent); !ok {
			return "", false
		}
	}
	return strings.Join(ls, ""), true
}

// replaceValue swaps one scalar's text in the file's lines.
func replaceValue(ls []string, old, new *yaml.Node, indent int) ([]string, bool) {
	row, col := old.Line-1, old.Column-1
	if row < 0 || row >= len(ls) || col > len(ls[row]) {
		return nil, false
	}
	line := ls[row]
	eol := strings.TrimRight(line, "\r\n")
	ending := line[len(eol):]
	n := *new
	n.Style = old.Style
	n.HeadComment, n.LineComment, n.FootComment = "", "", "" // the original's stay where they are
	switch old.Style {
	case yaml.LiteralStyle, yaml.FoldedStyle:
		// The header (| or >, with any indicators) ends the line; the content
		// is the following lines indented more than the field, up to the last
		// non-blank one.
		end := row + 1
		contentIndent := -1
		for i := row + 1; i < len(ls); i++ {
			t := strings.TrimRight(ls[i], "\r\n")
			if strings.TrimSpace(t) == "" {
				continue
			}
			ind := len(t) - len(strings.TrimLeft(t, " "))
			if ind <= indent {
				break
			}
			if contentIndent < 0 {
				contentIndent = ind
			}
			end = i + 1
		}
		if contentIndent < 0 {
			return nil, false
		}
		width := 0
		for _, l := range ls[row+1 : end] {
			width = max(width, len(strings.TrimRight(l, "\r\n")))
		}
		header := eol[:col] + string(eol[col])
		if old.Style == yaml.FoldedStyle {
			header = eol[:col] + ">"
		}
		body, chomp, ok := blockBody(n.Value, old.Style == yaml.FoldedStyle, strings.Repeat(" ", contentIndent), max(width, 60), ending)
		if !ok {
			return nil, false
		}
		out := append([]string{}, ls[:row]...)
		out = append(out, header+chomp+ending)
		out = append(out, body...)
		return append(out, ls[end:]...), true
	}
	// A value on one line: plain, or quoted.
	rest := eol[col:]
	length := 0
	switch old.Style {
	case yaml.DoubleQuotedStyle:
		length = closingQuote(rest, '"')
	case yaml.SingleQuotedStyle:
		length = closingQuote(rest, '\'')
	default:
		length = len(rest)
		if i := strings.Index(rest, " #"); i >= 0 {
			length = i
		}
		length = len(strings.TrimRight(rest[:length], " "))
	}
	if length <= 0 || strings.Contains(n.Value, "\n") {
		return nil, false
	}
	text, ok := inlineScalar(&n)
	if !ok {
		return nil, false
	}
	out := append([]string{}, ls...)
	out[row] = eol[:col] + text + rest[length:] + ending
	return out, true
}

// closingQuote returns the length of a quoted scalar at the start of s,
// quotes included, or 0 if it doesn't close on this line.
func closingQuote(s string, q byte) int {
	for i := 1; i < len(s); i++ {
		switch {
		case q == '"' && s[i] == '\\':
			i++
		case s[i] == q && q == '\'' && i+1 < len(s) && s[i+1] == '\'':
			i++
		case s[i] == q:
			return i + 1
		}
	}
	return 0
}

// inlineScalar writes a scalar as it would appear after "key: ", keeping its
// style if the value allows it.
func inlineScalar(n *yaml.Node) (string, bool) {
	out, err := yaml.Marshal(&yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{scalar("k"), n}})
	if err != nil {
		return "", false
	}
	s := strings.TrimSuffix(strings.TrimPrefix(string(out), "k: "), "\n")
	return s, !strings.Contains(s, "\n")
}

// blockBody writes a literal or folded value's lines at the given indent,
// folded paragraphs wrapped at width, with the chomping indicator its ending
// needs. ok is false for a folded value with lines starting with a space,
// which folding treats specially.
func blockBody(value string, folded bool, indent string, width int, ending string) (body []string, chomp string, ok bool) {
	text := strings.TrimRight(value, "\n")
	switch trailing := len(value) - len(text); {
	case trailing == 0:
		chomp = "-"
	case trailing > 1:
		chomp = "+"
	}
	for i, l := range strings.Split(text, "\n") {
		if !folded {
			if l == "" {
				body = append(body, ending)
			} else {
				body = append(body, indent+l+ending)
			}
			continue
		}
		// Folded: a blank line stands for each line break in the value, and
		// one line break between words reads as a space, so long lines wrap.
		if strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") {
			return nil, "", false
		}
		if i > 0 {
			body = append(body, ending)
		}
		if l == "" {
			continue
		}
		if strings.Contains(l, "  ") { // wrapping could start a line with a space, which reads differently
			body = append(body, indent+l+ending)
			continue
		}
		cur := indent
		for _, w := range strings.Split(l, " ") {
			if cur != indent && len(cur)+1+len(w) > width {
				body = append(body, cur+ending)
				cur = indent
			}
			if cur != indent {
				cur += " "
			}
			cur += w
		}
		body = append(body, cur+ending)
	}
	if chomp == "+" {
		for i := 1; i < len(value)-len(text); i++ {
			body = append(body, ending)
		}
	}
	return body, chomp, true
}

// ---------- structural changes ----------

// keepFormatting returns the edited file: original with the change from
// before to after (the encoder's output without and with the change). If
// the merged text doesn't mean exactly what after means, it returns after.
func keepFormatting(original, before, after string) string {
	if before == after {
		return original
	}
	merged := merge3(lines(original), lines(before), lines(after))
	if sameYAML(merged, after) {
		return merged
	}
	return after
}

func lines(s string) []string { return strings.SplitAfter(s, "\n") }

func sameYAML(a, b string) bool {
	var x, y any
	if yaml.Unmarshal([]byte(a), &x) != nil || yaml.Unmarshal([]byte(b), &y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}

// merge3 applies the changes from base to ours onto theirs, where theirs is
// base with formatting differences. Where both changed the same lines, ours
// wins.
func merge3(theirs, base, ours []string) string {
	toTheirs := matchLines(base, theirs) // base index -> theirs index, or -1
	toOurs := matchLines(base, ours)
	var b strings.Builder
	ti, oi, bi := 0, 0, 0
	for {
		// The next base line that both sides kept.
		next := bi
		for next < len(base) && (toTheirs[next] < ti || toOurs[next] < oi) {
			next++
		}
		tEnd, oEnd := len(theirs), len(ours)
		if next < len(base) {
			tEnd, oEnd = toTheirs[next], toOurs[next]
		}
		baseChunk, theirChunk, ourChunk := base[bi:next], theirs[ti:tEnd], ours[oi:oEnd]
		if equalLines(ourChunk, baseChunk) {
			write(&b, theirChunk) // unchanged: keep the original's text
		} else {
			// Changed (or both changed): the edit wins, between whatever the
			// original had around those lines.
			lead, trail := extras(baseChunk, theirChunk)
			write(&b, lead)
			write(&b, ourChunk)
			write(&b, trail)
		}
		if next >= len(base) {
			break
		}
		b.WriteString(theirs[tEnd])
		ti, oi, bi = tEnd+1, oEnd+1, next+1
	}
	return b.String()
}

// extras returns the lines theirs has before and after base's lines, such
// as the blank line and comment above a changed setting, when every line of
// base survives in theirs. When theirs rewrote some of them, there is no
// telling which lines go where, and nothing is kept.
func extras(base, theirs []string) (lead, trail []string) {
	if len(base) == 0 {
		return theirs, nil
	}
	m := matchLines(base, theirs)
	for _, i := range m {
		if i < 0 {
			return nil, nil
		}
	}
	return theirs[:m[0]], theirs[m[len(m)-1]+1:]
}

func write(b *strings.Builder, ls []string) {
	for _, l := range ls {
		b.WriteString(l)
	}
}

func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// matchLines pairs lines of a with equal lines of b, in order, as many as
// possible (longest common subsequence). It returns, for each line of a, the
// index of its pair in b, or -1.
func matchLines(a, b []string) []int {
	n, m := len(a), len(b)
	// Lines both sides start or end with need no table.
	pre := 0
	for pre < n && pre < m && a[pre] == b[pre] {
		pre++
	}
	suf := 0
	for suf < n-pre && suf < m-pre && a[n-1-suf] == b[m-1-suf] {
		suf++
	}
	out := make([]int, n)
	for i := range out {
		out[i] = -1
	}
	for i := 0; i < pre; i++ {
		out[i] = i
	}
	for i := 0; i < suf; i++ {
		out[n-1-i] = m - 1 - i
	}
	a2, b2 := a[pre:n-suf], b[pre:m-suf]
	w := len(b2) + 1
	L := make([]int32, (len(a2)+1)*w)
	for i := len(a2) - 1; i >= 0; i-- {
		for j := len(b2) - 1; j >= 0; j-- {
			if a2[i] == b2[j] {
				L[i*w+j] = L[(i+1)*w+j+1] + 1
			} else if x, y := L[(i+1)*w+j], L[i*w+j+1]; x >= y {
				L[i*w+j] = x
			} else {
				L[i*w+j] = y
			}
		}
	}
	for i, j := 0, 0; i < len(a2) && j < len(b2); {
		switch {
		case a2[i] == b2[j]:
			out[pre+i] = pre + j
			i++
			j++
		case L[(i+1)*w+j] >= L[i*w+j+1]:
			i++
		default:
			j++
		}
	}
	return out
}
