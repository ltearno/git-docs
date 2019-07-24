package repository

type WorkflowElement struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Condition   *string  `json:"condition"`
	AddTags     []string `json:"addTags"`
	RemoveTags  []string `json:"removeTags"`
}

type WorkflowConfiguration map[string][]WorkflowElement

func executeWorkflow(config *WorkflowElement, currentMetadata *DocumentMetadata, metadata *DocumentMetadata) {
	if config.Condition == nil || tagsMatchSearch(currentMetadata.getTags(), *config.Condition) {
		for _, tagToAdd := range config.AddTags {
			metadata.addTag(tagToAdd)
		}
		for _, tagToRemove := range config.RemoveTags {
			metadata.removeTag(tagToRemove)
		}
	}
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
