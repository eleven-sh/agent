package env

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/eleven-sh/agent/internal/system"
)

type ConfigServedPort string
type ConfigServedPorts map[ConfigServedPort]bool

type ConfigLongRunningProcessWD string
type ConfigLongRunningProcessCmd string
type ConfigLongRunningProcesses map[ConfigLongRunningProcessWD]ConfigLongRunningProcessCmd

type Config struct {
	Workspace            *WorkspaceConfig           `json:"workspace"`
	ServedPorts          ConfigServedPorts          `json:"served_ports"`
	LongRunningProcesses ConfigLongRunningProcesses `json:"long_running_processes"`
}

var configLock sync.RWMutex

func NewConfig() *Config {
	return &Config{
		Workspace:            newWorkspaceConfig(),
		ServedPorts:          ConfigServedPorts{},
		LongRunningProcesses: ConfigLongRunningProcesses{},
	}
}

func LoadConfig(
	configFilePath string,
) (*Config, error) {

	configLock.RLock()
	defer configLock.RUnlock()

	configFileContent, err := os.ReadFile(configFilePath)

	if err != nil {
		return nil, err
	}

	var config *Config
	err = json.Unmarshal(configFileContent, &config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func LoadConfigIfExists(
	configFilePath string,
) (*Config, error) {

	agentConfigExists, err := system.DoesFileExist(configFilePath)

	if err != nil {
		return nil, err
	}

	if agentConfigExists {
		return LoadConfig(configFilePath)
	}

	return nil, nil
}

func SaveConfigAsFile(
	configFilePath string,
	config *Config,
) error {

	configLock.Lock()
	defer configLock.Unlock()

	configAsJSON, err := json.Marshal(config)

	if err != nil {
		return err
	}

	err = os.WriteFile(
		configFilePath,
		configAsJSON,
		os.FileMode(0600),
	)

	if err != nil {
		return err
	}

	// Overwrite umask.
	// See: https://stackoverflow.com/questions/50257981/ioutils-writefile-not-respecting-permissions
	return os.Chmod(
		configFilePath,
		0600,
	)
}
