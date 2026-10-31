// Package filters reads and changes SkyrimNet's filter lists: which
// factions, races, and genders may (or may not) form memories, or speak.
// Each list is a file in a content layer's filters/ folder, with a kind
// (memory for the memory filter, actor for the actor filter); a list in a
// higher layer replaces one with the same id below it, and every list of a
// kind adds its entries to the six whitelists and blacklists SkyrimNet uses.
// The player's lists, and their copies of other lists, are in overlay/.
package filters

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/store"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// Fields are the six lists a filter list can add to, in SkyrimNet's order.
var Fields = []string{"FactionWhitelist", "FactionBlacklist", "RaceWhitelist", "RaceBlacklist", "GenderWhitelist", "GenderBlacklist"}

var kinds = map[string]bool{"memory": true, "actor": true}

type Filters struct {
	store   *store.Store
	library *layers.Library
}

func New(s *store.Store, l *layers.Library) *Filters {
	return &Filters{store: s, library: l}
}

// List is one filter list, as the layer it comes from has it.
type List struct {
	ID          string              `json:"id"` // the file's name, without .yaml
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Entries     map[string][]string `json:"entries"` // by field, in Fields
	Layer       string              `json:"layer"`   // the layer's id
	LayerKind   string              `json:"layerKind"`
	Source      string              `json:"source"`             // the layer, for people
	Replaces    string              `json:"replaces,omitempty"` // the source of the list this copy replaces
	Modified    int64               `json:"modified"`
	// Enabled is whether SkyrimNet uses the list: its switch in
	// content-registry.json, or else its file's enabled:. A list switched
	// off adds nothing.
	Enabled bool `json:"enabled"`
}

// Source is one list an entry in effect comes from.
type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Entry is one value in effect in a field, with the lists that add it.
type Entry struct {
	Value string   `json:"value"`
	From  []Source `json:"from"`
}

type ListsResponse struct {
	Kind     string             `json:"kind"`
	Lists    []List             `json:"lists"`    // those in effect, the player's first, then by name
	InEffect map[string][]Entry `json:"inEffect"` // by field
}

// ListEdit is a list as it should be.
type ListEdit struct {
	Modified    int64               `json:"modified"` // 0 for a new list, or a first copy of another layer's
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Entries     map[string][]string `json:"entries"`
}

type Removal struct {
	Modified int64 `json:"modified"`
}

var idRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Lists reads every list of a kind, the one in effect for each id.
func (f *Filters) Lists(kind string) (ListsResponse, error) {
	if !kinds[kind] {
		return ListsResponse{}, apperr.New(apperr.Invalid, "There's no %q filter", kind)
	}
	out := ListsResponse{Kind: kind, Lists: []List{}, InEffect: map[string][]Entry{}}
	seen := map[string]int{} // folded id → index in out.Lists
	acts := f.library.Activations()
	for _, l := range f.library.Layers() {
		dir := layers.Lookup(l.Root, "filters")
		if dir == "" {
			continue
		}
		files, _ := os.ReadDir(dir)
		for _, e := range files {
			name := e.Name()
			if e.IsDir() || filepath.Ext(name) != ".yaml" { // SkyrimNet skips .YAML
				continue
			}
			list, ok := readList(filepath.Join(dir, name), kind)
			if !ok {
				continue
			}
			key := strings.ToLower(list.ID)
			if i, have := seen[key]; have {
				// A higher layer has it already; note what it replaces.
				if out.Lists[i].Replaces == "" {
					out.Lists[i].Replaces = l.Source
				}
				continue
			}
			list.Layer, list.LayerKind, list.Source = l.ID, l.Kind, l.Source
			list.Enabled = layers.Activation(acts[strings.ToLower("filters/"+name)], list.Enabled)
			seen[key] = len(out.Lists)
			out.Lists = append(out.Lists, list)
		}
	}
	sort.SliceStable(out.Lists, func(i, j int) bool {
		a, b := out.Lists[i], out.Lists[j]
		if (a.LayerKind == "overlay") != (b.LayerKind == "overlay") {
			return a.LayerKind == "overlay"
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
	for _, field := range Fields {
		var entries []Entry
		at := map[string]int{}
		for _, l := range out.Lists {
			if !l.Enabled {
				continue
			}
			for _, v := range l.Entries[field] {
				k := strings.ToLower(v)
				if i, ok := at[k]; ok {
					entries[i].From = append(entries[i].From, Source{ID: l.ID, Name: l.Name})
					continue
				}
				at[k] = len(entries)
				entries = append(entries, Entry{Value: v, From: []Source{{ID: l.ID, Name: l.Name}}})
			}
		}
		sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Value) < strings.ToLower(entries[j].Value) })
		if entries == nil {
			entries = []Entry{}
		}
		out.InEffect[field] = entries
	}
	return out, nil
}

// readList reads a list file, if it's a list of kind.
func readList(path, kind string) (List, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return List{}, false
	}
	var in map[string]any
	if yaml.Unmarshal(b, &in) != nil || in == nil {
		return List{}, false
	}
	if k, _ := in["kind"].(string); k != kind {
		return List{}, false
	}
	l := List{ID: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), Entries: map[string][]string{}}
	l.Name, _ = in["name"].(string)
	e, ok := in["enabled"].(bool)
	l.Enabled = !ok || e
	l.Description, _ = in["description"].(string)
	if l.Name == "" {
		l.Name = l.ID
	}
	for _, field := range Fields {
		l.Entries[field] = []string{}
		if vs, ok := in[field].([]any); ok {
			for _, v := range vs {
				if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
					l.Entries[field] = append(l.Entries[field], strings.TrimSpace(s))
				}
			}
		}
	}
	if info, err := os.Stat(path); err == nil {
		l.Modified = info.ModTime().UnixMilli()
	}
	return l, true
}

// overlayDir is the player's filters folder, and the layers.
func (f *Filters) overlayDir() (string, []layers.Layer, error) {
	lay := f.library.Layers()
	if len(lay) == 0 || lay[0].Kind != "overlay" {
		return "", nil, apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	return filepath.Join(lay[0].Root, "filters"), lay, nil
}

// Save writes the player's list: theirs, or their copy of another layer's,
// which then replaces it.
func (f *Filters) Save(kind, id string, e ListEdit) (ListsResponse, error) {
	if !kinds[kind] {
		return ListsResponse{}, apperr.New(apperr.Invalid, "There's no %q filter", kind)
	}
	if !idRe.MatchString(id) {
		return ListsResponse{}, apperr.New(apperr.Invalid, "Invalid list %q", id)
	}
	e.Name = strings.TrimSpace(e.Name)
	if e.Name == "" {
		return ListsResponse{}, apperr.New(apperr.Invalid, "A list needs a name")
	}
	err := f.store.ChangeFiles("Saved the "+kind+" filter list "+id, func() error {
		dir, _, err := f.overlayDir()
		if err != nil {
			return err
		}
		path := layers.Lookup(dir, id+".yaml")
		var old []byte
		if path != "" {
			info, err := os.Stat(path)
			if err != nil {
				return err
			}
			if info.ModTime().UnixMilli() != e.Modified {
				return apperr.New(apperr.Conflict, "The list changed after you opened it. Reload to see the latest, then make your change again.")
			}
			if old, err = os.ReadFile(path); err != nil {
				return err
			}
			if l, ok := readList(path, kind); !ok || !strings.EqualFold(l.ID, id) {
				return apperr.New(apperr.Conflict, "%s.yaml in your filters isn't a %s filter list", id, kind)
			}
		} else {
			path = filepath.Join(dir, id+".yaml")
		}
		// A first copy starts from the list it replaces, keeping its other
		// fields and its line endings.
		from := old
		if from == nil {
			from = f.below(id)
		}
		text, err := encodeList(from, kind, e)
		if err != nil {
			return err
		}
		if old != nil {
			if err := f.backup(path, old); err != nil {
				return err
			}
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		return textfile.WriteAtomic(path, []byte(text), 0o644)
	})
	if err != nil {
		return ListsResponse{}, err
	}
	return f.Lists(kind)
}

// Create adds a new list of the player's, named by an id made from its name.
func (f *Filters) Create(kind string, e ListEdit) (ListsResponse, List, error) {
	if !kinds[kind] {
		return ListsResponse{}, List{}, apperr.New(apperr.Invalid, "There's no %q filter", kind)
	}
	// Ids are shared across kinds and layers, so a new one must be free in
	// every layer, whatever its kind.
	taken := map[string]bool{}
	for _, l := range f.library.Layers() {
		if dir := layers.Lookup(l.Root, "filters"); dir != "" {
			files, _ := os.ReadDir(dir)
			for _, fe := range files {
				taken[strings.ToLower(strings.TrimSuffix(fe.Name(), filepath.Ext(fe.Name())))] = true
			}
		}
	}
	base := slug(e.Name) + "_" + kind
	id := base
	for n := 2; taken[strings.ToLower(id)]; n++ {
		id = base + "_" + itoa(n)
	}
	e.Modified = 0
	r, err := f.Save(kind, id, e)
	if err != nil {
		return ListsResponse{}, List{}, err
	}
	for _, l := range r.Lists {
		if l.ID == id {
			return r, l, nil
		}
	}
	return r, List{ID: id}, nil
}

// Remove deletes the player's list, or their copy of another layer's, so
// the original applies again. A backup is kept.
func (f *Filters) Remove(kind, id string, r Removal) (ListsResponse, error) {
	if !idRe.MatchString(id) {
		return ListsResponse{}, apperr.New(apperr.Invalid, "Invalid list %q", id)
	}
	err := f.store.ChangeFiles("Removed the "+kind+" filter list "+id, func() error {
		dir, _, err := f.overlayDir()
		if err != nil {
			return err
		}
		path := layers.Lookup(dir, id+".yaml")
		if path == "" {
			return apperr.New(apperr.NotFound, "%s isn't one of your lists; a built-in or plugin list can be replaced by your own copy, but not removed here", id)
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != r.Modified {
			return apperr.New(apperr.Conflict, "The list changed after you opened it. Reload to see the latest, then try again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := f.backup(path, old); err != nil {
			return err
		}
		return textfile.Remove(path, filepath.Dir(dir))
	})
	if err != nil {
		return ListsResponse{}, err
	}
	return f.Lists(kind)
}

func (f *Filters) backup(path string, old []byte) error {
	rel, err := filepath.Rel(f.store.Dir(), path)
	if err != nil {
		rel = filepath.Base(path)
	}
	if _, err := textfile.Backup(filepath.Join(f.store.BackupDir(), rel), old, 0o644); err != nil {
		return apperr.New(apperr.Conflict, "Could not back up %s, so nothing was changed: %v", filepath.Base(path), err)
	}
	return nil
}

// encodeList writes a list: kind, name, description, then the fields with
// entries, in SkyrimNet's order. Keys this page doesn't manage are kept.
// below is the list a first copy of yours replaces: the file of that id in
// the highest layer under your edits, or nil when there's none.
func (f *Filters) below(id string) []byte {
	for _, l := range f.library.Layers() {
		if l.Kind == "overlay" {
			continue
		}
		if dir := layers.Lookup(l.Root, "filters"); dir != "" {
			if path := layers.Lookup(dir, id+".yaml"); path != "" {
				if b, err := os.ReadFile(path); err == nil {
					return b
				}
			}
		}
	}
	return nil
}

func encodeList(old []byte, kind string, e ListEdit) (string, error) {
	doc := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	if old != nil {
		var root yaml.Node
		if err := yaml.Unmarshal(old, &root); err == nil && len(root.Content) == 1 && root.Content[0].Kind == yaml.MappingNode {
			doc = root.Content[0]
		}
	}
	managed := map[string]bool{"kind": true, "name": true, "description": true}
	for _, f := range Fields {
		managed[f] = true
	}
	var extra []*yaml.Node
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if !managed[doc.Content[i].Value] {
			extra = append(extra, doc.Content[i], doc.Content[i+1])
		}
	}
	str := func(s string) *yaml.Node {
		n := &yaml.Node{}
		n.SetString(s)
		return n
	}
	out := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	add := func(k string, v *yaml.Node) { out.Content = append(out.Content, str(k), v) }
	add("kind", str(kind))
	add("name", str(e.Name))
	if d := strings.TrimSpace(e.Description); d != "" {
		add("description", str(d))
	}
	for _, field := range Fields {
		var vs []string
		seen := map[string]bool{}
		for _, v := range e.Entries[field] {
			v = strings.TrimSpace(v)
			if v != "" && !seen[strings.ToLower(v)] {
				seen[strings.ToLower(v)] = true
				vs = append(vs, v)
			}
		}
		if len(vs) == 0 {
			continue
		}
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, v := range vs {
			seq.Content = append(seq.Content, str(v))
		}
		add(field, seq)
	}
	out.Content = append(out.Content, extra...)
	var b strings.Builder
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(out); err != nil {
		return "", err
	}
	enc.Close()
	text := b.String()
	if old != nil {
		text = textfile.MatchLineEndings(text, string(old))
	}
	return text, nil
}

// slug makes an id from a name: "No Bandits" → no_bandits.
func slug(name string) string {
	var b strings.Builder
	last := '_'
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			last = r
		} else if last != '_' {
			b.WriteRune('_')
			last = '_'
		}
	}
	if s := strings.Trim(b.String(), "_"); s != "" {
		return s
	}
	return "list"
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
