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

	"github.com/julienschmidt/httprouter"

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

type RenameIssueRequest struct {
	Name string `json:"name"`
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

func handlerWebUi(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	relativePath := p.ByName("requested_resource")
	if strings.HasPrefix(relativePath, "/") {
		relativePath = relativePath[1:]
	}

	rawContentBytes, err := assetsgen.Asset("assets/webui/" + relativePath)
	if err != nil {
		errorResponse(w, 404, fmt.Sprintf("not found '%s'", relativePath))
		return
	}

	content := string(rawContentBytes)
	contentType := "application/octet-stream"

	if strings.HasSuffix(relativePath, ".md") {
		context := &PageContext{
			Name: "First context member",
		}

		contentType = "application/markdown"
		interpolated := interpolate(relativePath, content, context)
		if interpolated != nil {
			content = *interpolated
		}
	} else if strings.HasSuffix(relativePath, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(relativePath, ".js") {
		contentType = "application/javascript"
	} else if strings.HasSuffix(relativePath, ".html") {
		contentType = "text/html"
	}

	w.Header().Set("Content-Type", contentType)
	httpResponse(w, 200, content)
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
	Clean         bool   `json:"clean"`
	Text          string `json:"text"`
	GitRepository string `json:"gitRepository"`
}

func handlerStatusRestAPI(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	status, err := server.magic.GetStatus()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	clean, err := server.magic.IsClean()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	response := StatusResponse{
		Clean:         clean,
		Text:          *status,
		GitRepository: server.magic.GitRepositoryDir(),
	}

	jsonResponse(w, 200, response)
}

func handlerTagsRestAPI(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	tags, err := server.magic.GetAllTags()
	if err != nil {
		errorResponse(w, 500, "internal error")
	} else {
		jsonResponse(w, 200, tags)
	}
}

func handlerGetIssues(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	if r.URL.Query().Get("q") == "" {
		issues, err := server.magic.GetIssues()
		if err != nil {
			errorResponse(w, 500, "internal error")
		} else {
			jsonResponse(w, 200, issues)
		}
	} else {
		issues, err := server.magic.SearchIssues(r.URL.Query().Get("q"))
		if err != nil {
			errorResponse(w, 500, "internal error")
		} else {
			jsonResponse(w, 200, issues)
		}
	}
}

func handlerGetIssueMetadata(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

	metadata, err := server.magic.GetIssueMetadata(name)
	if err != nil {
		errorResponse(w, 404, "not found metadata")
	} else {
		jsonResponse(w, 200, metadata)
	}
}

func handlerGetIssueContent(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

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
}

func handlerDeleteIssue(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

	result, err := server.magic.DeleteIssue(name)
	if err != nil {
		errorResponse(w, 500, fmt.Sprintf("delete error : %v", err))
	} else {
		jsonResponse(w, 200, map[string]bool{"deleted": result})
	}
}

func handlerPostIssue(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

	if server.magic.AddIssue(name) {
		messageResponse(w, "issue added")
	} else {
		errorResponse(w, 500, "error (maybe already exists ?)")
	}
}

func handlerPostIssueRename(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "error in body")
	} else {
		request := &RenameIssueRequest{}

		err = json.Unmarshal(out, request)
		if err != nil {
			errorResponse(w, 400, "error malformatted json")
		} else {
			if server.magic.RenameIssue(name, request.Name) {
				messageResponse(w, "issue renamed")
			} else {
				errorResponse(w, 500, "error")
			}
		}

	}
}

func handlerPutIssueMetadata(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "error in body")
	} else {
		metadata := &repository.IssueMetadata{}

		err = json.Unmarshal(out, metadata)
		if err != nil {
			errorResponse(w, 400, "error malformatted json")
		} else {
			ok, err := server.magic.SetIssueMetadata(name, metadata)
			if err != nil || !ok {
				errorResponse(w, 400, "error setting metadata")
			} else {
				messageResponse(w, "issue metadata updated")
			}
		}

	}
}

func handlerPutIssueContent(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("issue_name")

	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "error in body")
	} else {
		ok, err := server.magic.SetIssueContent(name, string(out))
		if err != nil || !ok {
			errorResponse(w, 400, fmt.Sprintf("error setting content : %v", err))
		} else {
			messageResponse(w, "issue content updated")
		}
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

func makeHandle(handle func(http.ResponseWriter, *http.Request, httprouter.Params, *WebServer), server *WebServer) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handle(w, r, p, server)
	}
}

func (self *WebServer) Init(router *httprouter.Router) {
	router.GET("/webui/*requested_resource", makeHandle(handlerWebUi, self))
	router.GET("/api/status", makeHandle(handlerStatusRestAPI, self))
	router.GET("/api/tags", makeHandle(handlerTagsRestAPI, self))
	router.GET("/api/issues", makeHandle(handlerGetIssues, self))
	router.GET("/api/issues/:issue_name/metadata", makeHandle(handlerGetIssueMetadata, self))
	router.GET("/api/issues/:issue_name/content", makeHandle(handlerGetIssueContent, self))
	router.POST("/api/issues/:issue_name", makeHandle(handlerPostIssue, self))
	router.POST("/api/issues/:issue_name/rename", makeHandle(handlerPostIssueRename, self))
	router.PUT("/api/issues/:issue_name/metadata", makeHandle(handlerPutIssueMetadata, self))
	router.PUT("/api/issues/:issue_name/content", makeHandle(handlerPutIssueContent, self))
	router.DELETE("/api/issues/:issue_name", makeHandle(handlerDeleteIssue, self))
}

/* Run runs the Web server... */
func Run(magic *repository.MagicGitRepository) {
	fmt.Println("starting web server")

	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	server := NewWebServer(magic)
	server.Init(router)

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}
