package optargs

import (
	"strconv"
	"testing"
	"testing/quick"
	"time"
)

// TestPropertySliceValueAppendGetSlice verifies that for any slice value type
// and any sequence of valid element strings, calling Append() for each element
// and then GetSlice() returns a slice whose tail contains the string
// representations of the appended elements in order.
// **Validates: Requirements 4.2**
func TestPropertySliceValueAppendGetSlice(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("stringSlice", func(t *testing.T) {
		f := func(initial []string, appended []string) bool {
			// Filter out strings containing commas (Set splits on comma).
			initial = filterNoComma(initial)
			appended = filterNoComma(appended)

			var dest []string
			v := NewStringSliceValue(initial, &dest).(*sliceValue)

			for _, s := range appended {
				if err := v.Append(s); err != nil {
					return false
				}
			}

			got := v.GetSlice()
			if len(got) != len(initial)+len(appended) {
				return false
			}
			// Tail must match appended elements.
			tail := got[len(initial):]
			for i, s := range appended {
				if tail[i] != s {
					return false
				}
			}
			return true
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("intSlice", func(t *testing.T) {
		f := func(initial []int, appended []int) bool {
			var dest []int
			v := NewIntSliceValue(initial, &dest).(*sliceValue)

			for _, n := range appended {
				if err := v.Append(strconv.Itoa(n)); err != nil {
					return false
				}
			}

			got := v.GetSlice()
			if len(got) != len(initial)+len(appended) {
				return false
			}
			tail := got[len(initial):]
			for i, n := range appended {
				if tail[i] != strconv.Itoa(n) {
					return false
				}
			}
			return true
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("durationSlice", func(t *testing.T) {
		f := func(initialNs []int64, appendedNs []int64) bool {
			initial := toDurations(initialNs)
			appended := toDurations(appendedNs)

			var dest []time.Duration
			v := NewDurationSliceValue(initial, &dest).(*durationSliceValue)

			for _, d := range appended {
				if err := v.Append(d.String()); err != nil {
					return false
				}
			}

			got := v.GetSlice()
			if len(got) != len(initial)+len(appended) {
				return false
			}
			tail := got[len(initial):]
			for i, d := range appended {
				if tail[i] != d.String() {
					return false
				}
			}
			return true
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// TestPropertySliceValueReplaceGetSlice verifies that calling Replace() with a
// list of valid element strings and then GetSlice() returns exactly those
// strings.
// **Validates: Requirements 4.2**
func TestPropertySliceValueReplaceGetSlice(t *testing.T) {
	cfg := &quick.Config{MaxCount: 100}

	t.Run("stringSlice", func(t *testing.T) {
		f := func(initial []string, replacement []string) bool {
			initial = filterNoComma(initial)
			replacement = filterNoComma(replacement)

			var dest []string
			v := NewStringSliceValue(initial, &dest).(*sliceValue)

			if err := v.Replace(replacement); err != nil {
				return false
			}

			got := v.GetSlice()
			return slicesEqual(got, replacement)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("intSlice", func(t *testing.T) {
		f := func(initial []int, replacement []int) bool {
			var dest []int
			v := NewIntSliceValue(initial, &dest).(*sliceValue)

			strs := make([]string, len(replacement))
			for i, n := range replacement {
				strs[i] = strconv.Itoa(n)
			}

			if err := v.Replace(strs); err != nil {
				return false
			}

			got := v.GetSlice()
			return slicesEqual(got, strs)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})

	t.Run("durationSlice", func(t *testing.T) {
		f := func(initialNs []int64, replacementNs []int64) bool {
			initial := toDurations(initialNs)
			replacement := toDurations(replacementNs)

			var dest []time.Duration
			v := NewDurationSliceValue(initial, &dest).(*durationSliceValue)

			strs := make([]string, len(replacement))
			for i, d := range replacement {
				strs[i] = d.String()
			}

			if err := v.Replace(strs); err != nil {
				return false
			}

			got := v.GetSlice()
			// Duration round-trips through String(), so compare via String().
			expected := make([]string, len(replacement))
			for i, d := range replacement {
				expected[i] = d.String()
			}
			return slicesEqual(got, expected)
		}
		if err := quick.Check(f, cfg); err != nil {
			t.Error(err)
		}
	})
}

// filterNoComma returns only strings that don't contain commas,
// since Set() splits on commas which would confuse element counting.
func filterNoComma(ss []string) []string {
	out := make([]string, 0, len(ss))
	for _, s := range ss {
		if !containsComma(s) {
			out = append(out, s)
		}
	}
	return out
}

func containsComma(s string) bool {
	for _, c := range s {
		if c == ',' {
			return true
		}
	}
	return false
}

func slicesEqual(a, b []string) bool {
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

func toDurations(ns []int64) []time.Duration {
	ds := make([]time.Duration, len(ns))
	for i, n := range ns {
		ds[i] = time.Duration(n)
	}
	return ds
}
