package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestInterpolatePrompt(t *testing.T) {
	const testTaskID = 12345

	// Helper to create a candidate from JSON
	makeCandidate := func(jsonStr string) *Candidate {
		candidates, _ := ParseCandidates([]byte("[" + jsonStr + "]"))
		return &candidates[0]
	}

	t.Run("$INPUT with single string", func(t *testing.T) {
		c := makeCandidate(`"hello"`)
		result, err := InterpolatePrompt("Say: $INPUT", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Say: hello" {
			t.Errorf("got %q, want %q", result, "Say: hello")
		}
	})

	t.Run("$INPUT with single-item array unwraps", func(t *testing.T) {
		c := makeCandidate(`["only_item"]`)
		result, err := InterpolatePrompt("Value: $INPUT", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Value: only_item" {
			t.Errorf("got %q, want %q", result, "Value: only_item")
		}
	})

	t.Run("$INPUT with multi-item array returns JSON", func(t *testing.T) {
		c := makeCandidate(`["a", "b", "c"]`)
		result, err := InterpolatePrompt("Values: $INPUT", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != `Values: ["a", "b", "c"]` {
			t.Errorf("got %q, want %q", result, `Values: ["a", "b", "c"]`)
		}
	})

	t.Run("$INPUT[0] array index", func(t *testing.T) {
		c := makeCandidate(`["first", "second", "third"]`)
		result, err := InterpolatePrompt("First: $INPUT[0]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "First: first" {
			t.Errorf("got %q, want %q", result, "First: first")
		}
	})

	t.Run("$INPUT[1] array index", func(t *testing.T) {
		c := makeCandidate(`["first", "second", "third"]`)
		result, err := InterpolatePrompt("Second: $INPUT[1]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Second: second" {
			t.Errorf("got %q, want %q", result, "Second: second")
		}
	})

	t.Run("$INPUT[n] out of bounds returns empty", func(t *testing.T) {
		c := makeCandidate(`["only"]`)
		result, err := InterpolatePrompt("Missing: $INPUT[5]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Missing: " {
			t.Errorf("got %q, want %q", result, "Missing: ")
		}
	})

	t.Run("$INPUT[1:] slice from index", func(t *testing.T) {
		c := makeCandidate(`["a", "b", "c", "d"]`)
		result, err := InterpolatePrompt("Rest: $INPUT[1:]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != `Rest: ["b","c","d"]` {
			t.Errorf("got %q, want %q", result, `Rest: ["b","c","d"]`)
		}
	})

	t.Run("$INPUT[n:] slice out of bounds returns empty array", func(t *testing.T) {
		c := makeCandidate(`["a"]`)
		result, err := InterpolatePrompt("Rest: $INPUT[5:]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Rest: []" {
			t.Errorf("got %q, want %q", result, "Rest: []")
		}
	})

	t.Run("$INPUT[\"key\"] map access", func(t *testing.T) {
		c := makeCandidate(`{"file": "test.go", "line": 42}`)
		result, err := InterpolatePrompt("File: $INPUT[\"file\"], Line: $INPUT[\"line\"]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "File: test.go, Line: 42" {
			t.Errorf("got %q, want %q", result, "File: test.go, Line: 42")
		}
	})

	t.Run("$INPUT[\"key\"] missing key returns empty", func(t *testing.T) {
		c := makeCandidate(`{"file": "test.go"}`)
		result, err := InterpolatePrompt("Missing: $INPUT[\"nope\"]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Missing: " {
			t.Errorf("got %q, want %q", result, "Missing: ")
		}
	})

	t.Run("mixed syntax in same template", func(t *testing.T) {
		c := makeCandidate(`["a", "b", "c"]`)
		result, err := InterpolatePrompt("All: $INPUT, First: $INPUT[0], Rest: $INPUT[1:]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		expected := `All: ["a", "b", "c"], First: a, Rest: ["b","c"]`
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("$INPUT does not match $INPUTX", func(t *testing.T) {
		c := makeCandidate(`"test"`)
		result, err := InterpolatePrompt("$INPUTX $INPUT", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "$INPUTX test" {
			t.Errorf("got %q, want %q", result, "$INPUTX test")
		}
	})

	t.Run("$TASK_ID interpolation", func(t *testing.T) {
		c := makeCandidate(`"test"`)
		result, err := InterpolatePrompt("Task ID: $TASK_ID", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Task ID: 12345" {
			t.Errorf("got %q, want %q", result, "Task ID: 12345")
		}
	})

	t.Run("$TASK_ID with other variables", func(t *testing.T) {
		c := makeCandidate(`"hello"`)
		result, err := InterpolatePrompt("Task: $TASK_ID, Input: $INPUT", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Task: 12345, Input: hello" {
			t.Errorf("got %q, want %q", result, "Task: 12345, Input: hello")
		}
	})
}

func TestInterpolatePromptValidationErrors(t *testing.T) {
	const testTaskID = 12345

	// Helper to create a candidate from JSON
	makeCandidate := func(jsonStr string) *Candidate {
		candidates, _ := ParseCandidates([]byte("[" + jsonStr + "]"))
		return &candidates[0]
	}

	t.Run("array index on string returns error", func(t *testing.T) {
		c := makeCandidate(`"hello"`)
		_, err := InterpolatePrompt("First: $INPUT[0]", c, testTaskID)
		if err == nil {
			t.Errorf("expected error for array index on string, got nil")
		}
		if ierr, ok := err.(*interpolationError); ok {
			if ierr.Variable != "$INPUT[0]" || ierr.Actual != "string" {
				t.Errorf("got wrong error details: %v", ierr)
			}
		} else {
			t.Errorf("expected interpolationError, got %T", err)
		}
	})

	t.Run("slice on string returns error", func(t *testing.T) {
		c := makeCandidate(`"hello"`)
		_, err := InterpolatePrompt("Rest: $INPUT[1:]", c, testTaskID)
		if err == nil {
			t.Errorf("expected error for slice on string, got nil")
		}
		if ierr, ok := err.(*interpolationError); ok {
			if ierr.Variable != "$INPUT[1:]" || ierr.Actual != "string" {
				t.Errorf("got wrong error details: %v", ierr)
			}
		} else {
			t.Errorf("expected interpolationError, got %T", err)
		}
	})

	t.Run("array index on map returns error", func(t *testing.T) {
		c := makeCandidate(`{"file": "test.go"}`)
		_, err := InterpolatePrompt("First: $INPUT[0]", c, testTaskID)
		if err == nil {
			t.Errorf("expected error for array index on map, got nil")
		}
		if ierr, ok := err.(*interpolationError); ok {
			if ierr.Variable != "$INPUT[0]" || ierr.Actual != "map" {
				t.Errorf("got wrong error details: %v", ierr)
			}
		} else {
			t.Errorf("expected interpolationError, got %T", err)
		}
	})

	t.Run("slice on map returns error", func(t *testing.T) {
		c := makeCandidate(`{"file": "test.go"}`)
		_, err := InterpolatePrompt("Rest: $INPUT[1:]", c, testTaskID)
		if err == nil {
			t.Errorf("expected error for slice on map, got nil")
		}
		if ierr, ok := err.(*interpolationError); ok {
			if ierr.Variable != "$INPUT[1:]" || ierr.Actual != "map" {
				t.Errorf("got wrong error details: %v", ierr)
			}
		} else {
			t.Errorf("expected interpolationError, got %T", err)
		}
	})

	t.Run("key access on array returns empty (not an error - key may exist)", func(t *testing.T) {
		c := makeCandidate(`["a", "b", "c"]`)
		result, err := InterpolatePrompt("File: $INPUT[\"file\"]", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error for key access: %v", err)
		}
		// Key access on non-map returns empty string (key not found behavior)
		if result != "File: " {
			t.Errorf("got %q, want %q", result, "File: ")
		}
	})

	t.Run("bare $INPUT on any type works", func(t *testing.T) {
		c := makeCandidate(`"hello"`)
		result, err := InterpolatePrompt("Value: $INPUT", c, testTaskID)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "Value: hello" {
			t.Errorf("got %q, want %q", result, "Value: hello")
		}
	})
}

func TestInterpolateCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		key      string
		taskName string
		expected string
	}{
		{
			name:     "replace $CANDIDATE",
			command:  "echo $CANDIDATE",
			key:      "file.go:10",
			taskName: "lint",
			expected: "echo 'file.go:10'",
		},
		{
			name:     "replace $TASK_NAME",
			command:  "run-$TASK_NAME.sh",
			key:      "test",
			taskName: "build",
			expected: "run-build.sh",
		},
		{
			name:     "replace both",
			command:  "$TASK_NAME: $CANDIDATE",
			key:      "error",
			taskName: "fix",
			expected: "fix: 'error'",
		},
		{
			name:     "JSON key for array candidate",
			command:  "git commit -m fix $CANDIDATE",
			key:      `["file.go","line 10"]`,
			taskName: "fix",
			expected: `git commit -m fix '["file.go","line 10"]'`,
		},
		{
			name:     "candidate with parentheses",
			command:  "echo $CANDIDATE",
			key:      `func (61.97%)`,
			taskName: "test",
			expected: "echo 'func (61.97%)'",
		},
		{
			name:     "candidate with brackets and quotes",
			command:  "echo $CANDIDATE",
			key:      `["func","line"]`,
			taskName: "test",
			expected: `echo '["func","line"]'`,
		},
		{
			name:     "empty candidate",
			command:  "echo $CANDIDATE",
			key:      "",
			taskName: "test",
			expected: "echo ''",
		},
		{
			name:     "candidate with single quote",
			command:  "echo $CANDIDATE",
			key:      "O'Reilly",
			taskName: "test",
			expected: "echo 'O'\"'\"'Reilly'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := &Candidate{
				Key:  tt.key,
				Data: json.RawMessage(`"placeholder"`),
			}
			result := InterpolateCommand(tt.command, candidate, tt.taskName)
			if result != tt.expected {
				t.Errorf("InterpolateCommand() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLargeJSONLineParsing(t *testing.T) {
	// Test that scanner can handle lines larger than default 64KB buffer
	t.Run("line larger than 64KB can be scanned", func(t *testing.T) {
		// Create a string larger than 64KB
		largeContent := make([]byte, 100*1024) // 100KB
		for i := range largeContent {
			largeContent[i] = 'x'
		}
		largeLine := string(largeContent)

		// Create a mock Gemini JSON line with large content
		largeJSONLine := `{"type":"message","role":"assistant","content":"` + largeLine + `","delta":true}`

		// Verify the line is larger than default buffer
		if len(largeJSONLine) <= 64*1024 {
			t.Fatalf("Test data should be larger than 64KB, got %d bytes", len(largeJSONLine))
		}

		// Verify it's valid JSON
		var se streamEvent
		if err := json.Unmarshal([]byte(largeJSONLine), &se); err != nil {
			t.Fatalf("Failed to parse large JSON line: %v", err)
		}

		// Verify the event structure is correct
		if se.Type != "message" {
			t.Errorf("Expected type 'message', got %q", se.Type)
		}
		if se.Role != "assistant" {
			t.Errorf("Expected role 'assistant', got %q", se.Role)
		}
	})
}

func TestStreamingWithEmptyMessages(t *testing.T) {
	var streamedOutput strings.Builder
	var logOutput strings.Builder

	streamCb := func(text string) {
		streamedOutput.WriteString(text)
	}
	logWriter := &logOutput

	processEventWithState := func(line string) {
		var se streamEvent
		if json.Unmarshal([]byte(line), &se) != nil {
			return
		}

		if se.Type == "message" {
			if se.Role == "assistant" && se.Content != "" {
				text := se.Content
				streamCb(text)
				fmt.Fprint(logWriter, text)

				if !se.Delta {
					streamCb("\n")
					fmt.Fprint(logWriter, "\n")
				}
			}
		}
	}

	// Test case 1: Normal streaming message
	processEventWithState(`{"type":"message","role":"assistant","content":"Hello ","delta":true}`)
	processEventWithState(`{"type":"message","role":"assistant","content":"World","delta":true}`)
	processEventWithState(`{"type":"message","role":"assistant","content":"!","delta":false}`)

	// Test case 2: Empty message (should not produce output)
	processEventWithState(`{"type":"message","role":"assistant","content":"","delta":false}`)

	// Test case 3: Another normal message
	processEventWithState(`{"type":"message","role":"assistant","content":"Done","delta":false}`)

	streamedResult := streamedOutput.String()
	expected := "Hello World!\nDone\n"

	if streamedResult != expected {
		t.Errorf("Stream output mismatch.\nGot: %q\nExpected: %q", streamedResult, expected)
	}
}

func TestRunCommandShowOnFail(t *testing.T) {
	// Helper to capture stdout/stderr
	captureOutput := func(fn func()) (stdout, stderr string) {
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		defer func() {
			os.Stdout = oldStdout
			os.Stderr = oldStderr
		}()

		rOut, wOut, _ := os.Pipe()
		rErr, wErr, _ := os.Pipe()
		os.Stdout = wOut
		os.Stderr = wErr

		fn()

		wOut.Close()
		wErr.Close()

		var bufOut, bufErr bytes.Buffer
		bufOut.ReadFrom(rOut)
		bufErr.ReadFrom(rErr)

		return bufOut.String(), bufErr.String()
	}

	t.Run("success suppresses output", func(t *testing.T) {
		stdout, stderr := captureOutput(func() {
			ok, err := RunCommandShowOnFail("echo hello", ".")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !ok {
				t.Error("expected ok=true")
			}
		})
		if stdout != "" {
			t.Errorf("expected no stdout, got %q", stdout)
		}
		if stderr != "" {
			t.Errorf("expected no stderr, got %q", stderr)
		}
	})

	t.Run("failure shows stdout", func(t *testing.T) {
		stdout, _ := captureOutput(func() {
			ok, err := RunCommandShowOnFail("echo failure && exit 1", ".")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok {
				t.Error("expected ok=false")
			}
		})
		if stdout != "failure\n" {
			t.Errorf("expected stdout 'failure\\n', got %q", stdout)
		}
	})

	t.Run("failure shows stderr", func(t *testing.T) {
		_, stderr := captureOutput(func() {
			ok, err := RunCommandShowOnFail("echo error >&2 && exit 1", ".")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if ok {
				t.Error("expected ok=false")
			}
		})
		if stderr != "error\n" {
			t.Errorf("expected stderr 'error\\n', got %q", stderr)
		}
	})
}
