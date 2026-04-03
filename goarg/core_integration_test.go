package goarg

import (
	"reflect"
	"strconv"
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
			args := []string{"--count", strconv.Itoa(n)}
			if err := ParseArgs(&s, args); err != nil {
				return false
			}

			want, err := optargs.Convert(strconv.Itoa(n), reflect.TypeFor[int]())
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
			args := []string{"--num", strconv.Itoa(a), "--num", strconv.Itoa(b)}
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

// Feature: goarg-optargs-integration, Property 5: Non-invoked subcommand fields are nil
func TestPropertySubcommandNilOut(t *testing.T) {
	type ServerCmd struct {
		Port int `arg:"--port"`
	}
	type ClientCmd struct {
		URL string `arg:"--url"`
	}
	type BackupCmd struct {
		Path string `arg:"--path"`
	}

	// Sub-property A: Invoking one subcommand nils out the others.
	t.Run("two_subcommands", func(t *testing.T) {
		f := func(port uint16) bool {
			type Args struct {
				Server *ServerCmd `arg:"subcommand:server"`
				Client *ClientCmd `arg:"subcommand:client"`
			}
			a := Args{}
			args := []string{"server", "--port", strconv.FormatUint(uint64(port), 10)}
			if err := ParseArgs(&a, args); err != nil {
				return false
			}
			return a.Server != nil && a.Client == nil && a.Server.Port == int(port)
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property B: Three subcommands — two non-invoked are nil.
	t.Run("three_subcommands", func(t *testing.T) {
		f := func(url string) bool {
			if len(url) == 0 || len(url) > 50 {
				return true
			}
			if len(url) > 0 && url[0] == '-' {
				return true
			}

			type Args struct {
				Server *ServerCmd `arg:"subcommand:server"`
				Client *ClientCmd `arg:"subcommand:client"`
				Backup *BackupCmd `arg:"subcommand:backup"`
			}
			a := Args{}
			args := []string{"client", "--url", url}
			if err := ParseArgs(&a, args); err != nil {
				return false
			}
			return a.Client != nil && a.Server == nil && a.Backup == nil && a.Client.URL == url
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 200}); err != nil {
			t.Error(err)
		}
	})

	// Sub-property C: No subcommand invoked — all fields are nil.
	t.Run("no_subcommand", func(t *testing.T) {
		f := func(_ uint8) bool {
			type Args struct {
				Verbose bool       `arg:"-v,--verbose"`
				Server  *ServerCmd `arg:"subcommand:server"`
				Client  *ClientCmd `arg:"subcommand:client"`
			}
			a := Args{}
			if err := ParseArgs(&a, []string{"-v"}); err != nil {
				return false
			}
			return a.Server == nil && a.Client == nil
		}
		if err := quick.Check(f, &quick.Config{MaxCount: 50}); err != nil {
			t.Error(err)
		}
	})
}
