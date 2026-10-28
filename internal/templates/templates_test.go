package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const bio = "{% block summary %}Serana is a vampire.{% endblock %}\n\n" +
	"{% block aspirations %}\n- Find her mother\n{% endblock %}\n```\n"

func TestSyntaxErrors(t *testing.T) {
	flat := Rules{FlatBlocks: true}
	for raw, want := range map[string]string{
		"{% block summary %}Hi":                                 "Line 1: block summary is never closed",
		"{% block summary %}\n{% block skills %}{% endblock %}": "Line 2: block skills starts inside block summary",
		"Hi{% endblock %}":                                      "endblock has no matching block",
		"{% block %}x{% endblock %}":                            "block has no name",
		"{% block a %}{{ name {% endblock %}":                   "{{ is never closed",
		"{% block a %}{% if x %}y{% endblock %}":                "endblock has no matching block",
		"{% if %}{% endif %}":                                   "if needs a condition",
		"{% else %}":                                            "else isn't inside an if",
		"{% if a %}{% else %}{% else if b %}{% endif %}":        "else if comes after the else",
		"{% set x %}":                                           "set needs a name and =",
		"{% raw %}{{ unclosed":                                  "raw is never closed",
		"{% endraw %}":                                          "endraw has no matching raw",
		"{% frobnicate x %}":                                    `unknown statement "frobnicate"`,
		"{% %}":                                                 "isn't a statement",
	} {
		err := Check(raw, flat)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("Check(%q) = %v, want %q", raw, err, want)
		}
	}
}

func TestValidTemplates(t *testing.T) {
	for _, raw := range []string{
		bio + "{# note #}{% if x %}{{ y }}{% endif %}",
		"{% for a in b %}{{ a }}{% endfor %}{% set n = 1 %}{% include \"x\" %}",
		"{% raw %}{% block unclosed {% endraw %}",
		"{% if a %}1{% else if b %}2{% else %}3{% endif %}",
		"plain text with a { brace and a lone }",
	} {
		if err := Check(raw, Rules{FlatBlocks: true}); err != nil {
			t.Errorf("Check(%q) = %v, want nil", raw, err)
		}
	}
	// Prompts may nest blocks; bios may not.
	nested := "{% block a %}{% block b %}{% endblock %}{% endblock %}"
	if err := Check(nested, Rules{}); err != nil {
		t.Errorf("nested blocks in a prompt: %v", err)
	}
	if err := Check(nested, Rules{FlatBlocks: true}); err == nil {
		t.Error("nested blocks in a bio passed")
	}
}

// Every bio in a real SkyrimNet folder should pass the check. Run with
// HERMAEUS_TEST_SKYRIMNET=/path/to/SkyrimNet go test ./internal/templates -run TestRealBios -v
func TestRealBiosPassSyntaxCheck(t *testing.T) {
	dir := os.Getenv("HERMAEUS_TEST_SKYRIMNET")
	if dir == "" {
		t.Skip("HERMAEUS_TEST_SKYRIMNET not set")
	}
	n := 0
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".prompt") || !strings.Contains(filepath.ToSlash(p), "/characters/") {
			return nil
		}
		n++
		b, _ := os.ReadFile(p)
		if err := Check(string(b), Rules{FlatBlocks: true}); err != nil {
			t.Errorf("%s: %v", p, err)
		}
		return nil
	})
	t.Logf("checked %d bios", n)
}
