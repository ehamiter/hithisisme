package sitegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
)

// RenderOptions holds CLI options.
type RenderOptions struct {
	Input   string
	Out     string
	DataDir string
	Layout  string
}

// Render performs full render pipeline.
func Render(opts RenderOptions) error {
	in, err := os.ReadFile(opts.Input)
	if err != nil {
		return err
	}
	bindings, nodes, err := Parse(bytes.NewReader(in))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(opts.DataDir, 0o755); err != nil {
		return err
	}
	fetcher := NewFetcher(opts.DataDir)
	fetcher.LoadETags()
	ctx := &context{
		bindings: make(map[string]interface{}),
		lazy:     make(map[string]*lazyBinding),
		fetcher:  fetcher,
	}
	// resolve bindings
	for _, b := range bindings {
		if b.Glob {
			pattern := strings.Replace(b.Target, "**", "*.md", 1)
			posts, err := LoadPosts(pattern)
			if err != nil {
				return err
			}
			ctx.bindings[b.Name] = posts
			continue
		}
		if b.URL != "" && !b.Lazy {
			body, err := fetcher.Fetch(b.Target, b.URL)
			if err != nil {
				// if file exists use cache
				path := filepath.Join(opts.DataDir, b.Target)
				if body, err = os.ReadFile(path); err != nil {
					return err
				}
			}
			var v interface{}
			if err := json.Unmarshal(body, &v); err != nil {
				return err
			}
			ctx.bindings[b.Name] = v
			continue
		}
		if b.URL != "" && b.Lazy {
			lb := &lazyBinding{Target: b.Target, Template: b.URL, Data: make(map[string]interface{}), Fetched: make(map[string]bool)}
			// load cache
			path := filepath.Join(opts.DataDir, b.Target)
			if data, err := os.ReadFile(path); err == nil {
				json.Unmarshal(data, &lb.Data)
			}
			ctx.lazy[b.Name] = lb
			continue
		}
		if b.Manual {
			path := filepath.Join(opts.DataDir, b.Target)
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var v interface{}
			if err := json.Unmarshal(body, &v); err != nil {
				return err
			}
			ctx.bindings[b.Name] = v
		}
	}
	var buf strings.Builder
	if err := ctx.renderNodes(nodes, make(map[string]interface{}), &buf); err != nil {
		return err
	}
	// load layout
	layout, err := os.ReadFile(opts.Layout)
	if err != nil {
		return err
	}
	bodyHTML := buf.String()
	outHTML := strings.Replace(string(layout), "<!--CONTENT-->", bodyHTML, 1)
	// last updated: use now
	outHTML = strings.Replace(outHTML, "<!--LAST_UPDATED-->", time.Now().Format(time.RFC3339), 1)
	if err := os.MkdirAll(filepath.Dir(opts.Out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(opts.Out, []byte(outHTML), 0o644); err != nil {
		return err
	}
	fetcher.SaveETags()
	// save lazy caches
	for _, lb := range ctx.lazy {
		path := filepath.Join(opts.DataDir, lb.Target)
		b, _ := json.MarshalIndent(lb.Data, "", "  ")
		os.WriteFile(path, b, 0o644)
	}
	return nil
}

type context struct {
	bindings map[string]interface{}
	lazy     map[string]*lazyBinding
	fetcher  *Fetcher
}

type lazyBinding struct {
	Target   string
	Template string
	Data     map[string]interface{}
	Fetched  map[string]bool
}

func (c *context) renderNodes(nodes []Node, vars map[string]interface{}, buf *strings.Builder) error {
	for _, n := range nodes {
		switch t := n.(type) {
		case Section:
			md := goldmark.New()
			var h bytes.Buffer
			if err := md.Convert([]byte(t.Text), &h); err != nil {
				return err
			}
			buf.WriteString(fmt.Sprintf("<section id=\"%s\">%s</section>", t.ID, h.String()))
		case Field:
			v := c.resolvePath(t.Path, vars)
			buf.WriteString(fmt.Sprintf("<p>%s</p>", htmlEscape(fmt.Sprint(v))))
		case Loop:
			if err := c.renderLoop(t, vars, buf); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *context) renderLoop(l Loop, vars map[string]interface{}, buf *strings.Builder) error {
	src := c.resolvePath(l.Source, vars)
	switch arr := src.(type) {
	case []interface{}:
		items := append([]interface{}{}, arr...)
		SortSlice(items, l.Sort)
		buf.WriteString("<section>")
		for _, it := range items {
			nv := map[string]interface{}{}
			if len(l.Vars) > 0 {
				nv[l.Vars[0]] = it
			}
			buf.WriteString("<article class=\"item\">")
			if err := c.renderNodes(l.Body, merge(vars, nv), buf); err != nil {
				return err
			}
			buf.WriteString("</article>")
		}
		buf.WriteString("</section>")
	case map[string]interface{}:
		keys := make([]interface{}, 0, len(arr))
		for k, v := range arr {
			keys = append(keys, map[string]interface{}{"key": k, "value": v})
		}
		// for map, sort by provided sort keys applied on struct with key/value
		SortSlice(keys, l.Sort)
		buf.WriteString("<section>")
		for _, kv := range keys {
			m := kv.(map[string]interface{})
			nv := map[string]interface{}{}
			if len(l.Vars) == 2 {
				nv[l.Vars[0]] = m["key"]
				nv[l.Vars[1]] = m["value"]
			} else if len(l.Vars) == 1 {
				nv[l.Vars[0]] = m["value"]
			}
			buf.WriteString("<article class=\"item\">")
			if err := c.renderNodes(l.Body, merge(vars, nv), buf); err != nil {
				return err
			}
			buf.WriteString("</article>")
		}
		buf.WriteString("</section>")
	}
	return nil
}

func merge(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func (c *context) resolvePath(path string, vars map[string]interface{}) interface{} {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}
	head := parts[0]
	var obj interface{}
	if v, ok := vars[head]; ok {
		obj = v
	} else if v, ok := c.bindings[head]; ok {
		obj = v
	} else if lb, ok := c.lazy[head]; ok {
		obj = c.resolveLazy(lb, head, vars)
	}
	if obj == nil {
		return nil
	}
	if len(parts) == 1 {
		return obj
	}
	return Resolve(obj, strings.Join(parts[1:], "."))
}

var tmplRe = regexp.MustCompile(`\{([^}]+)\}`)

func (c *context) resolveLazy(lb *lazyBinding, name string, vars map[string]interface{}) interface{} {
	// need key from template
	url := lb.Template
	var key string
	matches := tmplRe.FindAllStringSubmatch(lb.Template, -1)
	for _, m := range matches {
		expr := m[1]
		parts := strings.SplitN(expr, ".", 2)
		if len(parts) != 2 {
			continue
		}
		v := vars[parts[0]]
		if rest := parts[1]; rest != "" {
			v = Resolve(v, rest)
		}
		s := fmt.Sprintf("%v", v)
		url = strings.ReplaceAll(url, m[0], s)
		if key == "" {
			key = s
		}
	}
	if key == "" {
		key = url
	}
	if val, ok := lb.Data[key]; ok {
		return val
	}
	if lb.Fetched[key] {
		return nil
	}
	lb.Fetched[key] = true
	body, err := c.fetcher.Fetch(lb.Target, url)
	if err != nil {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return nil
	}
	lb.Data[key] = v
	return v
}

func htmlEscape(s string) string {
	var buf bytes.Buffer
	for _, r := range s {
		switch r {
		case '&':
			buf.WriteString("&amp;")
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '"':
			buf.WriteString("&quot;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}
