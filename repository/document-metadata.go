package repository

type DocumentMetadata map[string]interface{}

func (metadata *DocumentMetadata) getTags() []string {
	result := []string{}
	for _, tag := range (*metadata)["tags"].([]interface{}) {
		result = append(result, tag.(string))
	}
	return result
}

func (metadata *DocumentMetadata) addTag(addedTag string) bool {
	for _, tag := range (*metadata)["tags"].([]interface{}) {
		if tag.(string) == addedTag {
			return false
		}
	}

	(*metadata)["tags"] = append((*metadata)["tags"].([]interface{}), interface{}(addedTag))

	return true
}

func (metadata *DocumentMetadata) removeTag(removedTag string) bool {
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
