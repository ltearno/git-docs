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

	"git-docs/assetsgen"
	"git-docs/repository"
)

type PageContext struct {
	Name string
}

type DocumentContext struct {
	Document struct {
		Name string
	}
}

type WebServer struct {
	repo *repository.GitDocsRepository
}

type RenameDocumentRequest struct {
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
	status, err := server.repo.GetStatus()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	clean, err := server.repo.IsClean()
	if err != nil {
		errorResponse(w, 500, "internal error")
	}

	response := StatusResponse{
		Clean:         clean,
		Text:          *status,
		GitRepository: *server.repo.GitRepositoryDir(),
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

func handlerGetDocuments(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")

	if r.URL.Query().Get("q") == "" {
		documents, err := server.repo.GetDocuments(category)
		if err != nil {
			errorResponse(w, 500, "internal error")
		} else {
			jsonResponse(w, 200, documents)
		}
	} else {
		documents, err := server.repo.SearchDocuments(category, r.URL.Query().Get("q"))
		if err != nil {
			errorResponse(w, 500, "internal error")
		} else {
			jsonResponse(w, 200, documents)
		}
	}
}

func handlerGetDocumentMetadata(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")

	metadata, err := server.repo.GetDocumentMetadata(category, name)
	if err != nil {
		errorResponse(w, 404, "not found metadata")
	} else {
		jsonResponse(w, 200, metadata)
	}
}

func handlerGetDocumentContent(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")

	content, err := server.repo.GetDocumentContent(category, name)
	if err != nil {
		errorResponse(w, 404, "not found content")
	} else {
		w.Header().Set("Content-Type", "text/markdown")
		if r.URL.Query().Get("interpolated") == "true" {
			context := &DocumentContext{}
			context.Document.Name = name

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

func handlerDeleteDocument(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")

	result, err := server.repo.DeleteDocument(category, name)
	if err != nil {
		errorResponse(w, 500, fmt.Sprintf("delete error : %v", err))
	} else {
		jsonResponse(w, 200, map[string]bool{"deleted": result})
	}
}

func handlerPostDocument(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")

	if server.repo.AddDocument(category, name) {
		messageResponse(w, "document added")
	} else {
		errorResponse(w, 500, "error (maybe already exists ?)")
	}
}

func handlerPostDocumentRename(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")

	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "error in body")
	} else {
		request := &RenameDocumentRequest{}

		err = json.Unmarshal(out, request)
		if err != nil {
			errorResponse(w, 400, "error malformatted json")
		} else {
			if server.repo.RenameDocument(category, name, request.Name) {
				messageResponse(w, "document renamed")
			} else {
				errorResponse(w, 500, "error")
			}
		}

	}
}

func handlerPutDocumentMetadata(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")
	actionName := r.URL.Query().Get("action_name")

	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "error in body")
	} else {
		metadata := &repository.DocumentMetadata{}

		err = json.Unmarshal(out, metadata)
		if err != nil {
			errorResponse(w, 400, "error malformatted json")
		} else {
			ok, err := server.repo.SetDocumentMetadata(category, name, metadata, &actionName)
			if err != nil || !ok {
				errorResponse(w, 400, "error setting metadata")
			} else {
				messageResponse(w, "document metadata updated")
			}
		}

	}
}

func handlerPutDocumentContent(w http.ResponseWriter, r *http.Request, p httprouter.Params, server *WebServer) {
	category := p.ByName("category_name")
	name := p.ByName("document_name")

	out, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errorResponse(w, 400, "error in body")
	} else {
		ok, err := server.repo.SetDocumentContent(category, name, string(out))
		if err != nil || !ok {
			errorResponse(w, 400, fmt.Sprintf("error setting content : %v", err))
		} else {
			messageResponse(w, "document content updated")
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

func NewWebServer(repo *repository.GitDocsRepository) *WebServer {
	return &WebServer{
		repo: repo,
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
	router.GET("/api/tags/:category_name", makeHandle(handlerTagsRestAPI, self))
	router.GET("/api/workflows/:category_name", makeHandle(handlerGetWorkflow, self))
	router.GET("/api/categories", makeHandle(handlerGetCategories, self))
	router.POST("/api/categories/:category_name", makeHandle(handlerPostCategories, self))
	router.GET("/api/documents/:category_name", makeHandle(handlerGetDocuments, self))
	router.GET("/api/documents/:category_name/:document_name/metadata", makeHandle(handlerGetDocumentMetadata, self))
	router.GET("/api/documents/:category_name/:document_name/content", makeHandle(handlerGetDocumentContent, self))
	router.POST("/api/documents/:category_name/:document_name", makeHandle(handlerPostDocument, self))
	router.POST("/api/documents/:category_name/:document_name/rename", makeHandle(handlerPostDocumentRename, self))
	router.PUT("/api/documents/:category_name/:document_name/metadata", makeHandle(handlerPutDocumentMetadata, self))
	router.PUT("/api/documents/:category_name/:document_name/content", makeHandle(handlerPutDocumentContent, self))
	router.DELETE("/api/documents/:category_name/:document_name", makeHandle(handlerDeleteDocument, self))
}

// Run runs a webserver hosting the GitDocs application
func Run(repo *repository.GitDocsRepository) {
	fmt.Println("starting web server")

	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	server := NewWebServer(repo)
	server.Init(router)

	fmt.Println("\n you can use your internet browser to go here : http://127.0.0.1:8080/webui/index.html")

	log.Fatal(http.ListenAndServe("127.0.0.1:8080", router))
}
