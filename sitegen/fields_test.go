package sitegen

import "testing"

func TestResolveNested(t *testing.T) {
	data := map[string]interface{}{
		"repo": map[string]interface{}{
			"languages": []interface{}{
				map[string]interface{}{"name": "Go"},
			},
		},
	}
	if got := Resolve(data, "repo.languages.0.name"); got != "Go" {
		t.Fatalf("expected Go, got %v", got)
	}
	if Resolve(data, "repo.missing") != nil {
		t.Fatalf("expected nil for missing field")
	}
}
