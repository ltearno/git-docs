package webserver

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"git-docs/repository"
)

// injects the WebServer context in http-router handler
func (server *WebServer) makeHandler(handler func(http.ResponseWriter, *http.Request, httprouter.Params, *WebServer)) func(http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		handler(w, r, p, server)
	}
}

func (server *WebServer) init(router *httprouter.Router) {
	router.GET("/", server.makeHandler(handlerRedirectHome))
	router.GET("/git-docs/webui/*requested_resource", server.makeHandler(handlerGetWebUI))
	router.GET("/git-docs/api/status", server.makeHandler(handlerGetStatus))
	router.GET("/git-docs/api/tags/:category_name", server.makeHandler(handlerTagsRestAPI))
	router.GET("/git-docs/api/workflows/:category_name", server.makeHandler(handlerGetWorkflow))
	router.GET("/git-docs/api/categories", server.makeHandler(handlerGetCategories))
	router.POST("/git-docs/api/categories/:category_name", server.makeHandler(handlerPostCategories))
	router.GET("/git-docs/api/documents/:category_name", server.makeHandler(handlerGetDocuments))
	router.GET("/git-docs/api/documents/:category_name/:document_name/metadata", server.makeHandler(handlerGetDocumentMetadata))
	router.GET("/git-docs/api/documents/:category_name/:document_name/content", server.makeHandler(handlerGetDocumentContent))
	router.POST("/git-docs/api/documents/:category_name/:document_name", server.makeHandler(handlerPostDocument))
	router.POST("/git-docs/api/documents/:category_name/:document_name/rename", server.makeHandler(handlerPostDocumentRename))
	router.PUT("/git-docs/api/documents/:category_name/:document_name/metadata", server.makeHandler(handlerPutDocumentMetadata))
	router.PUT("/git-docs/api/documents/:category_name/:document_name/content", server.makeHandler(handlerPutDocumentContent))
	router.DELETE("/git-docs/api/documents/:category_name/:document_name", server.makeHandler(handlerDeleteDocument))
}

// Start runs a webserver hosting the GitDocs application
func Start(repo *repository.GitDocsRepository, port int) {
	fmt.Println("starting web server")

	router := httprouter.New()
	if router == nil {
		fmt.Printf("Failed to instantiate the router, exit\n")
	}

	server := &WebServer{
		repo: repo,
	}

	server.init(router)

	fmt.Printf("\n you can use your internet browser to go here : http://127.0.0.1:%d/git-docs/webui/index.html\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), router))
}
