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

type GitDocsRepository struct {
	gitRepositoryDir *string
	workingDir       string
}

type IssueMetadata struct {
	Tags []string `json:"tags"`
}

func NewGitDocsRepository(gitRepositoryDir *string, workingDir string) *GitDocsRepository {
	return &GitDocsRepository{
		gitRepositoryDir,
		workingDir,
	}
}

func (repo *GitDocsRepository) GitRepositoryDir() *string {
	return repo.gitRepositoryDir
}

func (repo *GitDocsRepository) GetDocuments() ([]string, interface{}) {
	files, err := ioutil.ReadDir(repo.getDocumentsPath())
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

func (repo *GitDocsRepository) GetAllTags() ([]string, interface{}) {
	documents, err := repo.GetDocuments()
	if err != nil {
		return nil, "cannot get documents"
	}

	tagSet := map[string]bool{}
	var result = []string{}

	for _, issue := range documents {
		metadata, err := repo.GetIssueMetadata(issue)
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

func (repo *GitDocsRepository) SearchDocuments(q string) ([]string, interface{}) {
	documents, err := repo.GetDocuments()
	if err != nil {
		return nil, "cannot get documents"
	}

	q = strings.ToLower(q)
	var result = []string{}

	for _, issue := range documents {
		metadata, err := repo.GetIssueMetadata(issue)
		if err != nil {
			return result, "cannot load one metadata"
		}

		if issueMatchSearch(metadata, q) {
			result = append(result, issue)
		}
	}

	return result, nil
}

func (repo *GitDocsRepository) getDocumentsPath() string {
	return path.Join(repo.workingDir, "issues")
}

func (repo *GitDocsRepository) getIssueDirPath(name string) string {
	return path.Join(repo.getDocumentsPath(), name)
}

func (repo *GitDocsRepository) getIssueMetadataFilePath(name string) string {
	return path.Join(repo.getIssueDirPath(name), "metadata.json")
}

func (repo *GitDocsRepository) getIssueContentFilePath(name string) string {
	return path.Join(repo.getIssueDirPath(name), "content.md")
}

func (repo *GitDocsRepository) GetIssueContent(name string) (*string, interface{}) {
	filePath := repo.getIssueContentFilePath(name)
	bytes, err := readFile(filePath)
	if err != nil {
		return nil, "no content"
	}

	content := string(bytes)

	return &content, nil
}

func (repo *GitDocsRepository) SetIssueContent(name string, content string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "invalid name"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	currentContent, err := repo.GetIssueContent(name)
	if err != nil {
		return false, "cannot read issue content"
	}

	if *(currentContent) == content {
		return true, nil
	}

	filePath := repo.getIssueContentFilePath(name)
	ok := writeFile(filePath, content)
	if !ok {
		return false, "error"
	}

	if repo.gitRepositoryDir != nil {
		ok = commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - updated issue content %s", name), repo.workingDir)
		if !ok {
			return false, "commit error"
		}
	}

	return true, nil
}

func (repo *GitDocsRepository) GetIssueMetadata(name string) (*IssueMetadata, interface{}) {
	filePath := repo.getIssueMetadataFilePath(name)
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

func (repo *GitDocsRepository) SetIssueMetadata(name string, metadata *IssueMetadata) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "invalid name"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	bytes, err := json.Marshal(*metadata)
	if err != nil {
		return false, "json error"
	}

	filePath := repo.getIssueMetadataFilePath(name)
	ok := writeFile(filePath, string(bytes))
	if !ok {
		return false, "write file error"
	}

	if repo.gitRepositoryDir != nil {
		ok = commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - updated issue metadata %s", name), repo.workingDir)
		if !ok {
			return false, "commit error"
		}
	}

	return true, nil
}

// list of authors, with their rank on first part of each line :
// git shortlog -sne --all

func (repo *GitDocsRepository) isGitRepositoryClean() bool {
	return isGitRepositoryClean(repo.gitRepositoryDir, repo.workingDir)
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

func (repo *GitDocsRepository) IsClean() (bool, interface{}) {
	if repo.gitRepositoryDir != nil {
		return repo.isGitRepositoryClean(), nil
	}

	return true, nil
}

func (repo *GitDocsRepository) GetStatus() (*string, interface{}) {
	if repo.gitRepositoryDir != nil {
		output, err := execCommand(*repo.gitRepositoryDir, "git", "status")
		if err != nil {
			return nil, err
		}

		return output, nil
	}

	res := "git repository not set"
	return &res, nil
}

func (repo *GitDocsRepository) RenameIssue(name string, newName string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false
		}
	}

	issueDir := repo.getIssueDirPath(name)
	if !tools.ExistsFile(issueDir) {
		return false
	}

	newIssueDir := repo.getIssueDirPath(newName)
	if tools.ExistsFile(newIssueDir) {
		return false
	}

	err := os.Rename(issueDir, newIssueDir)
	if err != nil {
		return false
	}

	if repo.gitRepositoryDir != nil {
		ok := commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - renamed issue %s to %s", name, newIssueDir), repo.workingDir)
		if !ok {
			return false
		}
	}

	return true
}

func (repo *GitDocsRepository) AddIssue(name string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false
		}
	}

	issueDir := repo.getIssueDirPath(name)
	if tools.ExistsFile(issueDir) {
		return false
	}

	os.Mkdir(issueDir, 0755)

	ok := writeFileJson(repo.getIssueMetadataFilePath(name), IssueMetadata{Tags: []string{}})
	if !ok {
		return false
	}

	issueContentModelBytes, err := assetsgen.Asset("assets/models/issue.md")
	if err != nil {
		return false
	}

	ok = writeFile(repo.getIssueContentFilePath(name), string(issueContentModelBytes))
	if !ok {
		return false
	}

	if repo.gitRepositoryDir != nil {
		ok = commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - added issue %s", name), repo.workingDir)
		if !ok {
			return false
		}
	}

	return true
}

func (repo *GitDocsRepository) DeleteIssue(name string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "'/' is forbidden in names"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "git repository is dirty"
		}
	}

	issueDir := repo.getIssueDirPath(name)
	if !tools.ExistsFile(issueDir) {
		return false, "issue does not exists"
	}

	os.RemoveAll(issueDir)

	if repo.gitRepositoryDir != nil {
		ok := commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - deleted issue %s", name), repo.workingDir)
		if !ok {
			return false, false
		}
	}

	return true, nil
}

func (repo *GitDocsRepository) EnsureWorkingSpaceReady() bool {
	if !tools.ExistsFile(repo.workingDir) {
		err := os.Mkdir(repo.getDocumentsPath(), 0755)
		if err != nil {
			fmt.Printf("error creating working dir %v\n!\n", err)
			return false
		}
	}

	return true
}
