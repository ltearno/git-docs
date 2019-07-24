package webserver

import "git-docs/repository"

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

type ErrorResponse struct {
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type StatusResponse struct {
	Clean            bool    `json:"clean"`
	Text             string  `json:"text"`
	WorkingDirectory string  `json:"workingDirectory"`
	GitRepository    *string `json:"gitRepository"`
}
