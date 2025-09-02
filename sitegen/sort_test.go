package sitegen

import "testing"

func TestParseSort(t *testing.T) {
	keys := ParseSort("stargazers_count, updated_at^")
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0].Path != "stargazers_count" || keys[0].Asc {
		t.Fatalf("first key wrong: %+v", keys[0])
	}
	if keys[1].Path != "updated_at" || !keys[1].Asc {
		t.Fatalf("second key wrong: %+v", keys[1])
	}
}

func TestSortSlice(t *testing.T) {
	items := []interface{}{
		map[string]interface{}{"a": 1, "b": 2},
		map[string]interface{}{"a": 2, "b": 1},
		map[string]interface{}{"b": 3}, // missing a
	}
	SortSlice(items, ParseSort("a^"))
	if Resolve(items[0], "a").(int) != 1 {
		t.Fatalf("expected first item a=1")
	}
	if Resolve(items[2], "a") != nil {
		t.Fatalf("expected last item missing a")
	}
}
