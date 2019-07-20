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
	gitRepositoryDir *string
	workingDir       string
}

type IssueMetadata struct {
	Tags []string `json:"tags"`
}

func NewMagicGitRepository(gitRepositoryDir *string, workingDir string) *MagicGitRepository {
	return &MagicGitRepository{
		gitRepositoryDir,
		workingDir,
	}
}

func (magic *MagicGitRepository) GitRepositoryDir() *string {
	return magic.gitRepositoryDir
}

func (magic *MagicGitRepository) GetIssues() ([]string, interface{}) {
	files, err := ioutil.ReadDir(magic.getIssuesPath())
	if err != nil {
		return nil, "cannot read dir"
	}

	var result = []string{}

	for _, f := range files {
		if f.IsDir() {
			result = append(result, f.Name())
		}
	}

	return result, nil
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

func readFileJson(path string, out interface{}) interface{} {
	bytes, err := readFile(path)
	if err != nil {
		return "error reading file"
	}

	err = json.Unmarshal(bytes, out)
	if err != nil {
		return "error parsing json"
	}

	return nil
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

func (magic *MagicGitRepository) GetAllTags() ([]string, interface{}) {
	issues, err := magic.GetIssues()
	if err != nil {
		return nil, "cannot get issues"
	}

	tagSet := map[string]bool{}
	var result = []string{}

	for _, issue := range issues {
		metadata, err := magic.GetIssueMetadata(issue)
		if err != nil {
			return result, "cannot load one metadata"
		}

		for _, tag := range metadata.Tags {
			_, alreadyRegistered := tagSet[tag]
			if !alreadyRegistered {
				tagSet[tag] = true
				result = append(result, tag)
			}
		}
	}

	return result, nil
}

func issueTagsContainText(metadata *IssueMetadata, q string) bool {
	for _, tag := range metadata.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}

	return false
}

func issueMatchSearch(metadata *IssueMetadata, q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return true
	} else if strings.HasPrefix(q, "&") {
		q = strings.TrimSpace(q[1:])
		separatorPos := strings.Index(q, " ")
		if separatorPos == 0 {
			return false
		}
		return issueMatchSearch(metadata, q[:separatorPos]) && issueMatchSearch(metadata, q[separatorPos+1:])
	} else if strings.HasPrefix(q, "|") {
		q = strings.TrimSpace(q[1:])
		separatorPos := strings.Index(q, " ")
		if separatorPos == 0 {
			return false
		}
		return issueMatchSearch(metadata, q[:separatorPos]) || issueMatchSearch(metadata, q[separatorPos+1:])
	} else {
		return issueTagsContainText(metadata, q)
	}
}

func (magic *MagicGitRepository) SearchIssues(q string) ([]string, interface{}) {
	issues, err := magic.GetIssues()
	if err != nil {
		return nil, "cannot get issues"
	}

	q = strings.ToLower(q)
	var result = []string{}

	for _, issue := range issues {
		metadata, err := magic.GetIssueMetadata(issue)
		if err != nil {
			return result, "cannot load one metadata"
		}

		if issueMatchSearch(metadata, q) {
			result = append(result, issue)
		}
	}

	return result, nil
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

func (magic *MagicGitRepository) SetIssueContent(name string, content string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "invalid name"
	}

	if magic.gitRepositoryDir != nil {
		if !magic.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	currentContent, err := magic.GetIssueContent(name)
	if err != nil {
		return false, "cannot read issue content"
	}

	if *(currentContent) == content {
		return true, nil
	}

	filePath := magic.getIssueContentFilePath(name)
	ok := writeFile(filePath, content)
	if !ok {
		return false, "error"
	}

	if magic.gitRepositoryDir != nil {
		ok = commitChanges(magic.gitRepositoryDir, fmt.Sprintf("issues() - updated issue content %s", name), magic.workingDir)
		if !ok {
			return false, "commit error"
		}
	}

	return true, nil
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

func (magic *MagicGitRepository) SetIssueMetadata(name string, metadata *IssueMetadata) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "invalid name"
	}

	if magic.gitRepositoryDir != nil {
		if !magic.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	bytes, err := json.Marshal(*metadata)
	if err != nil {
		return false, "json error"
	}

	filePath := magic.getIssueMetadataFilePath(name)
	ok := writeFile(filePath, string(bytes))
	if !ok {
		return false, "write file error"
	}

	if magic.gitRepositoryDir != nil {
		ok = commitChanges(magic.gitRepositoryDir, fmt.Sprintf("issues() - updated issue metadata %s", name), magic.workingDir)
		if !ok {
			return false, "commit error"
		}
	}

	return true, nil
}

// list of authors, with their rank on first part of each line :
// git shortlog -sne --all

func (magic *MagicGitRepository) isGitRepositoryClean() bool {
	return isGitRepositoryClean(magic.gitRepositoryDir, magic.workingDir)
}

func isGitRepositoryClean(gitDir *string, workingDir string) bool {
	if !path.IsAbs(*gitDir) || !path.IsAbs(workingDir) {
		return false
	}

	workingDirRelativePath := workingDir[len(*gitDir):]

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = *gitDir

	out, err := cmd.StdoutPipe()
	if err != nil {
		return false
	}

	err = cmd.Start()
	if err != nil {
		return false
	}

	scanner := bufio.NewScanner(out)
	scanner.Split(bufio.ScanLines)

	clean := true

	for scanner.Scan() {
		file := scanner.Text()[3:]
		if strings.HasPrefix(file, workingDirRelativePath) || strings.HasPrefix(file, "\""+workingDirRelativePath) {
			clean = false
			fmt.Printf("%s is not clean", file)
		}
	}

	if !clean {
		return false
	}

	err = cmd.Wait()
	if err != nil {
		return false
	}

	return true
}

func execCommand(cwd string, name string, args ...string) (*string, interface{}) {
	cmd := exec.Command(name, args...)
	cmd.Dir = cwd

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "cannot pipe stdout"
	}

	err = cmd.Start()
	if err != nil {
		return nil, "cannot start"
	}

	contentBytes, err := ioutil.ReadAll(out)
	if err != nil {
		return nil, "cannot read stdout"
	}

	content := string(contentBytes)

	err = cmd.Wait()
	if err != nil {
		return &content, "cannot wait"
	}

	// if commit has been ok, we should be clean
	return &content, nil
}

func commitChanges(gitRepositoryDir *string, message string, committedDir string) bool {
	if !path.IsAbs(*gitRepositoryDir) || !path.IsAbs(committedDir) {
		return false
	}

	output, err := execCommand(*gitRepositoryDir, "git", "add", committedDir)
	if err != nil {
		fmt.Printf("error staging changes %v\n%s", err, *output)
		return false
	}

	output, err = execCommand(*gitRepositoryDir, "git", "commit", "-m", message)
	if err != nil {
		fmt.Printf("error commit %v\n%s", err, *output)
		return false
	}

	// if commit has been ok, we should be clean
	return isGitRepositoryClean(gitRepositoryDir, committedDir)
}

func (magic *MagicGitRepository) IsClean() (bool, interface{}) {
	if magic.gitRepositoryDir != nil {
		return magic.isGitRepositoryClean(), nil
	}

	return true, nil
}

func (magic *MagicGitRepository) GetStatus() (*string, interface{}) {
	if magic.gitRepositoryDir != nil {
		output, err := execCommand(*magic.gitRepositoryDir, "git", "status")
		if err != nil {
			return nil, err
		}

		return output, nil
	}

	res := "git repository not set"
	return &res, nil
}

func (magic *MagicGitRepository) RenameIssue(name string, newName string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if magic.gitRepositoryDir != nil {
		if !magic.isGitRepositoryClean() {
			return false
		}
	}

	issueDir := magic.getIssueDirPath(name)
	if !tools.ExistsFile(issueDir) {
		return false
	}

	newIssueDir := magic.getIssueDirPath(newName)
	if tools.ExistsFile(newIssueDir) {
		return false
	}

	err := os.Rename(issueDir, newIssueDir)
	if err != nil {
		return false
	}

	if magic.gitRepositoryDir != nil {
		ok := commitChanges(magic.gitRepositoryDir, fmt.Sprintf("issues() - renamed issue %s to %s", name, newIssueDir), magic.workingDir)
		if !ok {
			return false
		}
	}

	return true
}

func (magic *MagicGitRepository) AddIssue(name string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if magic.gitRepositoryDir != nil {
		if !magic.isGitRepositoryClean() {
			return false
		}
	}

	issueDir := magic.getIssueDirPath(name)
	if tools.ExistsFile(issueDir) {
		return false
	}

	os.Mkdir(issueDir, 0755)

	ok := writeFileJson(magic.getIssueMetadataFilePath(name), IssueMetadata{Tags: []string{}})
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

	if magic.gitRepositoryDir != nil {
		ok = commitChanges(magic.gitRepositoryDir, fmt.Sprintf("issues() - added issue %s", name), magic.workingDir)
		if !ok {
			return false
		}
	}

	return true
}

func (magic *MagicGitRepository) DeleteIssue(name string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "'/' is forbidden in names"
	}

	if magic.gitRepositoryDir != nil {
		if !magic.isGitRepositoryClean() {
			return false, "git repository is dirty"
		}
	}

	issueDir := magic.getIssueDirPath(name)
	if !tools.ExistsFile(issueDir) {
		return false, "issue does not exists"
	}

	os.RemoveAll(issueDir)

	if magic.gitRepositoryDir != nil {
		ok := commitChanges(magic.gitRepositoryDir, fmt.Sprintf("issues() - deleted issue %s", name), magic.workingDir)
		if !ok {
			return false, false
		}
	}

	return true, nil
}

func (magic *MagicGitRepository) EnsureWorkingSpaceReady() bool {
	if !tools.ExistsFile(magic.workingDir) {
		err := os.Mkdir(magic.workingDir, 0755)
		if err != nil {
			return false
		}
	}

	if !tools.ExistsFile(magic.workingDir) {
		err := os.Mkdir(magic.getIssuesPath(), 0755)
		if err != nil {
			fmt.Printf("ERROR %v\n!\n", err)
			return false
		}
	}

	return true
}
