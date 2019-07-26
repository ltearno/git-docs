package webserver

import (
	"fmt"
	"git-docs/assetsgen"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func handlerRedirectHome(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	w.Header().Set("Location", "/git-docs/webui/index.html")
	w.WriteHeader(301)
	w.Write([]byte(""))
}

func handlerGetWebUI(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
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
		interpolated := interpolateTemplate(relativePath, content, context)
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

	httpResponse(w, 200, contentType, content)
}

func handlerGetStatus(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	status, err := server.repo.GetStatus()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	clean, err := server.repo.IsClean()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	response := StatusResponse{
		Clean:            clean,
		Text:             *status,
		WorkingDirectory: server.repo.GetWorkingDir(),
		GitRepository:    server.repo.GitRepositoryDir(),
	}

	jsonResponse(w, 200, response)
}

func handlerTagsRestAPI(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")

	tags, err := server.repo.GetAllTags(category)
	if err != nil {
		errorResponse(w, 500, "internal error")
	} else {
		jsonResponse(w, 200, tags)
	}
}

func handlerGetWorkflow(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")

	workflowConfiguration, err := server.repo.GetWorkflow(category)
	if err != nil {
		errorResponse(w, 500, "internal error")
	} else {
		jsonResponse(w, 200, workflowConfiguration)
	}
}

func handlerGetCategories(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	categories := server.repo.GetCategories()

	jsonResponse(w, 200, categories)
}

func handlerPostCategories(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	name := p.ByName("category_name")

	ok, err := server.repo.AddCategory(name)
	if err != nil {
		errorResponse(w, 500, "cannot create category")
		return
	}

	if ok {
		messageResponse(w, "category created")
	} else {
		messageResponse(w, "category cannot be created")
	}
}
