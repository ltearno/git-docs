package repository

import (
	"fmt"
	"os"
	"path"

	"../tools"
)

type MagicGitRepository struct {
	gitRepositoryDir string
	workingDir       string
}

func NewMagicGitRepository(gitRepositoryDir string) *MagicGitRepository {
	return &MagicGitRepository{
		gitRepositoryDir,
		path.Join(gitRepositoryDir, ".magic-git"),
	}
}

func (self *MagicGitRepository) EnsureWorkingSpaceReady() bool {
	if !tools.ExistsFile(self.workingDir) {
		var err = os.Mkdir(self.workingDir, 0755)
		if err != nil {
			return false
		}

		err = os.Mkdir(path.Join(self.workingDir, "issues"), 0755)
		if err != nil {
			fmt.Printf("ERROR %v\n!\n", err)
			return false
		}

		err = os.Mkdir(path.Join(self.workingDir, "tmp"), 0755)
		if err != nil {
			return false
		}
	}

	return true
}
