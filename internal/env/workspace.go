package env

import (
	"os"
	"path/filepath"

	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/system"
	"github.com/eleven-sh/agent/proto"
	"github.com/eleven-sh/eleven/entities"
)

func PrepareWorkspace(
	agentConfig *Config,
	envName string,
	repositories []*proto.EnvRepository,
) error {
	vscodeWorkspaceConfig := newVSCodeWorkspaceConfig()

	// The method "PrepareWorkspace" could
	// be called multiple times in case of error
	// so we need to make sure that our code is idempotent
	err := system.RemoveDirContent(
		config.WorkspaceDirPath,
	)

	if err != nil {
		return err
	}

	if len(repositories) == 0 {
		placeholderWorkspaceDir, err := createPlaceholderWorkspaceDir(envName)

		if err != nil {
			return err
		}

		vscodeWorkspaceConfig.Folders = []VSCodeWorkspaceConfigFolder{
			{
				Path: placeholderWorkspaceDir,
			},
		}

		agentConfig.Workspace.RootDirPath = placeholderWorkspaceDir
	}

	if len(repositories) > 0 {
		repoOwnersCount := countRepoOwners(repositories)

		for _, repository := range repositories {
			repoDirPathInWorkspace := getRepoDirPathInWorkspace(
				repository,
				repoOwnersCount,
			)

			err = addRepoToWorkspace(
				repository.Owner,
				repository.Name,
				repoDirPathInWorkspace,
				agentConfig.Workspace,
				vscodeWorkspaceConfig,
			)

			if err != nil {
				return err
			}
		}

		agentConfig.Workspace.RootDirPath = config.WorkspaceDirPath

		if len(repositories) == 1 {
			agentConfig.Workspace.RootDirPath = agentConfig.Workspace.Repositories[0].RootDirPath
		}
	}

	err = saveVSCodeWorkspaceConfigAsFile(
		config.GetVSCodeWorkspaceConfigFilePath(envName),
		vscodeWorkspaceConfig,
	)

	if err != nil {
		return err
	}

	return SaveConfigAsFile(
		config.ElevenAgentConfigFilePath,
		agentConfig,
	)
}

func addRepoToWorkspace(
	repoOwner string,
	repoName string,
	repoDirPathInWorkspace string,
	workspaceConfig *WorkspaceConfig,
	vscodeWorkspaceConfig *VSCodeWorkspaceConfig,
) error {

	err := cloneGitHubRepo(
		repoOwner,
		repoName,
		repoDirPathInWorkspace,
	)

	if err != nil {
		return err
	}

	workspaceConfigRepository := WorkspaceConfigRepository{
		Owner:       repoOwner,
		Name:        repoName,
		RootDirPath: repoDirPathInWorkspace,
	}

	workspaceConfig.Repositories = append(
		workspaceConfig.Repositories,
		workspaceConfigRepository,
	)

	vscodeWorkspaceConfig.Folders = append(
		vscodeWorkspaceConfig.Folders,
		VSCodeWorkspaceConfigFolder{
			Path: repoDirPathInWorkspace,
		},
	)

	return nil
}

func createPlaceholderWorkspaceDir(
	envName string,
) (string, error) {

	dirPathInWorkspace := filepath.Join(
		config.WorkspaceDirPath,
		entities.BuildSlugForEnv(envName),
	)

	err := os.Mkdir(
		dirPathInWorkspace,
		os.FileMode(0775),
	)

	if err != nil {
		return "", err
	}

	return dirPathInWorkspace, nil
}

func countRepoOwners(
	repositories []*proto.EnvRepository,
) int {

	repoOwnersSet := map[string]bool{}
	repoOwnersCt := 0

	for _, repository := range repositories {
		if _, ownerExists := repoOwnersSet[repository.Owner]; ownerExists {
			continue
		}

		repoOwnersSet[repository.Owner] = true
		repoOwnersCt++
	}

	return repoOwnersCt
}

func getRepoDirPathInWorkspace(
	repository *proto.EnvRepository,
	repoOwnersCount int,
) string {

	repoDirName := repository.Name

	if repoOwnersCount > 1 {
		repoDirName = repository.Owner + "-" + repository.Name
	}

	return filepath.Join(
		config.WorkspaceDirPath,
		entities.BuildSlugForEnv(repoDirName),
	)
}
