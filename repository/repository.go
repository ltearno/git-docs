package repository

import (
	"bufio"
	"fmt"
	"os/exec"
	"path"
	"strings"
)

type GitDocsRepository struct {
	gitRepositoryDir *string
	workingDir       string
}
type GitDocsConfiguration struct {
	Categories []string `json:"categories"`
}

func NewGitDocsRepository(gitRepositoryDir *string, workingDir string) *GitDocsRepository {
	return &GitDocsRepository{
		gitRepositoryDir,
		workingDir,
	}
}

func (repo *GitDocsRepository) GetConfiguration() GitDocsConfiguration {
	config := &GitDocsConfiguration{}
	err := readFileJson(repo.getGitDocsConfigurationFilePath(), config)
	if err != nil {
		config.Categories = []string{}
	}

	return *config
}

func (repo *GitDocsRepository) GetCategories() []string {
	configuration := repo.GetConfiguration()

	return configuration.Categories
}

func (repo *GitDocsRepository) GitRepositoryDir() *string {
	return repo.gitRepositoryDir
}

func (repo *GitDocsRepository) GetWorkingDir() string {
	return repo.workingDir
}

func (repo *GitDocsRepository) GetAllTags(category string) ([]string, interface{}) {
	documents, err := repo.GetDocuments(category)
	if err != nil {
		return nil, "cannot get documents"
	}

	tagSet := map[string]bool{}
	var result = []string{}

	readFileJson(repo.getConfigurationTagsPath(category), &result)
	for _, tag := range result {
		tagSet[tag] = true
	}

	for _, document := range documents {
		metadata, err := repo.GetDocumentMetadata(category, document)
		if err != nil {
			return result, "cannot load one metadata"
		}

		for _, tag := range metadata.getTags() {
			_, alreadyRegistered := tagSet[tag]
			if !alreadyRegistered {
				tagSet[tag] = true
				result = append(result, tag)
			}
		}
	}

	return result, nil
}

func (repo *GitDocsRepository) SetConfiguration(configuration *GitDocsConfiguration) bool {
	return writeFileJson(repo.getGitDocsConfigurationFilePath(), configuration)
}

func (repo *GitDocsRepository) AddCategory(category string) (bool, interface{}) {
	ok := repo.ensureWorkdirReady()
	if !ok {
		return false, "error write file"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	configuration := repo.GetConfiguration()
	if contains(configuration.Categories, category) {
		return true, nil
	}

	configuration.Categories = append(configuration.Categories, category)
	ok = repo.SetConfiguration(&configuration)
	if !ok {
		return false, "error write file"
	}

	ok = repo.ensureCategoryDocumentsDirectoryReady(category)
	if !ok {
		return false, "error init category documents directory"
	}

	ok = repo.ensureCategoryConfigurationDirectoryReady(category)
	if !ok {
		return false, "error init category configuration directory"
	}

	ok = ok && copyAsset("assets/models/model.md", repo.getConfigurationTemplateContentPath(category))
	ok = ok && copyAsset("assets/models/model.json", repo.getConfigurationTemplateMetadataPath(category))
	ok = ok && copyAsset("assets/models/workflow.json", repo.getConfigurationWorkflowPath(category))
	ok = ok && copyAsset("assets/models/tags.json", repo.getConfigurationTagsPath(category))

	if repo.gitRepositoryDir != nil {
		ok = commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - added category %s", category), repo.workingDir)
		if !ok {
			return false, "commit error"
		}
	}

	return ok, nil
}

// Note for later : list of authors, with their rank on first part of each line :
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
