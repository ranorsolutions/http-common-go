package formatter

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func newEntryWithFields(fields logrus.Fields) *logrus.Entry {
	return &logrus.Entry{
		Logger: logrus.New(),
		Data:   fields,
		Time:   time.Date(2024, 11, 10, 12, 0, 0, 0, time.UTC),
		Level:  logrus.InfoLevel,
	}
}

// --- extractPrefix() tests ---

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		input       string
		wantPrefix  string
		wantMessage string
	}{
		{"[prefix] actual message", "prefix", "actual message"},
		{"no prefix message", "", "no prefix message"},
		{"[p]msg", "p", "msg"},
	}

	for _, tt := range tests {
		p, m := extractPrefix(tt.input)
		if p != tt.wantPrefix || m != tt.wantMessage {
			t.Errorf("extractPrefix(%q) = (%q,%q), want (%q,%q)",
				tt.input, p, m, tt.wantPrefix, tt.wantMessage)
		}
	}
}

// --- needsQuoting() tests ---

func TestNeedsQuoting(t *testing.T) {
	f := &Formatter{}

	tests := map[string]bool{
		"simple":        false,
		"with space":    true,
		"special!":      true,
		"underscore_ok": true,
		"dash-ok":       false,
		"dot.ok":        false,
		"":              false,
	}

	for input, expected := range tests {
		got := f.needsQuoting(input)
		if got != expected {
			t.Errorf("needsQuoting(%q) = %v, want %v", input, got, expected)
		}
	}
}

// --- appendKeyValue() and appendValue() tests ---

func TestAppendKeyValue_StringAndError(t *testing.T) {
	f := &Formatter{QuoteCharacter: `"`}
	var b bytes.Buffer

	// string without quoting
	f.appendKeyValue(&b, "k1", "value", true)
	// string with quoting
	f.appendKeyValue(&b, "k2", "has space", true)
	// error quoting
	f.appendKeyValue(&b, "err", errors.New("some err"), false)

	out := b.String()
	if !strings.Contains(out, "k1:value") {
		t.Errorf("expected k1:value, got %s", out)
	}
	if !strings.Contains(out, `"has space"`) {
		t.Errorf("expected quoted string, got %s", out)
	}
	if !strings.Contains(out, `"some err"`) {
		t.Errorf("expected quoted error, got %s", out)
	}
}

// --- prefixFieldClashes() tests ---

func TestPrefixFieldClashes(t *testing.T) {
	data := logrus.Fields{
		"time":  "now",
		"msg":   "hi",
		"level": "warn",
	}
	prefixFieldClashes(data)

	if _, ok := data["fields.time"]; !ok {
		t.Error("expected fields.time to be added")
	}
	if _, ok := data["fields.msg"]; !ok {
		t.Error("expected fields.msg to be added")
	}
	if _, ok := data["fields.level"]; !ok {
		t.Error("expected fields.level to be added")
	}
}

// --- Format() plain mode tests ---

func TestFormatPlain(t *testing.T) {
	f := &Formatter{
		ForceFormatting: false,
		DisableColors:   true,
		DisableSorting:  false,
	}
	entry := newEntryWithFields(logrus.Fields{"k1": "v1", "k2": "v2"})
	entry.Message = "plain message"

	b, err := f.Format(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := string(b)
	if !strings.Contains(s, `msg:"plain message"`) {
		t.Errorf("expected quoted msg field, got %s", s)
	}
	if !strings.Contains(s, "k1:v1") || !strings.Contains(s, "k2:v2") {
		t.Errorf("expected fields in output, got %s", s)
	}
	if !strings.HasSuffix(s, "\n") {
		t.Error("expected newline at end of formatted log")
	}
}

// --- Format() colored mode tests ---

func TestFormatColored(t *testing.T) {
	f := &Formatter{
		ForceFormatting: true,
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	}

	entry := newEntryWithFields(logrus.Fields{
		"service": "svc",
		"version": "v1",
	})
	entry.Message = "[PREFIX] colored test"

	b, err := f.Format(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := string(b)
	if !strings.Contains(out, "PREFIX") {
		t.Errorf("expected prefix to appear, got %s", out)
	}
	if !strings.Contains(out, "svc@v1") {
		t.Errorf("expected service@version, got %s", out)
	}
	if !strings.Contains(out, "colored test") {
		t.Errorf("expected message in output, got %s", out)
	}
}

// --- DisableTimestamp & FullTimestamp tests ---

func TestFormat_DisableTimestamp(t *testing.T) {
	f := &Formatter{
		ForceFormatting:  true,
		DisableTimestamp: true,
		ForceColors:      false,
	}
	entry := newEntryWithFields(logrus.Fields{})
	entry.Message = "no ts"

	b, _ := f.Format(entry)
	if strings.Contains(string(b), "time") {
		t.Error("expected no timestamp in output when DisableTimestamp=true")
	}
}

// --- Custom ColorScheme tests ---

func TestSetColorScheme(t *testing.T) {
	f := &Formatter{}
	cs := &ColorScheme{InfoLevelStyle: "red"}
	f.SetColorScheme(cs)

	if f.colorScheme == nil {
		t.Fatal("expected colorScheme to be set")
	}
}

// --- printColored() isolated test ---

func TestPrintColoredBasic(t *testing.T) {
	f := &Formatter{DisableUppercase: true}
	entry := newEntryWithFields(logrus.Fields{
		"prefix":  "api",
		"service": "svc",
		"version": "v1",
	})
	entry.Message = "something happened"
	entry.Level = logrus.InfoLevel

	var buf bytes.Buffer
	f.printColored(&buf, entry, []string{"key1"}, time.RFC3339, defaultCompiledColorScheme)

	out := buf.String()
	if !strings.Contains(out, "INFO") && !strings.Contains(out, "info") {
		t.Errorf("expected level text, got %s", out)
	}
	if !strings.Contains(out, "svc@v1") {
		t.Errorf("expected service@version, got %s", out)
	}
	if !strings.Contains(out, "something happened") {
		t.Errorf("expected message in output, got %s", out)
	}
}

// --- checkIfTerminal() behavior ---

func TestCheckIfTerminalWithNonFile(t *testing.T) {
	f := &Formatter{}
	if f.checkIfTerminal(&bytes.Buffer{}) {
		t.Error("expected non-terminal writer to return false")
	}
}

// --- miniTS() sanity test ---

func TestMiniTSIncreases(t *testing.T) {
	ts1 := miniTS()
	time.Sleep(1 * time.Second)
	ts2 := miniTS()
	if ts2 <= ts1 {
		t.Error("expected miniTS to increase over time")
	}
}
