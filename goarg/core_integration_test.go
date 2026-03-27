package goarg

import (
	"fmt"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/major0/optargs"
)

// Feature: goarg-optargs-integration, Property 4: Handle callbacks produce correct field values
func TestPropertyHandleCallbackCorrectness(t *testing.T) {
	// Sub-property A: Scalar fields — the handler sets the field to the
	// value optargs.Convert would produce.
	t.Run("scalar_fields", func(t *testing.T) {
		f := func(n int) bool {
			type S struct {
				Count int `arg:"-c,--count"`
			}
			s := S{}
			args := []string{"--count", fmt.Sprint(n)}
			if err := ParseArgs(&s, args); err != nil {
				return false
			}

			want, err := optargs.Convert(fmt.Sprint(n), reflect.TypeOf(0))
			if err != nil {
				return false
			}
			return s.Count == want.(int)
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property B: Boolean fields with no argument — handler sets true.
	t.Run("bool_no_arg", func(t *testing.T) {
		f := func(_ uint8) bool {
			type S struct {
				Verbose bool `arg:"-v,--verbose"`
			}
			s := S{}
			if err := ParseArgs(&s, []string{"--verbose"}); err != nil {
				return false
			}
			return s.Verbose == true
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property C: String fields — handler sets the exact string.
	t.Run("string_fields", func(t *testing.T) {
		f := func(val string) bool {
			if len(val) == 0 || len(val) > 100 {
				return true
			}
			// Skip strings that look like options.
			if len(val) > 0 && val[0] == '-' {
				return true
			}

			type S struct {
				Name string `arg:"-n,--name"`
			}
			s := S{}
			if err := ParseArgs(&s, []string{"--name", val}); err != nil {
				return false
			}
			return s.Name == val
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property D: Slice fields — handler appends each element.
	t.Run("slice_append", func(t *testing.T) {
		f := func(a, b int) bool {
			type S struct {
				Nums []int `arg:"-n,--num"`
			}
			s := S{}
			args := []string{"--num", fmt.Sprint(a), "--num", fmt.Sprint(b)}
			if err := ParseArgs(&s, args); err != nil {
				return false
			}
			return len(s.Nums) == 2 && s.Nums[0] == a && s.Nums[1] == b
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})
}
