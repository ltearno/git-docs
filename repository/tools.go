package repository

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
)

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

	_, err = file.Write([]byte(content))
	if err != nil {
		return false
	}

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

	return content, nil
}

func tagsContainText(tags []string, q string) bool {
	if q == "" {
		return true
	}

	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}

	return false
}

func tagsMatchSearch(tags []string, q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return true
	} else if strings.HasPrefix(q, "!") {
		q = strings.TrimSpace(q[1:])
		return !tagsMatchSearch(tags, q)
	} else if strings.HasPrefix(q, "&") {
		q = strings.TrimSpace(q[1:])
		separatorPos := strings.Index(q, " ")
		if separatorPos == 0 {
			return false
		}
		return tagsMatchSearch(tags, q[:separatorPos]) && tagsMatchSearch(tags, q[separatorPos+1:])
	} else if strings.HasPrefix(q, "|") {
		q = strings.TrimSpace(q[1:])
		separatorPos := strings.Index(q, " ")
		if separatorPos == 0 {
			return false
		}
		return tagsMatchSearch(tags, q[:separatorPos]) || tagsMatchSearch(tags, q[separatorPos+1:])
	} else {
		return tagsContainText(tags, q)
	}
}
