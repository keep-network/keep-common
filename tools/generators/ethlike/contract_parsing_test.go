package main

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

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

func TestCamelCase(t *testing.T) {
	var tests = map[string]struct {
		input    string
		expected string
	}{
		"empty string": {
			input:    "",
			expected: "",
		},
		"no underscores": {
			input:    "HelloWorld",
			expected: "helloWorld",
		},
		"with underscores": {
			input:    "hello_world",
			expected: "helloWorld",
		},
		"one underscore first": {
			input:    "_beacon_callback",
			expected: "beaconCallback",
		},
		"multiple underscores first": {
			input:    "__beacon_callback",
			expected: "beaconCallback",
		},
	}

	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := camelCase(test.input)
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

func TestMethodStability(t *testing.T) {
	allMethods := make(map[string]abi.Method)
	allMethods["boop"] = abi.Method{Name: "boop", RawName: "boop"}
	allMethods["boop0"] = abi.Method{Name: "boop0", RawName: "boop"}
	allMethods["bap"] = abi.Method{Name: "bap", RawName: "bap", Constant: true}
	allMethods["sap"] = abi.Method{Name: "sap", RawName: "sap"}
	allMethods["map"] = abi.Method{Name: "map", RawName: "map", Constant: true}
	allMethods["map0"] = abi.Method{Name: "map0", RawName: "map"}

	payableMethods := make(map[string]struct{})
	payableMethods["boop"] = struct{}{}

	expectedConstMethodOrder := []string{"bap", "map"}
	expectedNonConstMethodOrder := []string{"boop", "boop0", "map0", "sap"}

	// Run 50 times to make sure we trigger Go's map key randomization, if
	// applicable.
	for i := 0; i < 50; i++ {
		constMethods, nonConstMethods := buildMethodInfo(payableMethods, allMethods)

		methodNames := []string{}
		for _, constMethod := range constMethods {
			methodNames = append(methodNames, constMethod.LowerName)
		}
		if !reflect.DeepEqual(methodNames, expectedConstMethodOrder) {
			t.Fatalf(
				"unexpected const method order\nexpected: [%v]\nactual:   [%v]",
				expectedConstMethodOrder,
				methodNames,
			)
		}

		methodNames = []string{}
		for _, nonConstMethod := range nonConstMethods {
			methodNames = append(methodNames, nonConstMethod.LowerName)
		}
		if !reflect.DeepEqual(methodNames, expectedNonConstMethodOrder) {
			t.Fatalf(
				"unexpected non-const method order\nexpected: [%v]\nactual:   [%v]",
				expectedNonConstMethodOrder,
				methodNames,
			)
		}

	}
}

func TestEventStability(t *testing.T) {
	allEvents := make(map[string]abi.Event)
	allEvents["boop"] = abi.Event{Name: "boop", RawName: "boop"}
	allEvents["bap"] = abi.Event{Name: "bap", RawName: "bap"}
	allEvents["sap"] = abi.Event{Name: "sap", RawName: "sap"}
	allEvents["map"] = abi.Event{Name: "map", RawName: "map"}

	expectedEventOrder := []string{"bap", "boop", "map", "sap"}

	// Run 50 times to make sure we trigger Go's map key randomization, if
	// applicable.
	for i := 0; i < 50; i++ {
		events := buildEventInfo("b", allEvents)

		eventNames := []string{}
		for _, event := range events {
			eventNames = append(eventNames, event.LowerName)
		}
		if !reflect.DeepEqual(eventNames, expectedEventOrder) {
			t.Fatalf(
				"unexpected const method order\nexpected: [%v]\nactual:   [%v]",
				expectedEventOrder,
				eventNames,
			)
		}
	}
}
