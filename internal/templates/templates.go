// Package templates checks SkyrimNet's prompts and bios, which are Inja
// templates: {{ expressions }}, {% statements %} and {# comments #}. It
// catches the mistakes that would break a template outright: unclosed
// delimiters, unknown or misplaced statements, and if/for/block/raw sections
// that aren't closed in the right order. It doesn't evaluate expressions.
package templates

import (
	"fmt"
	"regexp"
	"strings"

	"hermaeus/internal/apperr"
)

// Rules says what a particular kind of template may do.
type Rules struct {
	// Bios are a flat list of {% block %} sections, so a block inside a
	// block is a mistake there; prompts may nest them.
	FlatBlocks bool
}

var (
	statementRe = regexp.MustCompile(`^\{%-?\s*(\w+)(?:\s+(if)\b)?\s*(.*?)\s*-?%\}$`)
	setRe       = regexp.MustCompile(`^[A-Za-z_][\w.]*\s*=`)
	nameRe      = regexp.MustCompile(`^\w+`)
	endRawRe    = regexp.MustCompile(`\{%-?\s*endraw\s*-?%\}`)
)

func bad(format string, a ...any) error { return apperr.New(apperr.Invalid, format, a...) }

// Check returns an error saying what's broken in content, or nil.
func Check(content string, rules Rules) error {
	line := func(i int) int { return strings.Count(content[:i], "\n") + 1 }
	type open struct {
		tag, name string
		line      int
		sawElse   bool
	}
	var stack []open
	describe := func(o open) string {
		if o.name != "" {
			return o.tag + " " + o.name
		}
		return o.tag
	}
	for i := 0; i < len(content); {
		start := strings.IndexByte(content[i:], '{')
		if start < 0 || i+start+1 >= len(content) {
			break
		}
		start += i
		var closer string
		switch content[start+1] {
		case '{':
			closer = "}}"
		case '%':
			closer = "%}"
		case '#':
			closer = "#}"
		default:
			i = start + 1
			continue
		}
		end := strings.Index(content[start+2:], closer)
		if end < 0 {
			return bad("Line %d: %s is never closed with %s", line(start), content[start:start+2], closer)
		}
		end += start + 2 + len(closer)
		i = end
		if closer != "%}" {
			continue
		}
		m := statementRe.FindStringSubmatch(content[start:end])
		if m == nil {
			return bad("Line %d: %s isn't a statement", line(start), strings.TrimSpace(content[start:end]))
		}
		tag, rest, at := m[1], m[3], line(start)
		if m[2] == "if" { // "else if"
			tag = "elif"
		}
		top := func() *open {
			if len(stack) == 0 {
				return nil
			}
			return &stack[len(stack)-1]
		}
		closeWith := func(want string) error {
			if t := top(); t == nil || t.tag != want {
				opened := ""
				if t != nil {
					opened = fmt.Sprintf(" (%s on line %d is still open)", describe(*t), t.line)
				}
				return bad("Line %d: end%s has no matching %s%s", at, want, want, opened)
			}
			stack = stack[:len(stack)-1]
			return nil
		}
		switch tag {
		case "block":
			name := nameRe.FindString(rest)
			if name == "" {
				return bad("Line %d: block has no name", at)
			}
			if rules.FlatBlocks {
				for _, o := range stack {
					if o.tag == "block" {
						return bad("Line %d: block %s starts inside block %s (line %d), which has no endblock before it", at, name, o.name, o.line)
					}
				}
			}
			stack = append(stack, open{tag: "block", name: name, line: at})
		case "if", "for":
			if rest == "" {
				return bad("Line %d: %s needs a condition", at, tag)
			}
			stack = append(stack, open{tag: tag, line: at})
		case "elif", "else":
			t := top()
			if t == nil || t.tag != "if" {
				return bad("Line %d: %s isn't inside an if", at, strings.Replace(tag, "elif", "else if", 1))
			}
			if t.sawElse {
				return bad("Line %d: %s comes after the else of the if on line %d", at, strings.Replace(tag, "elif", "else if", 1), t.line)
			}
			if tag == "else" {
				t.sawElse = true
			}
		case "endblock", "endif", "endfor":
			if err := closeWith(strings.TrimPrefix(tag, "end")); err != nil {
				return err
			}
		case "set":
			if !setRe.MatchString(rest) {
				return bad("Line %d: set needs a name and =, as in {%% set name = value %%}", at)
			}
		case "extends", "include":
		case "raw":
			// Everything up to endraw is plain text.
			endRaw := endRawRe.FindStringIndex(content[end:])
			if endRaw == nil {
				return bad("Line %d: raw is never closed with endraw", at)
			}
			i = end + endRaw[1]
		case "endraw":
			return bad("Line %d: endraw has no matching raw", at)
		default:
			return bad("Line %d: unknown statement %q", at, tag)
		}
	}
	if len(stack) > 0 {
		o := stack[len(stack)-1]
		return bad("Line %d: %s is never closed with end%s", o.line, describe(o), o.tag)
	}
	return nil
}

var referenceRe = regexp.MustCompile(`\b(render_template|render_subcomponent)\s*\(\s*"([^"]+)"`)

// Warnings lists the templates content refers to by a literal name that
// don't exist: render_template("components\\event_history") needs
// components/event_history.prompt, and render_subcomponent("x") needs
// submodules/x. exists reports whether a path under the prompts folder is
// there in any layer. These are warnings, not errors, since SkyrimNet may
// fill the name in a way this check can't see.
func Warnings(content string, exists func(path string) bool) []string {
	var out []string
	seen := map[string]bool{}
	for _, m := range referenceRe.FindAllStringSubmatch(content, -1) {
		name := strings.ReplaceAll(m[2], `\\`, "/")
		name = strings.ReplaceAll(name, `\`, "/")
		path := "submodules/" + name
		if m[1] == "render_template" {
			path = name + ".prompt"
		}
		if seen[path] || exists(path) {
			continue
		}
		seen[path] = true
		out = append(out, fmt.Sprintf("%s(%q) refers to %s, which doesn't exist", m[1], m[2], path))
	}
	return out
}
