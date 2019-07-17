package repository

import (
	"fmt"
	"io/ioutil"
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

func (self *MagicGitRepository) Issues() []string {
	files, err := ioutil.ReadDir(path.Join(self.workingDir, "issues"))
	if err != nil {
		return []string{}
	}

	var result = []string{}

	for _, f := range files {
		if f.IsDir() {
			result = append(result, f.Name())
		}
	}

	return result
}

func (self *MagicGitRepository) AddIssue(name string) bool {
	issueDir := path.Join(self.workingDir, "issues", name)
	if tools.ExistsFile(issueDir) {
		return false
	}

	os.Mkdir(issueDir, 0755)

	return true
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
