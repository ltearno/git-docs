package repository

import "path"

func (repo *GitDocsRepository) getGitDocsConfigurationFilePath() string {
	return path.Join(repo.workingDir, "git-docs.json")
}

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
