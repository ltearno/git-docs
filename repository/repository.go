package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"../assetsgen"
	"../tools"
)

type MagicGitRepository struct {
	gitRepositoryDir string
	workingDir       string
}

type IssueMetadata struct {
	Flags []string `json:"flags"`
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

func writeFileJson(path string, data interface{}) bool {
	json, err := json.Marshal(data)
	if err != nil {
		return false
	}

	return writeFile(path, string(json))
}

func writeFile(path string, content string) bool {
	file, err := os.Create(path)
	if err != nil {
		return false
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	_, err = writer.WriteString(content)
	if err != nil {
		return false
	}

	return true
}

func (self *MagicGitRepository) AddIssue(name string) bool {
	issueDir := path.Join(self.workingDir, "issues", name)
	if tools.ExistsFile(issueDir) {
		return false
	}

	os.Mkdir(issueDir, 0755)

	ok := writeFileJson(path.Join(issueDir, "metadata.json"), IssueMetadata{Flags: []string{}})
	if !ok {
		return false
	}

	issueContentModelBytes, err := assetsgen.Asset("assets/models/issue.md")
	if err != nil {
		return false
	}

	ok = writeFile(path.Join(issueDir, "content.md"), string(issueContentModelBytes))
	if !ok {
		return false
	}

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
