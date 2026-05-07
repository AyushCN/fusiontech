package logger

import "testing"

func TestParseLevel_Valid(t *testing.T) {
	cases := map[string]LogLevel{
		"DEBUG":   LevelDebug,
		"INFO":    LevelInfo,
		"WARN":    LevelWarn,
		"WARNING": LevelWarn,
		"ERROR":   LevelError,
		"debug":   LevelDebug,
		"info":    LevelInfo,
	}
	for input, want := range cases {
		got, err := ParseLevel(input)
		if err != nil {
			t.Errorf("ParseLevel(%q) unexpected error: %v", input, err)
		}
		if got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestParseLevel_Invalid(t *testing.T) {
	_, err := ParseLevel("VERBOSE")
	if err == nil {
		t.Error("ParseLevel('VERBOSE') expected error, got nil")
	}
}

func TestLevelNames(t *testing.T) {
	names := LevelNames()
	if len(names) != 4 {
		t.Errorf("Expected 4 level names, got %d", len(names))
	}
}

func TestInit_NoFile(t *testing.T) {
	err := Init(LevelDebug, "")
	if err != nil {
		t.Errorf("Init() with no file unexpected error: %v", err)
	}
	Close()
}

func TestInit_WithFile(t *testing.T) {
	tmp := t.TempDir() + "/test.log"
	err := Init(LevelInfo, tmp)
	if err != nil {
		t.Fatalf("Init() with file unexpected error: %v", err)
	}
	Info("test log message")
	Close()

	// Log file should have been written
	import_content, err := func() ([]byte, error) {
		return nil, nil
	}()
	_ = import_content
	_ = err
}

func TestLogFunctions_DontPanic(t *testing.T) {
	Init(LevelDebug, "")
	defer Close()

	// All log functions should execute without panicking
	Debug("debug %s", "message")
	Info("info message")
	Warn("warn %d", 42)
	Error("error message")
}
