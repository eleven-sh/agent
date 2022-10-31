package env

type WorkspaceConfig struct {
	Repositories []WorkspaceConfigRepository `json:"repositories"`
	Runtimes     WorkspaceConfigRuntimes     `json:"runtimes"`
	VSCode       WorkspaceConfigVSCode       `json:"vscode"`
	RootDirPath  string                      `json:"root_dir_path"`
}

type WorkspaceConfigRepository struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	RootDirPath string `json:"root_dir_path"`
}

type WorkspaceConfigRuntimes map[string]string

type WorkspaceConfigVSCode struct {
	Extensions []string `json:"extensions"`
}

func (wc *WorkspaceConfig) SetRuntimes(runtimes WorkspaceConfigRuntimes) {
	// We don't want the JSON file to
	// have "null" as value for runtimes
	if runtimes == nil {
		runtimes = WorkspaceConfigRuntimes{}
	}

	wc.Runtimes = runtimes
	wc.VSCode.Extensions = getVSCodeExtFromWorkspaceRuntimes(runtimes)
}

func newWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Repositories: []WorkspaceConfigRepository{},
		Runtimes:     map[string]string{},
		VSCode: WorkspaceConfigVSCode{
			Extensions: []string{},
		},
		RootDirPath: "",
	}
}
