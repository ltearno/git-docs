package repository

import (
	"bufio"
	"encoding/json"
	"fmt"
	"git-docs/assetsgen"
	"git-docs/tools"
	"io/ioutil"
	"os"
	"os/exec"
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
