package perf

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStartAndSpans(t *testing.T) {
	p := New()
	end := p.Start("test-op", "section-a")
	time.Sleep(5 * time.Millisecond)
	end()

	spans := p.Spans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "test-op" {
		t.Errorf("expected name test-op, got %s", spans[0].Name)
	}
	if spans[0].Parent != "section-a" {
		t.Errorf("expected parent section-a, got %s", spans[0].Parent)
	}
	if spans[0].Duration < 1.0 {
		t.Errorf("expected duration >= 1ms, got %.3f ms", spans[0].Duration)
	}
}

func TestConcurrentSpans(t *testing.T) {
	p := New()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			end := p.Start("concurrent", "")
			time.Sleep(time.Millisecond)
			end()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if len(p.Spans()) != 10 {
		t.Fatalf("expected 10 spans, got %d", len(p.Spans()))
	}
}

func TestWriteFile(t *testing.T) {
	p := New()
	end := p.Start("write-test", "")
	end()

	path := filepath.Join(t.TempDir(), "profile.json")
	if err := p.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading profile: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("profile file is empty")
	}
}

func TestWriteFileContent(t *testing.T) {
	p := New()
	end := p.Start("op-a", "sec-1")
	end()
	end2 := p.Start("op-b", "")
	end2()

	path := filepath.Join(t.TempDir(), "profile.json")
	if err := p.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var spans []Span
	if err := json.Unmarshal(data, &spans); err != nil {
		t.Fatalf("profile is not valid JSON: %v", err)
	}
	if len(spans) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(spans))
	}
	if spans[0].Name != "op-a" || spans[0].Parent != "sec-1" {
		t.Errorf("span 0: expected op-a/sec-1, got %s/%s", spans[0].Name, spans[0].Parent)
	}
	if spans[1].Name != "op-b" {
		t.Errorf("span 1: expected op-b, got %s", spans[1].Name)
	}
	if spans[0].Duration < 0 || spans[1].Duration < 0 {
		t.Error("expected non-negative durations")
	}
}

func TestWriteFileIndented(t *testing.T) {
	p := New()
	end := p.Start("op-a", "sec-1")
	end()

	path := filepath.Join(t.TempDir(), "profile.json")
	if err := p.WriteFile(path); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Should be indented (contains newline + spaces)
	if !strings.Contains(string(data), "\n  ") {
		t.Error("expected indented JSON output")
	}
}

func TestSpansReturnsCopy(t *testing.T) {
	p := New()
	end := p.Start("original", "")
	end()

	s1 := p.Spans()
	s1[0].Name = "mutated"

	s2 := p.Spans()
	if s2[0].Name != "original" {
		t.Errorf("Spans() did not return a copy: mutation leaked, got %s", s2[0].Name)
	}
}

func TestPrintEmpty(t *testing.T) {
	p := New()

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	p.Print()

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if buf.Len() > 0 {
		t.Errorf("expected no output for empty profile, got %q", buf.String())
	}
}

func TestPrintGroupsByParent(t *testing.T) {
	p := New()
	end := p.Start("top-op", "")
	end()
	end2 := p.Start("sec-op", "MySection")
	end2()

	origStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	p.Print()

	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "performance profile") {
		t.Error("expected header in print output")
	}
	if !strings.Contains(output, "top-op") {
		t.Error("expected top-op span in output")
	}
	if !strings.Contains(output, "[MySection]") {
		t.Error("expected [MySection] group header in output")
	}
	if !strings.Contains(output, "sec-op") {
		t.Error("expected sec-op span in output")
	}
}
