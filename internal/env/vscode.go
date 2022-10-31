package env

import (
	"encoding/json"
	"os"
)

var RuntimeVSCodeExtensions = map[string][]string{
	"go":     {"golang.go"},
	"ruby":   {"rebornix.Ruby"},
	"rust":   {"rust-lang.rust-analyzer"},
	"python": {"ms-python.python"},
	"java":   {"vscjava.vscode-java-pack"},
	"clang":  {"ms-vscode.cpptools-extension-pack"},
	"docker": {"ms-azuretools.vscode-docker"},
}

// VSCodeWorkspaceConfig matches .code-workspace schema.
// See: https://code.visualstudio.com/docs/editor/multi-root-workspaces#_workspace-file-schema
type VSCodeWorkspaceConfig struct {
	Folders  []VSCodeWorkspaceConfigFolder `json:"folders"`
	Settings map[string]interface{}        `json:"settings"`
}

type VSCodeWorkspaceConfigFolder struct {
	Path string `json:"path"`
}

func newVSCodeWorkspaceConfig() *VSCodeWorkspaceConfig {
	return &VSCodeWorkspaceConfig{
		Folders: []VSCodeWorkspaceConfigFolder{},
		Settings: map[string]interface{}{
			"remote.autoForwardPorts":      true,
			"remote.restoreForwardedPorts": true,
			// Auto-detect (using "/proc") and forward opened port.
			// Way better than "output" that parse terminal output.
			// See: https://github.com/microsoft/vscode/issues/143958#issuecomment-1050959241
			"remote.autoForwardPortsSource": "process",
			// We overwrite the $PATH environment variable in integrated terminal
			// because RVM displays warnings when VSCode changes the order of the paths.
			// See: https://github.com/microsoft/vscode/issues/70248
			"terminal.integrated.env.linux": map[string]interface{}{
				"PATH": "${env:PATH}",
			},
			// Socket is not supported by the container agent.
			"remote.SSH.remoteServerListenOnSocket": false,
			"remote.downloadExtensionsLocally":      false,
		},
	}
}

func saveVSCodeWorkspaceConfigAsFile(
	vscodeWorkspaceConfigFilePath string,
	vscodeWorkspaceConfig *VSCodeWorkspaceConfig,
) error {

	vscodeWorkspaceConfigAsJSON, err := json.Marshal(&vscodeWorkspaceConfig)

	if err != nil {
		return err
	}

	err = os.WriteFile(
		vscodeWorkspaceConfigFilePath,
		vscodeWorkspaceConfigAsJSON,
		os.FileMode(0600),
	)

	if err != nil {
		return err
	}

	// Overwrite umask.
	// See: https://stackoverflow.com/questions/50257981/ioutils-writefile-not-respecting-permissions
	return os.Chmod(
		vscodeWorkspaceConfigFilePath,
		0600,
	)
}

func getVSCodeExtFromWorkspaceRuntimes(
	runtimes WorkspaceConfigRuntimes,
) []string {

	vscodeExtensions := []string{}

	for runtimeName := range runtimes {
		if extensions, hasExtensions := RuntimeVSCodeExtensions[runtimeName]; hasExtensions {
			vscodeExtensions = append(vscodeExtensions, extensions...)
		}
	}

	return vscodeExtensions
}
