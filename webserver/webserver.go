package webserver

import (
	"fmt"
	"log"
	"net/http"

	"../repository"
)

func handler(w http.ResponseWriter, r *http.Request) {
	//name := r.URL.Path[len("/webui/"):]

	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

/* Run runs the Web server... */
func Run(magic *repository.MagicGitRepository) {
	fmt.Println("starting web server")

	http.HandleFunc("/webui/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
