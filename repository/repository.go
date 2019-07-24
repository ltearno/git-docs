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

	"git-docs/assetsgen"
	"git-docs/tools"
)

type GitDocsRepository struct {
	gitRepositoryDir *string
	workingDir       string
}

type DocumentMetadata map[string]interface{}

type GitDocsConfiguration struct {
	Categories []string `json:"categories"`
}

type WorkflowElement struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Condition   *string  `json:"condition"`
	AddTags     []string `json:"addTags"`
	RemoveTags  []string `json:"removeTags"`
}

type WorkflowConfiguration map[string][]WorkflowElement

func executeWorkflow(config *WorkflowElement, currentMetadata *DocumentMetadata, metadata *DocumentMetadata) {
	if config.Condition == nil || tagsMatchSearch(currentMetadata.GetTags(), *config.Condition) {
		for _, tagToAdd := range config.AddTags {
			metadata.AddTag(tagToAdd)
		}
		for _, tagToRemove := range config.RemoveTags {
			metadata.RemoveTag(tagToRemove)
		}
	}
}

func (metadata *DocumentMetadata) GetTags() []string {
	result := []string{}
	for _, tag := range (*metadata)["tags"].([]interface{}) {
		result = append(result, tag.(string))
	}
	return result
}

func (metadata *DocumentMetadata) AddTag(addedTag string) bool {
	for _, tag := range (*metadata)["tags"].([]interface{}) {
		if tag.(string) == addedTag {
			return false
		}
	}

	(*metadata)["tags"] = append((*metadata)["tags"].([]interface{}), interface{}(addedTag))

	return true
}

func (metadata *DocumentMetadata) RemoveTag(removedTag string) bool {
	newTags := []interface{}{}
	done := false

	for _, tag := range (*metadata)["tags"].([]interface{}) {
		if tag.(string) != removedTag {
			newTags = append(newTags, tag)
		} else {
			done = true
		}
	}

	(*metadata)["tags"] = newTags

	return done
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

func (repo *GitDocsRepository) GetWorkingDir() string {
	return repo.workingDir
}

func (repo *GitDocsRepository) GetDocuments(category string) ([]string, interface{}) {
	files, err := ioutil.ReadDir(repo.getDocumentsPath(category))
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

		for _, tag := range metadata.GetTags() {
			_, alreadyRegistered := tagSet[tag]
			if !alreadyRegistered {
				tagSet[tag] = true
				result = append(result, tag)
			}
		}
	}

	return result, nil
}

func (repo *GitDocsRepository) SearchDocuments(category string, q string) ([]string, interface{}) {
	documents, err := repo.GetDocuments(category)
	if err != nil {
		return nil, "cannot get documents"
	}

	q = strings.ToLower(q)
	var result = []string{}

	for _, document := range documents {
		metadata, err := repo.GetDocumentMetadata(category, document)
		if err != nil {
			return result, "cannot load one metadata"
		}

		if tagsMatchSearch(metadata.GetTags(), q) {
			result = append(result, document)
		}
	}

	return result, nil
}

func (repo *GitDocsRepository) getGitDocsConfigurationFilePath() string {
	return path.Join(repo.workingDir, "git-docs.json")
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func (repo *GitDocsRepository) ensureDirectoryReady(path string) bool {
	if !tools.ExistsFile(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			fmt.Printf("error creating working dir %v\n!\n", err)
			return false
		}
	}

	return true
}

func (repo *GitDocsRepository) ensureWorkdirReady() bool {
	return repo.ensureDirectoryReady(repo.workingDir)
}

func (repo *GitDocsRepository) ensureCategoryDirectoryReady(category string) bool {
	return repo.ensureDirectoryReady(repo.getCategoryPath(category))
}

func (repo *GitDocsRepository) ensureCategoryDocumentsDirectoryReady(category string) bool {
	ok := repo.ensureCategoryDirectoryReady(category)
	if !ok {
		return false
	}

	ok = repo.ensureDirectoryReady(repo.getDocumentsPath(category))
	if !ok {
		return false
	}

	return true
}

func (repo *GitDocsRepository) ensureCategoryConfigurationDirectoryReady(category string) bool {
	ok := repo.ensureCategoryDirectoryReady(category)
	if !ok {
		return false
	}

	ok = repo.ensureDirectoryReady(repo.getConfigurationPath(category))
	if !ok {
		return false
	}

	return true
}

func copyAsset(assetPath string, targetPath string) bool {
	assetBytes, err := assetsgen.Asset(assetPath)
	if err != nil {
		return false
	}

	ok := writeFile(targetPath, string(assetBytes))
	if !ok {
		return false
	}

	return true
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

/*
	Path locations
*/

func (repo *GitDocsRepository) getCategoryPath(category string) string {
	return path.Join(repo.workingDir, category)
}

func (repo *GitDocsRepository) getConfigurationPath(category string) string {
	return path.Join(repo.getCategoryPath(category), "conf")
}

func (repo *GitDocsRepository) getConfigurationWorkflowPath(category string) string {
	return path.Join(repo.getConfigurationPath(category), "workflow.json")
}

func (repo *GitDocsRepository) getConfigurationTemplateContentPath(category string) string {
	return path.Join(repo.getConfigurationPath(category), "model.md")
}

func (repo *GitDocsRepository) getConfigurationTemplateMetadataPath(category string) string {
	return path.Join(repo.getConfigurationPath(category), "model.json")
}

func (repo *GitDocsRepository) getConfigurationTagsPath(category string) string {
	return path.Join(repo.getConfigurationPath(category), "tags.json")
}

func (repo *GitDocsRepository) getDocumentsPath(category string) string {
	return path.Join(repo.getCategoryPath(category), "docs")
}

func (repo *GitDocsRepository) getDocumentDirPath(category string, name string) string {
	return path.Join(repo.getDocumentsPath(category), name)
}

func (repo *GitDocsRepository) getDocumentMetadataFilePath(category string, name string) string {
	return path.Join(repo.getDocumentDirPath(category, name), "metadata.json")
}

func (repo *GitDocsRepository) getDocumentContentFilePath(category string, name string) string {
	return path.Join(repo.getDocumentDirPath(category, name), "content.md")
}

/*
 */

func (repo *GitDocsRepository) GetDocumentContent(category string, name string) (*string, interface{}) {
	filePath := repo.getDocumentContentFilePath(category, name)
	bytes, err := readFile(filePath)
	if err != nil {
		return nil, "no content"
	}

	content := string(bytes)

	return &content, nil
}

func (repo *GitDocsRepository) SetDocumentContent(category string, name string, content string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "invalid name"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	currentContent, err := repo.GetDocumentContent(category, name)
	if err != nil {
		return false, "cannot read document content"
	}

	if *(currentContent) == content {
		return true, nil
	}

	filePath := repo.getDocumentContentFilePath(category, name)
	ok := writeFile(filePath, content)
	if !ok {
		return false, "error"
	}

	if repo.gitRepositoryDir != nil {
		ok = commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - updated document content %s", name), repo.workingDir)
		if !ok {
			return false, "commit error"
		}
	}

	return true, nil
}

func (repo *GitDocsRepository) GetDocumentMetadata(category string, name string) (*DocumentMetadata, interface{}) {
	filePath := repo.getDocumentMetadataFilePath(category, name)
	bytes, err := readFile(filePath)
	if err != nil {
		return nil, "no content"
	}

	result := &DocumentMetadata{}

	err = json.Unmarshal(bytes, result)
	if err != nil {
		return nil, "unmarshall"
	}

	return result, nil
}

func getTagsDifference(old []string, new []string) ([]string, []string) {
	oldSet := map[string]bool{}
	for _, tag := range old {
		oldSet[tag] = true
	}

	listNew := []string{}

	for _, tag := range new {
		if _, ok := oldSet[tag]; !ok {
			listNew = append(listNew, tag)
		} else {
			delete(oldSet, tag)
		}
	}

	listOld := []string{}

	for tag, value := range oldSet {
		if value {
			listOld = append(listOld, tag)
		}
	}

	return listNew, listOld
}

func chooseWorkflowElement(elements []WorkflowElement, actionName *string) *WorkflowElement {
	if elements == nil || len(elements) == 0 {
		return nil
	}

	if actionName == nil || *actionName == "" {
		return &elements[0]
	}

	for _, element := range elements {
		if element.Name != nil && *element.Name == *actionName {
			return &element
		}
	}

	return nil
}

func (repo *GitDocsRepository) GetWorkflow(category string) (*WorkflowConfiguration, interface{}) {
	workflowConfiguration := &WorkflowConfiguration{}
	err := readFileJson(repo.getConfigurationWorkflowPath(category), workflowConfiguration)
	if err != nil {
		return nil, "cannot read workflow"
	}

	return workflowConfiguration, nil
}

func (repo *GitDocsRepository) SetDocumentMetadata(category string, name string, metadata *DocumentMetadata, actionName *string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "invalid name"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "repository is dirty"
		}
	}

	filePath := repo.getDocumentMetadataFilePath(category, name)

	currentMetadata := &DocumentMetadata{}
	readFileJson(filePath, currentMetadata)

	// process trigger
	// TODO : should be recurrent, at the moment we don't trigger triggers for created and removed tags
	workflowConfiguration, err := repo.GetWorkflow(category)
	if err != nil {
		return false, "cannot get workflow"
	}
	addedTags, removedTags := getTagsDifference(currentMetadata.GetTags(), metadata.GetTags())
	for _, addedTag := range addedTags {
		config, ok := (*workflowConfiguration)[fmt.Sprintf("when-added-%s", addedTag)]
		if ok {
			workflowElement := chooseWorkflowElement(config, actionName)
			if workflowElement != nil {
				executeWorkflow(workflowElement, currentMetadata, metadata)
			}
		}
	}
	for _, removedTag := range removedTags {
		config, ok := (*workflowConfiguration)[fmt.Sprintf("when-removed-%s", removedTag)]
		if ok {
			workflowElement := chooseWorkflowElement(config, actionName)
			if workflowElement != nil {
				executeWorkflow(workflowElement, currentMetadata, metadata)
			}
		}
	}

	bytes, err := json.Marshal(*metadata)
	if err != nil {
		return false, "json error"
	}

	ok := writeFile(filePath, string(bytes))
	if !ok {
		return false, "write file error"
	}

	if repo.gitRepositoryDir != nil {
		ok = commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - updated document metadata %s", name), repo.workingDir)
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

func (repo *GitDocsRepository) RenameDocument(category string, name string, newName string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false
		}
	}

	documentDir := repo.getDocumentDirPath(category, name)
	if !tools.ExistsFile(documentDir) {
		return false
	}

	newDocumentDir := repo.getDocumentDirPath(category, newName)
	if tools.ExistsFile(newDocumentDir) {
		return false
	}

	err := os.Rename(documentDir, newDocumentDir)
	if err != nil {
		return false
	}

	if repo.gitRepositoryDir != nil {
		ok := commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - renamed document %s to %s", name, newDocumentDir), repo.workingDir)
		if !ok {
			return false
		}
	}

	return true
}

func (repo *GitDocsRepository) AddDocument(category string, name string) bool {
	if strings.Contains(name, "/") {
		return false
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false
		}
	}

	documentDir := repo.getDocumentDirPath(category, name)
	if tools.ExistsFile(documentDir) {
		return false
	}

	os.Mkdir(documentDir, 0755)

	documentMetadataModelBytes, err := readFile(repo.getConfigurationTemplateMetadataPath(category))
	if err == nil {
		ok := writeFile(repo.getDocumentMetadataFilePath(category, name), string(documentMetadataModelBytes))
		if !ok {
			return false
		}
	}

	documentContentModelBytes, err := readFile(repo.getConfigurationTemplateContentPath(category))
	if err == nil {
		ok := writeFile(repo.getDocumentContentFilePath(category, name), string(documentContentModelBytes))
		if !ok {
			return false
		}
	}

	if repo.gitRepositoryDir != nil {
		ok := commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - added document %s", name), repo.workingDir)
		if !ok {
			return false
		}
	}

	return true
}

func (repo *GitDocsRepository) DeleteDocument(category string, name string) (bool, interface{}) {
	if strings.Contains(name, "/") {
		return false, "'/' is forbidden in names"
	}

	if repo.gitRepositoryDir != nil {
		if !repo.isGitRepositoryClean() {
			return false, "git repository is dirty"
		}
	}

	documentDir := repo.getDocumentDirPath(category, name)
	if !tools.ExistsFile(documentDir) {
		return false, "document does not exists"
	}

	os.RemoveAll(documentDir)

	if repo.gitRepositoryDir != nil {
		ok := commitChanges(repo.gitRepositoryDir, fmt.Sprintf("documents() - deleted document %s", name), repo.workingDir)
		if !ok {
			return false, false
		}
	}

	return true, nil
}
