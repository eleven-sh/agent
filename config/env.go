package config

import (
	"path/filepath"

	"github.com/eleven-sh/eleven/entities"
)

const (
	ElevenConfigDirPath = "/eleven"

	ElevenAgentConfigDirPath  = ElevenConfigDirPath + "/agent"
	ElevenAgentConfigFilePath = ElevenAgentConfigDirPath + "/config.json"

	VSCodeConfigDirPath = ElevenConfigDirPath + "/vscode"

	ElevenUserName        = "eleven"
	ElevenUserHomeDirPath = "/home/" + ElevenUserName
	ElevenUserShellPath   = "/usr/bin/zsh"

	WorkspaceDirPath = ElevenUserHomeDirPath + "/workspace"
)

var EnvReservedPorts = []string{
	DefaultSSHServerListenPort,
	SSHServerListenPort,
	HTTPServerListenPort,
	HTTPSServerListenPort,
	CaddyAPIListenPort,
}

func GetVSCodeWorkspaceConfigFilePath(envName string) string {
	return filepath.Join(
		VSCodeConfigDirPath,
		entities.BuildSlugForEnv(envName)+".code-workspace",
	)
}
