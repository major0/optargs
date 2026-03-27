package goarg

import (
	"testing"
)

// migratedConversionTests verifies that parsing through the full GoArg
// pipeline (which now uses optargs.Convert) produces identical results
// for all supported types.
var migratedConversionTests = []struct {
	name  string
	dest  interface{}
	args  []string
	check func(t *testing.T, dest interface{})
}{
	{
		name: "bool_flag",
		dest: &struct {
			Verbose bool `arg:"-v,--verbose"`
		}{},
		args: []string{"--verbose"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Verbose bool `arg:"-v,--verbose"`
			})
			if !s.Verbose {
				t.Error("expected Verbose=true")
			}
		},
	},
	{
		name: "int_scalar",
		dest: &struct {
			Count int `arg:"-c,--count"`
		}{},
		args: []string{"--count", "42"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Count int `arg:"-c,--count"`
			})
			if s.Count != 42 {
				t.Errorf("Count = %d, want 42", s.Count)
			}
		},
	},
	{
		name: "float64_scalar",
		dest: &struct {
			Rate float64 `arg:"--rate"`
		}{},
		args: []string{"--rate", "3.14"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Rate float64 `arg:"--rate"`
			})
			if s.Rate != 3.14 {
				t.Errorf("Rate = %f, want 3.14", s.Rate)
			}
		},
	},
	{
		name: "string_scalar",
		dest: &struct {
			Name string `arg:"-n,--name"`
		}{},
		args: []string{"--name", "hello"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Name string `arg:"-n,--name"`
			})
			if s.Name != "hello" {
				t.Errorf("Name = %q, want %q", s.Name, "hello")
			}
		},
	},
	{
		name: "int_slice",
		dest: &struct {
			Nums []int `arg:"-n,--num"`
		}{},
		args: []string{"--num", "1", "--num", "2", "--num", "3"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Nums []int `arg:"-n,--num"`
			})
			if len(s.Nums) != 3 || s.Nums[0] != 1 || s.Nums[1] != 2 || s.Nums[2] != 3 {
				t.Errorf("Nums = %v, want [1 2 3]", s.Nums)
			}
		},
	},
	{
		name: "default_int",
		dest: &struct {
			Port int `arg:"--port" default:"8080"`
		}{},
		args: []string{},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Port int `arg:"--port" default:"8080"`
			})
			if s.Port != 8080 {
				t.Errorf("Port = %d, want 8080", s.Port)
			}
		},
	},
	{
		name: "default_string",
		dest: &struct {
			Host string `arg:"--host" default:"localhost"`
		}{},
		args: []string{},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Host string `arg:"--host" default:"localhost"`
			})
			if s.Host != "localhost" {
				t.Errorf("Host = %q, want %q", s.Host, "localhost")
			}
		},
	},
	{
		name: "positional_string",
		dest: &struct {
			File string `arg:"positional"`
		}{},
		args: []string{"input.txt"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				File string `arg:"positional"`
			})
			if s.File != "input.txt" {
				t.Errorf("File = %q, want %q", s.File, "input.txt")
			}
		},
	},
	{
		name: "short_option",
		dest: &struct {
			Verbose bool `arg:"-v"`
		}{},
		args: []string{"-v"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Verbose bool `arg:"-v"`
			})
			if !s.Verbose {
				t.Error("expected Verbose=true")
			}
		},
	},
	{
		name: "override_default",
		dest: &struct {
			Port int `arg:"--port" default:"8080"`
		}{},
		args: []string{"--port", "9090"},
		check: func(t *testing.T, dest interface{}) {
			t.Helper()
			s := dest.(*struct {
				Port int `arg:"--port" default:"8080"`
			})
			if s.Port != 9090 {
				t.Errorf("Port = %d, want 9090", s.Port)
			}
		},
	},
}

func TestMigratedTypeConversion(t *testing.T) {
	for _, tt := range migratedConversionTests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseArgs(tt.dest, tt.args); err != nil {
				t.Fatalf("ParseArgs: %v", err)
			}
			tt.check(t, tt.dest)
		})
	}
}
