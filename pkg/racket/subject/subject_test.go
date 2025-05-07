package subject

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input       string
		expected    Subject
		expectError bool
	}{
		{"a.b.c", Subject{Parts: []string{"a", "b", "c"}}, false},
		{"a.b.*", Subject{Parts: []string{"a", "b", "*"}}, false},
		{"a.*.c", Subject{}, true},
		{"*", Subject{Parts: []string{"*"}}, false},
		{"", Subject{Parts: []string{""}}, false},
	}

	for _, test := range tests {
		result, err := Parse(test.input)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for input %q, got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %q: %v", test.input, err)
			}
			if len(result.Parts) != len(test.expected.Parts) {
				t.Errorf("expected %v, got %v", test.expected.Parts, result.Parts)
			}
			for i := range result.Parts {
				if result.Parts[i] != test.expected.Parts[i] {
					t.Errorf("expected %v, got %v", test.expected.Parts, result.Parts)
				}
			}
		}
	}
}

func TestHasWildcard(t *testing.T) {
	tests := []struct {
		subject  Subject
		expected bool
	}{
		{Subject{Parts: []string{"a", "b", "*"}}, true},
		{Subject{Parts: []string{"a", "b", "c"}}, false},
		{Subject{Parts: []string{"*"}}, true},
		{Subject{Parts: []string{"a", "*"}}, true},
		{Subject{Parts: []string{"a"}}, false},
	}

	for _, test := range tests {
		result := test.subject.HasWildcard()
		if result != test.expected {
			t.Errorf("expected %v, got %v for subject %v", test.expected, result, test.subject.Parts)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		subject  Subject
		expected string
	}{
		{Subject{Parts: []string{"a", "b", "c"}}, "a.b.c"},
		{Subject{Parts: []string{"a", "b", "*"}}, "a.b.*"},
		{Subject{Parts: []string{"*"}}, "*"},
		{Subject{Parts: []string{"a"}}, "a"},
		{Subject{Parts: []string{""}}, ""},
	}

	for _, test := range tests {
		result := test.subject.String()
		if result != test.expected {
			t.Errorf("expected %q, got %q for subject %v", test.expected, result, test.subject.Parts)
		}
	}
}
