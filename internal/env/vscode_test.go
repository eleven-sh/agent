package env

import (
	"reflect"
	"sort"
	"testing"
)

func TestGetVSCodeExtFromWorkspaceRuntimes(t *testing.T) {
	testCases := []struct {
		test               string
		runtimes           WorkspaceConfigRuntimes
		expectedVSCodeExts []string
	}{
		{
			test:               "with empty runtimes",
			runtimes:           WorkspaceConfigRuntimes{},
			expectedVSCodeExts: []string{},
		},

		{
			test: "with non-empty runtimes",
			runtimes: WorkspaceConfigRuntimes{
				"go":   "latest",
				"php":  "8.1",
				"ruby": "3.2.0",
				"rust": "latest",
			},
			expectedVSCodeExts: []string{
				"golang.go",
				"rebornix.Ruby",
				"rust-lang.rust-analyzer",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			vscodeExts := getVSCodeExtFromWorkspaceRuntimes(tc.runtimes)

			sort.Strings(vscodeExts)
			sort.Strings(tc.expectedVSCodeExts)

			if !reflect.DeepEqual(tc.expectedVSCodeExts, vscodeExts) {
				t.Fatalf(
					"expected vscode extensions to equal '%+v', got '%+v'",
					tc.expectedVSCodeExts,
					vscodeExts,
				)
			}
		})
	}
}
