// precedence_test.go: configuration source precedence tests.
//
// Precedence contract, from lowest to highest:
//
//	default < config file < environment variable < command line
//
// Copyright (c) 2025 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// writeConfig writes a temporary JSON config file and returns its path.
func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}

// newPrecedenceSet builds a FlagSet wired to a config file and the APP_ env prefix.
func newPrecedenceSet(t *testing.T, configBody string) *FlagSet {
	t.Helper()
	fs := New("app")
	fs.String("host", "default-host", "Server host")
	fs.Int("port", 1111, "Server port")
	fs.Bool("debug", false, "Debug mode")
	fs.SetConfigFile(writeConfig(t, configBody))
	fs.SetEnvPrefix("APP")
	fs.EnableEnvLookup()
	return fs
}

func TestPrecedence_EnvOverridesConfig(t *testing.T) {
	fs := newPrecedenceSet(t, `{"host":"config-host","port":2222}`)
	t.Setenv("APP_HOST", "env-host")

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if got := fs.GetString("host"); got != "env-host" {
		t.Errorf("host = %q, want %q (env must beat config)", got, "env-host")
	}
	// A flag present only in the config file keeps the config value.
	if got := fs.GetInt("port"); got != 2222 {
		t.Errorf("port = %d, want 2222 (config must beat default)", got)
	}
}

func TestPrecedence_CLIOverridesEnv(t *testing.T) {
	fs := newPrecedenceSet(t, `{}`)
	t.Setenv("APP_HOST", "env-host")

	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "cli-host" {
		t.Errorf("host = %q, want %q (CLI must beat env)", got, "cli-host")
	}
}

func TestPrecedence_CLIOverridesConfig(t *testing.T) {
	fs := newPrecedenceSet(t, `{"host":"config-host"}`)

	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "cli-host" {
		t.Errorf("host = %q, want %q (CLI must beat config)", got, "cli-host")
	}
}

func TestPrecedence_ConfigOverridesDefault(t *testing.T) {
	fs := newPrecedenceSet(t, `{"host":"config-host"}`)

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "config-host" {
		t.Errorf("host = %q, want %q (config must beat default)", got, "config-host")
	}
}

// TestPrecedence_FullChain exercises all four sources in a single parse: each
// flag is set by a different subset of sources and must resolve to the highest.
func TestPrecedence_FullChain(t *testing.T) {
	fs := New("app")
	fs.String("from-cli", "d", "set everywhere, CLI wins")
	fs.String("from-env", "d", "set by config+env, env wins")
	fs.String("from-config", "d", "set by config only")
	fs.String("from-default", "d", "set nowhere")
	fs.SetConfigFile(writeConfig(t, `{
		"from-cli":"config","from-env":"config","from-config":"config"}`))
	fs.SetEnvPrefix("APP")
	fs.EnableEnvLookup()
	t.Setenv("APP_FROM_CLI", "env")
	t.Setenv("APP_FROM_ENV", "env")

	if err := fs.Parse([]string{"--from-cli", "cli"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	for _, tc := range []struct{ flag, want string }{
		{"from-cli", "cli"},
		{"from-env", "env"},
		{"from-config", "config"},
		{"from-default", "d"},
	} {
		if got := fs.GetString(tc.flag); got != tc.want {
			t.Errorf("%s = %q, want %q", tc.flag, got, tc.want)
		}
	}
}

// TestPrecedence_EnvTypedValues verifies that env vars override config values
// for every supported flag type, not just strings.
func TestPrecedence_EnvTypedValues(t *testing.T) {
	fs := New("app")
	fs.String("str", "d", "")
	fs.Int("num", 0, "")
	fs.Bool("flag", false, "")
	fs.Float64("ratio", 0, "")
	fs.Duration("wait", 0, "")
	fs.StringSlice("tags", nil, "")
	fs.SetConfigFile(writeConfig(t, `{
		"str":"config","num":1,"flag":false,"ratio":1.5,
		"wait":"1s","tags":["config"]}`))
	fs.SetEnvPrefix("APP")
	fs.EnableEnvLookup()
	t.Setenv("APP_STR", "env")
	t.Setenv("APP_NUM", "42")
	t.Setenv("APP_FLAG", "true")
	t.Setenv("APP_RATIO", "2.5")
	t.Setenv("APP_WAIT", "30s")
	t.Setenv("APP_TAGS", "a,b")

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if got := fs.GetString("str"); got != "env" {
		t.Errorf("str = %q, want env", got)
	}
	if got := fs.GetInt("num"); got != 42 {
		t.Errorf("num = %d, want 42", got)
	}
	if got := fs.GetBool("flag"); !got {
		t.Errorf("flag = false, want true")
	}
	if got := fs.GetFloat64("ratio"); got != 2.5 {
		t.Errorf("ratio = %v, want 2.5", got)
	}
	if got := fs.GetDuration("wait"); got != 30*time.Second {
		t.Errorf("wait = %v, want 30s", got)
	}
	if got := fs.GetStringSlice("tags"); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("tags = %v, want [a b]", got)
	}
}

// TestSource_ReportsOrigin checks the Source accessors report where each
// resolved value came from.
func TestSource_ReportsOrigin(t *testing.T) {
	fs := New("app")
	fs.String("from-cli", "d", "")
	fs.String("from-env", "d", "")
	fs.String("from-config", "d", "")
	fs.String("from-default", "d", "")
	fs.SetConfigFile(writeConfig(t, `{"from-env":"config","from-config":"config"}`))
	fs.SetEnvPrefix("APP")
	fs.EnableEnvLookup()
	t.Setenv("APP_FROM_ENV", "env")

	if err := fs.Parse([]string{"--from-cli", "cli"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	for _, tc := range []struct{ flag, want string }{
		{"from-cli", "cli"},
		{"from-env", "env"},
		{"from-config", "config"},
		{"from-default", "default"},
	} {
		if got := fs.Source(tc.flag); got != tc.want {
			t.Errorf("fs.Source(%q) = %q, want %q", tc.flag, got, tc.want)
		}
		if got := fs.Lookup(tc.flag).Source(); got != tc.want {
			t.Errorf("Lookup(%q).Source() = %q, want %q", tc.flag, got, tc.want)
		}
	}

	if got := fs.Source("nonexistent"); got != "default" {
		t.Errorf("Source of unknown flag = %q, want %q", got, "default")
	}
}

// TestChanged_TrueForEverySource pins the documented Changed() contract: it
// reports "not the default value", regardless of which source supplied it.
func TestChanged_TrueForEverySource(t *testing.T) {
	fs := New("app")
	fs.String("from-cli", "d", "")
	fs.String("from-env", "d", "")
	fs.String("from-config", "d", "")
	fs.String("from-default", "d", "")
	fs.SetConfigFile(writeConfig(t, `{"from-config":"config"}`))
	fs.SetEnvPrefix("APP")
	fs.EnableEnvLookup()
	t.Setenv("APP_FROM_ENV", "env")

	if err := fs.Parse([]string{"--from-cli", "cli"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	for _, name := range []string{"from-cli", "from-env", "from-config"} {
		if !fs.Changed(name) {
			t.Errorf("Changed(%q) = false, want true", name)
		}
		if !fs.Lookup(name).Changed() {
			t.Errorf("Lookup(%q).Changed() = false, want true", name)
		}
	}
	if fs.Changed("from-default") {
		t.Error("Changed(from-default) = true, want false")
	}
}

// TestReset_ClearsSource verifies Reset restores the default source so that a
// subsequent Parse re-applies the full precedence chain.
func TestReset_ClearsSource(t *testing.T) {
	fs := newPrecedenceSet(t, `{"host":"config-host"}`)

	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.Source("host"); got != "cli" {
		t.Fatalf("Source(host) = %q, want cli", got)
	}

	fs.Reset()
	if got := fs.Source("host"); got != "default" {
		t.Errorf("after Reset, Source(host) = %q, want default", got)
	}
	if fs.Changed("host") {
		t.Error("after Reset, Changed(host) = true, want false")
	}

	if err := fs.ResetFlag("port"); err != nil {
		t.Fatalf("ResetFlag: %v", err)
	}
	if got := fs.Source("port"); got != "default" {
		t.Errorf("after ResetFlag, Source(port) = %q, want default", got)
	}
}

// TestLoadEnvironmentVariables_DoesNotDowngradeCLI guards the manual-call path:
// calling the loader after Parse must not clobber a command-line value.
func TestLoadEnvironmentVariables_DoesNotDowngradeCLI(t *testing.T) {
	fs := newPrecedenceSet(t, `{}`)
	t.Setenv("APP_HOST", "env-host")

	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := fs.LoadEnvironmentVariables(); err != nil {
		t.Fatalf("LoadEnvironmentVariables: %v", err)
	}
	if got := fs.GetString("host"); got != "cli-host" {
		t.Errorf("host = %q, want cli-host", got)
	}
}

// TestLoadConfig_DoesNotDowngradeHigherSources guards the same for config.
func TestLoadConfig_DoesNotDowngradeHigherSources(t *testing.T) {
	fs := newPrecedenceSet(t, `{"host":"config-host"}`)
	t.Setenv("APP_PORT", "3333")

	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := fs.LoadConfig(); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if got := fs.GetString("host"); got != "cli-host" {
		t.Errorf("host = %q, want cli-host (config must not downgrade CLI)", got)
	}
	if got := fs.GetInt("port"); got != 3333 {
		t.Errorf("port = %d, want 3333 (config must not downgrade env)", got)
	}
}

// TestConcurrentReadsAfterParse documents the actual concurrency guarantee:
// a FlagSet is safe for unlimited concurrent reads once Parse has returned and
// no mutating call is in flight. Run under -race to make this meaningful.
func TestConcurrentReadsAfterParse(t *testing.T) {
	fs := newPrecedenceSet(t, `{"host":"config-host","port":2222}`)
	if err := fs.Parse([]string{"--debug"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = fs.GetString("host")
				_ = fs.GetInt("port")
				_ = fs.GetBool("debug")
				_ = fs.Changed("host")
				_ = fs.Source("host")
			}
		}()
	}
	wg.Wait()
}

// TestConfig_DurationValues covers duration flags loaded from a config file.
// JSON has no duration type, so both the human-readable string form and the
// numeric nanosecond form (what encoding/json emits for a time.Duration) are
// accepted.
func TestConfig_DurationValues(t *testing.T) {
	tests := []struct {
		name string
		body string
		want time.Duration
	}{
		{"string form", `{"wait":"1m30s"}`, 90 * time.Second},
		{"nanoseconds", `{"wait":1500000000}`, 1500 * time.Millisecond},
		{"zero", `{"wait":0}`, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := New("app")
			fs.Duration("wait", time.Hour, "")
			fs.SetConfigFile(writeConfig(t, tt.body))
			if err := fs.Parse([]string{}); err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got := fs.GetDuration("wait"); got != tt.want {
				t.Errorf("wait = %v, want %v", got, tt.want)
			}
			if got := fs.Source("wait"); got != "config" {
				t.Errorf("Source(wait) = %q, want config", got)
			}
		})
	}
}

// TestConfig_DurationInvalid checks the error paths of duration config loading.
func TestConfig_DurationInvalid(t *testing.T) {
	tests := []struct{ name, body string }{
		{"unparsable string", `{"wait":"not-a-duration"}`},
		{"wrong type", `{"wait":true}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := New("app")
			fs.Duration("wait", time.Hour, "")
			fs.SetConfigFile(writeConfig(t, tt.body))
			if err := fs.Parse([]string{}); err == nil {
				t.Fatal("Parse succeeded, want error")
			}
			if got := fs.GetDuration("wait"); got != time.Hour {
				t.Errorf("wait = %v, want the default to be untouched", got)
			}
		})
	}
}

// TestConfig_PointerBinding verifies that config-loaded values reach the
// pointer returned at declaration time, for every supported type.
func TestConfig_PointerBinding(t *testing.T) {
	fs := New("app")
	str := fs.String("str", "d", "")
	num := fs.Int("num", 0, "")
	flag := fs.Bool("flag", false, "")
	ratio := fs.Float64("ratio", 0, "")
	wait := fs.Duration("wait", 0, "")
	tags := fs.StringSlice("tags", nil, "")
	fs.SetConfigFile(writeConfig(t, `{
		"str":"c","num":7,"flag":true,"ratio":1.25,
		"wait":"5s","tags":["x","y"]}`))

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if *str != "c" {
		t.Errorf("*str = %q, want c", *str)
	}
	if *num != 7 {
		t.Errorf("*num = %d, want 7", *num)
	}
	if !*flag {
		t.Error("*flag = false, want true")
	}
	if *ratio != 1.25 {
		t.Errorf("*ratio = %v, want 1.25", *ratio)
	}
	if *wait != 5*time.Second {
		t.Errorf("*wait = %v, want 5s", *wait)
	}
	if len(*tags) != 2 || (*tags)[0] != "x" {
		t.Errorf("*tags = %v, want [x y]", *tags)
	}
}

// --- Centralized precedence guard -------------------------------------------
//
// The rule "a source may write a flag only when it ranks at least as high as
// the source that currently owns the value" is enforced inside the two write
// paths, not by their callers. The tests below drive those paths directly so a
// future loader that forgets to guard cannot reintroduce the inversion.

func TestCanSet(t *testing.T) {
	all := []flagSource{sourceDefault, sourceConfig, sourceEnv, sourceCLI}
	for _, current := range all {
		for _, incoming := range all {
			f := &Flag{source: current}
			want := incoming >= current
			if got := f.canSet(incoming); got != want {
				t.Errorf("Flag{source:%s}.canSet(%s) = %t, want %t",
					current, incoming, got, want)
			}
		}
	}
}

// TestSetFlagValueFrom_RejectsDowngrade drives the string parsing path (used by
// the command line and the environment) with an unguarded lower-priority write.
func TestSetFlagValueFrom_RejectsDowngrade(t *testing.T) {
	fs := New("app")
	fs.String("host", "default-host", "")
	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// A lower-priority source must be ignored, without reporting an error:
	// being outranked is the normal outcome, not a failure.
	for _, src := range []flagSource{sourceDefault, sourceConfig, sourceEnv} {
		if err := fs.setFlagValueFrom("host", "downgraded", src); err != nil {
			t.Errorf("setFlagValueFrom(%s) returned error %v, want nil", src, err)
		}
		if got := fs.GetString("host"); got != "cli-host" {
			t.Errorf("after write from %s, host = %q, want cli-host", src, got)
		}
		if got := fs.Source("host"); got != "cli" {
			t.Errorf("after write from %s, Source = %q, want cli", src, got)
		}
	}
}

// TestSetFlagValueFrom_AllowsEqualPriority pins the reason the comparison is
// >= and not >: a repeated command-line flag must keep the last value.
func TestSetFlagValueFrom_AllowsEqualPriority(t *testing.T) {
	fs := New("app")
	fs.String("host", "default-host", "")
	fs.Int("port", 0, "")

	if err := fs.Parse([]string{"--host", "first", "--host", "second", "--port", "1", "--port", "2"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "second" {
		t.Errorf("host = %q, want second (a repeated flag keeps the last value)", got)
	}
	if got := fs.GetInt("port"); got != 2 {
		t.Errorf("port = %d, want 2", got)
	}
}

// TestSetFlagValueFrom_ValidatesOnlyWhenApplied checks that a value which is
// discarded for being outranked is not validated either: rejecting a value that
// will never be used would turn a harmless leftover into a parse failure.
func TestSetFlagValueFrom_ValidatesOnlyWhenApplied(t *testing.T) {
	fs := New("app")
	fs.Int("port", 8080, "")
	if err := fs.SetValidator("port", func(val interface{}) error {
		port, ok := val.(int)
		if !ok {
			return fmt.Errorf("expected int, got %T", val)
		}
		if port < 1024 {
			return fmt.Errorf("port must be >= 1024, got %d", port)
		}
		return nil
	}); err != nil {
		t.Fatalf("SetValidator: %v", err)
	}
	if err := fs.Parse([]string{"--port", "9000"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// 80 would fail validation, but it is outranked and never applied.
	if err := fs.setFlagValueFrom("port", "80", sourceEnv); err != nil {
		t.Errorf("setFlagValueFrom returned %v, want nil", err)
	}
	if got := fs.GetInt("port"); got != 9000 {
		t.Errorf("port = %d, want 9000", got)
	}
}

// TestSetFlagValueFromConfig_RejectsDowngrade drives the typed config path,
// which converts JSON values directly and does not share the string parser.
func TestSetFlagValueFromConfig_RejectsDowngrade(t *testing.T) {
	fs := New("app")
	fs.String("host", "default-host", "")
	fs.Int("port", 0, "")
	fs.SetEnvPrefix("APP")
	fs.EnableEnvLookup()
	t.Setenv("APP_PORT", "3333")

	if err := fs.Parse([]string{"--host", "cli-host"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if err := fs.setFlagValueFromConfig("host", "config-host"); err != nil {
		t.Errorf("setFlagValueFromConfig(host) returned %v, want nil", err)
	}
	if got := fs.GetString("host"); got != "cli-host" {
		t.Errorf("host = %q, want cli-host (config must not downgrade CLI)", got)
	}

	if err := fs.setFlagValueFromConfig("port", float64(4444)); err != nil {
		t.Errorf("setFlagValueFromConfig(port) returned %v, want nil", err)
	}
	if got := fs.GetInt("port"); got != 3333 {
		t.Errorf("port = %d, want 3333 (config must not downgrade env)", got)
	}
	if got := fs.Source("port"); got != "env" {
		t.Errorf("Source(port) = %q, want env", got)
	}
}

// TestSetFlagValueFromConfig_UnknownFlag keeps the existing error path covered.
func TestSetFlagValueFromConfig_UnknownFlag(t *testing.T) {
	fs := New("app")
	if err := fs.setFlagValueFromConfig("nope", "x"); err == nil {
		t.Error("setFlagValueFromConfig on unknown flag returned nil, want error")
	}
}
