package helm

import (
	"fmt"
	"os"
	"path/filepath"
)

func locateChartNoPanic(locate func() (string, error)) (chartPath string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("locate chart panic: %v", r)
		}
	}()
	return locate()
}

func prepareHelmSettingsForRepoURL(repositoryConfigPath, repositoryCachePath string) error {
	if err := os.MkdirAll(filepath.Dir(repositoryConfigPath), 0o755); err != nil {
		return fmt.Errorf("create repository config dir: %w", err)
	}
	if err := os.MkdirAll(repositoryCachePath, 0o755); err != nil {
		return fmt.Errorf("create repository cache dir: %w", err)
	}
	if _, err := os.Stat(repositoryConfigPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat repository config: %w", err)
	}

	// Helm expects repositories.yaml to exist when resolving charts via repo_url.
	content := []byte("repositories: []\n")
	if err := os.WriteFile(repositoryConfigPath, content, 0o644); err != nil {
		return fmt.Errorf("create repository config: %w", err)
	}
	return nil
}
