package webserver

import (
	"bytes"
	"html/template"
	"io/ioutil"
)

func interpolateTemplate(name string, templateContent string, context interface{}) *string {
	t, err := template.New(name).Parse(templateContent)
	if err != nil {
		return nil
	}

	buffer := bytes.NewBufferString("")

	t.Execute(buffer, context)

	out, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil
	}

	result := string(out)

	return &result
}
