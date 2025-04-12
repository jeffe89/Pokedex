package main

import "testing"

func TestCleanInput(t *testing.T) {

	// Test case struct with various input / expected outputs
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    " hello world ",
			expected: []string{"hello", "world"},
		},

		{
			input:    "does this work?",
			expected: []string{"does", "this", "work?"},
		},

		{
			input:    "Pikachu, I choose you!",
			expected: []string{"pikachu,", "i", "choose", "you!"},
		},
	}

	// Loop over each test case and run unit tests
	for _, c := range cases {
		actual := cleanInput(c.input)

		// Compare length of input slice against the expected slice
		if len(actual) != len(c.expected) {
			t.Errorf("cleanInput(%q) returned slice of length %d, want %d",
				c.input, len(actual), len(c.expected))
		}

		// Iterate through slices to compare each word
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]

			// Check if word matches expected output
			if word != expectedWord {
				t.Errorf("cleanInput(%q) returned word %q at position %d, want %q",
					c.input, word, i, expectedWord)
			}
		}
	}
}
