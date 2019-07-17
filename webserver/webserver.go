package webserver

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"../assetsgen"
	"../repository"
)

type PageContext struct {
	Name string
}

type WebServer struct {
	magic *repository.MagicGitRepository
}

func handler(w http.ResponseWriter, r *http.Request, path string, server *WebServer) {
	context := &PageContext{
		Name: "First context member",
	}

	rawTemplateBytes, err := assetsgen.Asset("assets/" + path)
	if err != nil {
		fmt.Fprintf(w, "Sorry, nothing here (%s)", path)
	} else {
		rawTemplate := string(rawTemplateBytes)
		t, err := template.New(path).Parse(string(rawTemplate))
		if err != nil {
			fmt.Fprintf(w, "Sorry, internal problem")
			fmt.Printf("error cannot parse %s %v", path, err)
		} else {
			t.Execute(w, context)
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

	w.WriteHeader(code)
	w.Write(body)
}

func handlerIssuesRestAPI(w http.ResponseWriter, r *http.Request, path string, server *WebServer) {
	w.Header().Set("Content-Type", "application/json")

	if path == "" && r.Method == http.MethodGet {
		issues := server.magic.Issues()
		jsonResponse(w, 200, issues)
	} else if path != "" && r.Method == http.MethodPost {
		messageResponse(w, "welcome")
	} else {
		errorResponse(w, 404, "not found")
	}
}

func addHandler(path string, fn func(http.ResponseWriter, *http.Request, string, *WebServer), server *WebServer) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len(path):]
		fn(w, r, path, server)
	}

	http.HandleFunc(path, handler)
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
}

/* Run runs the Web server... */
func Run(magic *repository.MagicGitRepository) {
	fmt.Println("starting web server")

	server := NewWebServer(magic)
	server.Init()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
