package repository

import (
	"encoding/json"
	"fmt"
	"git-docs/tools"
	"io/ioutil"
	"os"
	"strings"
)

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

		if tagsMatchSearch(metadata.getTags(), q) {
			result = append(result, document)
		}
	}

	return result, nil
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
	addedTags, removedTags := getTagsDifference(currentMetadata.getTags(), metadata.getTags())
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
