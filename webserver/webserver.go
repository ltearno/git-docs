package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"../assetsgen"
	"../repository"
)

type PageContext struct {
	Name string
}

type IssueContext struct {
	Issue struct {
		Name string
	}
}

type WebServer struct {
	magic *repository.MagicGitRepository
}

func interpolate(name string, templateContent string, context interface{}) *string {
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

func handler(w http.ResponseWriter, r *http.Request, relativePath string, server *WebServer) {
	context := &PageContext{
		Name: "First context member",
	}

	rawTemplateBytes, err := assetsgen.Asset("assets/" + relativePath)
	if err != nil {
		fmt.Fprintf(w, "Sorry, nothing here (%s)", relativePath)
	} else {
		rawTemplate := string(rawTemplateBytes)

		interpolated := interpolate(relativePath, rawTemplate, context)
		if interpolated != nil {
			httpResponse(w, 200, *interpolated)
		} else {
			httpResponse(w, 200, rawTemplate)
		}
	}
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

func messageResponse(w http.ResponseWriter, message string) {
	jsonResponse(w, 200, MessageResponse{message})
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	jsonResponse(w, code, ErrorResponse{message})
}

func jsonResponse(w http.ResponseWriter, code int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")

	body, err := json.Marshal(value)
	if err != nil {
		fmt.Fprintf(w, "{ \"message\": \"cannot marshall JSON\" }")
		return
	}

	httpResponse(w, code, string(body))
}

func httpResponse(w http.ResponseWriter, code int, body string) {
	w.WriteHeader(code)
	w.Write([]byte(body))
}

type StatusResponse struct {
	Clean bool   `json:"clean"`
	Text  string `json:"text"`
}

func handlerStatusRestAPI(w http.ResponseWriter, r *http.Request, relativePath string, server *WebServer) {
	status, err := server.magic.GetStatus()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	clean, err := server.magic.IsClean()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	response := StatusResponse{
		Clean: clean,
		Text:  *status,
	}

	jsonResponse(w, 200, response)
}

func handlerIssuesRestAPI(w http.ResponseWriter, r *http.Request, relativePath string, server *WebServer) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodGet {
		if relativePath == "" {
			issues := server.magic.Issues()
			jsonResponse(w, 200, issues)
		} else {
			if strings.HasSuffix(relativePath, "/metadata") {
				name := relativePath[0 : len(relativePath)-len("/metadata")]
				metadata, err := server.magic.GetIssueMetadata(name)
				if err != nil {
					errorResponse(w, 404, "not found metadata")
				} else {
					jsonResponse(w, 200, metadata)
				}
			} else if strings.HasSuffix(relativePath, "/content") {
				name := relativePath[0 : len(relativePath)-len("/content")]
				content, err := server.magic.GetIssueContent(name)
				if err != nil {
					errorResponse(w, 404, "not found content")
				} else {
					w.Header().Set("Content-Type", "text/markdown")
					if r.URL.Query().Get("interpolated") == "true" {
						context := &IssueContext{}
						context.Issue.Name = name

						interpolated := interpolate(name, *content, context)
						if interpolated != nil {
							httpResponse(w, 200, *interpolated)
						} else {
							errorResponse(w, 500, "cannot interpolate")
						}
					} else {
						httpResponse(w, 200, *content)
					}
				}
			} else {
				errorResponse(w, 404, "not found, or invalid path")
			}
		}
	} else if relativePath != "" && r.Method == http.MethodPost {
		if server.magic.AddIssue(relativePath) {
			messageResponse(w, "issue added")
		} else {
			errorResponse(w, 500, "error (maybe already exists ?)")
		}
	} else {
		errorResponse(w, 404, "not found")
	}
}

func addHandler(pathPrefix string, fn func(http.ResponseWriter, *http.Request, string, *WebServer), server *WebServer) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		pathPrefix := r.URL.Path[len(pathPrefix):]
		fn(w, r, pathPrefix, server)
	}

	http.HandleFunc(pathPrefix, handler)
}

func NewWebServer(magic *repository.MagicGitRepository) *WebServer {
	return &WebServer{
		magic: magic,
	}
}

func (self *WebServer) Init() {
	addHandler("/webui/", handler, self)
	addHandler("/api/issues", handlerIssuesRestAPI, self)
	addHandler("/api/issues/", handlerIssuesRestAPI, self)
	addHandler("/api/status", handlerStatusRestAPI, self)
}

/* Run runs the Web server... */
func Run(magic *repository.MagicGitRepository) {
	fmt.Println("starting web server")

	server := NewWebServer(magic)
	server.Init()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
