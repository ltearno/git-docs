package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

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

	written, err := file.Write([]byte(content))
	if err != nil {
		return false
	}

	fmt.Printf("written %d bytes to %s\n", written, path)

	return true
}

func readFile(path string) ([]byte, interface{}) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "cannot open for read"
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, "cannot read"
	}

	fmt.Printf("read from file %s\n%s", path, content)

	return content, nil
}

func (self *MagicGitRepository) getIssuesPath() string {
	return path.Join(self.workingDir, "issues")
}

func (self *MagicGitRepository) getIssueDirPath(name string) string {
	return path.Join(self.getIssuesPath(), name)
}

func (self *MagicGitRepository) getIssueMetadataFilePath(name string) string {
	return path.Join(self.getIssueDirPath(name), "metadata.json")
}

func (self *MagicGitRepository) getIssueContentFilePath(name string) string {
	return path.Join(self.getIssueDirPath(name), "content.md")
}

func (self *MagicGitRepository) GetIssueContent(name string) (*string, interface{}) {
	filePath := self.getIssueContentFilePath(name)
	bytes, err := readFile(filePath)
	if err != nil {
		return nil, "no content"
	}

	content := string(bytes)

	return &content, nil
}

func (self *MagicGitRepository) GetIssueMetadata(name string) (*IssueMetadata, interface{}) {
	filePath := self.getIssueMetadataFilePath(name)
	bytes, err := readFile(filePath)
	if err != nil {
		return nil, "no content"
	}

	result := &IssueMetadata{}

	err = json.Unmarshal(bytes, result)
	if err != nil {
		return nil, "unmarshall"
	}

	return result, nil
}

func (self *MagicGitRepository) AddIssue(name string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	issueDir := self.getIssueDirPath(name)
	if tools.ExistsFile(issueDir) {
		return false
	}

	os.Mkdir(issueDir, 0755)

	ok := writeFileJson(self.getIssueMetadataFilePath(name), IssueMetadata{Flags: []string{}})
	if !ok {
		return false
	}

	issueContentModelBytes, err := assetsgen.Asset("assets/models/issue.md")
	if err != nil {
		return false
	}

	ok = writeFile(self.getIssueContentFilePath(name), string(issueContentModelBytes))
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
