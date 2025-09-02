package sitegen

import (
	"sort"
	"time"
)

// SortSlice sorts a slice of items based on sort keys.
func SortSlice(items []interface{}, keys []SortKey) {
	if len(keys) == 0 {
		return
	}
	sort.SliceStable(items, func(i, j int) bool {
		return compare(items[i], items[j], keys) < 0
	})
}

func compare(a, b interface{}, keys []SortKey) int {
	for _, k := range keys {
		va := Resolve(a, k.Path)
		vb := Resolve(b, k.Path)
		cmp := compareValues(va, vb)
		if cmp != 0 {
			if k.Asc {
				return cmp
			}
			return -cmp
		}
	}
	return 0
}

func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	switch va := a.(type) {
	case string:
		vb, _ := b.(string)
		// try parse date
		ta, errA := time.Parse(time.RFC3339, va)
		tb, errB := time.Parse(time.RFC3339, vb)
		if errA == nil && errB == nil {
			if ta.Before(tb) {
				return -1
			}
			if ta.After(tb) {
				return 1
			}
			return 0
		}
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
		return 0
	case float64:
		vb, _ := b.(float64)
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
		return 0
	case int:
		vb, _ := b.(int)
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
		return 0
	default:
		// attempt to handle numbers encoded as json.Number via float64
		if fa, ok := toFloat(va); ok {
			fb, _ := toFloat(b)
			if fa < fb {
				return -1
			}
			if fa > fb {
				return 1
			}
			return 0
		}
		sa := GetString(a)
		sb := GetString(b)
		if sa < sb {
			return -1
		}
		if sa > sb {
			return 1
		}
		return 0
	}
}

func toFloat(v interface{}) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	}
	return 0, false
}
