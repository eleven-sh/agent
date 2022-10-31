package env

import (
	"reflect"
	"sort"
	"testing"
)

func TestWorkspaceConfigSetRuntimes(t *testing.T) {
	testCases := []struct {
		test               string
		runtimes           WorkspaceConfigRuntimes
		expectedRuntimes   WorkspaceConfigRuntimes
		expectedVSCodeExts []string
	}{
		{
			test:               "with nil runtimes",
			runtimes:           nil,
			expectedRuntimes:   WorkspaceConfigRuntimes{},
			expectedVSCodeExts: []string{},
		},

		{
			test:               "with empty runtimes",
			runtimes:           WorkspaceConfigRuntimes{},
			expectedRuntimes:   WorkspaceConfigRuntimes{},
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
			expectedRuntimes: WorkspaceConfigRuntimes{
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
			workspaceConfig := newWorkspaceConfig()

			workspaceConfig.SetRuntimes(tc.runtimes)

			if !reflect.DeepEqual(tc.expectedRuntimes, workspaceConfig.Runtimes) {
				t.Fatalf(
					"expected runtimes to equal '%+v', got '%+v'",
					tc.expectedRuntimes,
					workspaceConfig.Runtimes,
				)
			}

			sort.Strings(tc.expectedVSCodeExts)
			sort.Strings(workspaceConfig.VSCode.Extensions)

			if !reflect.DeepEqual(tc.expectedVSCodeExts, workspaceConfig.VSCode.Extensions) {
				t.Fatalf(
					"expected vscode extensions to equal '%+v', got '%+v'",
					tc.expectedVSCodeExts,
					workspaceConfig.VSCode.Extensions,
				)
			}
		})
	}
}
