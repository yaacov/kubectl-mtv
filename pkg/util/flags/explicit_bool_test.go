package flags

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestExplicitBool_Set(t *testing.T) {
	tests := []struct {
		input   string
		want    bool
		wantErr bool
	}{
		{"true", true, false},
		{"false", false, false},
		{"True", true, false},
		{"False", false, false},
		{"TRUE", true, false},
		{"FALSE", false, false},
		{"1", true, false},
		{"0", false, false},
		{"yes", true, false},
		{"no", false, false},
		{"Yes", true, false},
		{"No", false, false},
		{"invalid", false, true},
		{"", false, true},
		{"maybe", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var val bool
			b := &ExplicitBool{value: &val}
			err := b.Set(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if err == nil && val != tt.want {
				t.Errorf("Set(%q) = %v, want %v", tt.input, val, tt.want)
			}
		})
	}
}

func TestExplicitBool_String(t *testing.T) {
	val := true
	b := &ExplicitBool{value: &val}
	if got := b.String(); got != "true" {
		t.Errorf("String() = %q, want %q", got, "true")
	}

	val = false
	if got := b.String(); got != "false" {
		t.Errorf("String() = %q, want %q", got, "false")
	}
}

func TestExplicitBool_StringNil(t *testing.T) {
	b := &ExplicitBool{}
	if got := b.String(); got != "false" {
		t.Errorf("String() with nil value = %q, want %q", got, "false")
	}
}

func TestExplicitBool_Type(t *testing.T) {
	b := &ExplicitBool{}
	if got := b.Type(); got != "bool" {
		t.Errorf("Type() = %q, want %q", got, "bool")
	}
}

func TestExplicitBool_IsBoolFlag(t *testing.T) {
	b := &ExplicitBool{}
	if b.IsBoolFlag() {
		t.Error("IsBoolFlag() = true, want false")
	}
}

func TestExplicitBoolVar(t *testing.T) {
	var val bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	ExplicitBoolVar(fs, &val, "my-flag", true, "test flag")

	if val != true {
		t.Errorf("default value = %v, want true", val)
	}

	if err := fs.Parse([]string{"--my-flag", "false"}); err != nil {
		t.Fatalf("Parse(--my-flag false) error: %v", err)
	}
	if val != false {
		t.Errorf("after --my-flag false: val = %v, want false", val)
	}
}

func TestExplicitBoolVar_DefaultFalse(t *testing.T) {
	var val bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	ExplicitBoolVar(fs, &val, "my-flag", false, "test flag")

	if val != false {
		t.Errorf("default value = %v, want false", val)
	}

	if err := fs.Parse([]string{"--my-flag", "true"}); err != nil {
		t.Fatalf("Parse(--my-flag true) error: %v", err)
	}
	if val != true {
		t.Errorf("after --my-flag true: val = %v, want true", val)
	}
}

func TestExplicitBoolVar_EqualsSyntax(t *testing.T) {
	var val bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	ExplicitBoolVar(fs, &val, "my-flag", true, "test flag")

	if err := fs.Parse([]string{"--my-flag=false"}); err != nil {
		t.Fatalf("Parse(--my-flag=false) error: %v", err)
	}
	if val != false {
		t.Errorf("after --my-flag=false: val = %v, want false", val)
	}
}

func TestExplicitBoolVar_Changed(t *testing.T) {
	var val bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	ExplicitBoolVar(fs, &val, "my-flag", false, "test flag")

	if fs.Changed("my-flag") {
		t.Error("flag should not be marked as changed before parsing")
	}

	if err := fs.Parse([]string{"--my-flag", "true"}); err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !fs.Changed("my-flag") {
		t.Error("flag should be marked as changed after parsing")
	}
}

func TestExplicitBoolVar_NoValueErrors(t *testing.T) {
	var val bool
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	ExplicitBoolVar(fs, &val, "my-flag", false, "test flag")

	err := fs.Parse([]string{"--my-flag"})
	if err == nil {
		t.Error("Parse(--my-flag) without value should error")
	}
}
