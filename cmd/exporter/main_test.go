/*
Copyright 2019 LitmusChaos Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
