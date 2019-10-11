package main

import (
	"os"
	"testing"
)

func TestGetOpenebsEnv(t *testing.T) {
	tests := map[string]struct {
		before        func() // Set "OPENEBS_NAMESPACE" Environment Variable
		after         func() // Unset "OPENEBS_NAMESPACE" Environment Variable
		inputKey      string // Env Variable to get
		fallbackValue string // return value, InCase the Env Variable is Not Available/Not set
		returnValue   string // Value of Env Variable
	}{
		"Test Positive-1": {
			before: func() {
				os.Setenv("OPENEBS_NAMESPACE", "openebs")
			},
			after: func() {
				os.Unsetenv("OPENEBS_NAMESPACE")
			},
			inputKey:      "OPENEBS_NAMESPACE",
			fallbackValue: "N/A",
			returnValue:   "openebs",
		},
		"Test Negative-1": {
			before:        func() {},
			after:         func() {},
			inputKey:      "OPENEBS_NAMESPACE",
			fallbackValue: "N/A",
			returnValue:   "N/A",
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			mock.before()
			actualresult := getOpenebsEnv(mock.inputKey, mock.fallbackValue)
			if mock.returnValue != actualresult {
				t.Fatalf("Test %q failed: expected value=%q, actual value=%q ", name, mock.returnValue, actualresult)
			}
			mock.after()
		})
	}
}
