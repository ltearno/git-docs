package webserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"git-docs/repository"
)

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
		if r.URL.Query().Get("interpolated") == "true" {
			context := &DocumentContext{}
			context.Document.Name = name

			interpolated := interpolateTemplate(name, *content, context)
			if interpolated != nil {
				httpResponse(w, 200, "text/markdown", *interpolated)
			} else {
				errorResponse(w, 500, "cannot interpolate")
			}
		} else {
			httpResponse(w, 200, "text/markdown", *content)
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
