package main

import "testing"

func TestLowercaseFirst(t *testing.T) {
	var tests = map[string]struct {
		input    string
		expected string
	}{
		"empty string": {
			input:    "",
			expected: "",
		},
		"first lower case": {
			input:    "helloWorld",
			expected: "helloWorld",
		},
		"first upper case": {
			input:    "HelloWorld",
			expected: "helloWorld",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := lowercaseFirst(test.input)
			if actual != test.expected {
				t.Errorf(
					"unexpected output\nexpected: [%v]\nactual:   [%v]",
					test.expected,
					actual,
				)
			}
		})
	}
}

func TestUppercaseFirst(t *testing.T) {
	var tests = map[string]struct {
		input    string
		expected string
	}{
		"empty string": {
			input:    "",
			expected: "",
		},
		"first upper case": {
			input:    "HelloWorld",
			expected: "HelloWorld",
		},
		"first lower case": {
			input:    "helloWorld",
			expected: "HelloWorld",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := uppercaseFirst(test.input)
			if actual != test.expected {
				t.Errorf(
					"unexpected output\nexpected: [%v]\nactual:   [%v]",
					test.expected,
					actual,
				)
			}
		})
	}
}
