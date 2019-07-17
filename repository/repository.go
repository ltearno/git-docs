package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

func (magic *MagicGitRepository) Issues() []string {
	files, err := ioutil.ReadDir(path.Join(magic.workingDir, "issues"))
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

func (magic *MagicGitRepository) getIssuesPath() string {
	return path.Join(magic.workingDir, "issues")
}

func (magic *MagicGitRepository) getIssueDirPath(name string) string {
	return path.Join(magic.getIssuesPath(), name)
}

func (magic *MagicGitRepository) getIssueMetadataFilePath(name string) string {
	return path.Join(magic.getIssueDirPath(name), "metadata.json")
}

func (magic *MagicGitRepository) getIssueContentFilePath(name string) string {
	return path.Join(magic.getIssueDirPath(name), "content.md")
}

func (magic *MagicGitRepository) GetIssueContent(name string) (*string, interface{}) {
	filePath := magic.getIssueContentFilePath(name)
	bytes, err := readFile(filePath)
	if err != nil {
		return nil, "no content"
	}

	content := string(bytes)

	return &content, nil
}

func (magic *MagicGitRepository) GetIssueMetadata(name string) (*IssueMetadata, interface{}) {
	filePath := magic.getIssueMetadataFilePath(name)
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

func isGitRepositoryClean(dir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir

	out, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}

	err = cmd.Start()
	if err != nil {
		return false
	}

	content, err := ioutil.ReadAll(out)
	if err != nil {
		return false
	}

	err = cmd.Wait()
	if err != nil {
		return false
	}

	if string(content) == "" {
		return true
	}

	fmt.Printf("git repository is not clean !")

	return false
}

func (magic *MagicGitRepository) AddIssue(name string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if !isGitRepositoryClean(magic.gitRepositoryDir) {
		return false
	}

	issueDir := magic.getIssueDirPath(name)
	if tools.ExistsFile(issueDir) {
		return false
	}

	os.Mkdir(issueDir, 0755)

	ok := writeFileJson(magic.getIssueMetadataFilePath(name), IssueMetadata{Flags: []string{}})
	if !ok {
		return false
	}

	issueContentModelBytes, err := assetsgen.Asset("assets/models/issue.md")
	if err != nil {
		return false
	}

	ok = writeFile(magic.getIssueContentFilePath(name), string(issueContentModelBytes))
	if !ok {
		return false
	}

	return true
}

func (magic *MagicGitRepository) EnsureWorkingSpaceReady() bool {
	if !tools.ExistsFile(magic.workingDir) {
		var err = os.Mkdir(magic.workingDir, 0755)
		if err != nil {
			return false
		}

		writeFile(path.Join(magic.workingDir, ".gitignore"), "tmp")

		err = os.Mkdir(path.Join(magic.workingDir, "issues"), 0755)
		if err != nil {
			fmt.Printf("ERROR %v\n!\n", err)
			return false
		}

		err = os.Mkdir(path.Join(magic.workingDir, "tmp"), 0755)
		if err != nil {
			return false
		}
	}

	return true
}
