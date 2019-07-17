package webserver

import (
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

func handler(w http.ResponseWriter, r *http.Request) {
	context := &PageContext{
		Name: "First context member",
	}

	name := r.URL.Path[len("/webui/"):]

	rawTemplateBytes, err := assetsgen.Asset("assets/" + name)
	if err != nil {
		fmt.Fprintf(w, "Sorry, nothing here (%s)", name)
	} else {
		rawTemplate := string(rawTemplateBytes)
		t, err := template.New(name).Parse(string(rawTemplate))
		if err != nil {
			fmt.Fprintf(w, "Sorry, internal problem")
			fmt.Printf("error cannot parse %s %v", name, err)
		} else {
			t.Execute(w, context)
		}
	}
}

/* Run runs the Web server... */
func Run(magic *repository.MagicGitRepository) {
	fmt.Println("starting web server")

	http.HandleFunc("/webui/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
