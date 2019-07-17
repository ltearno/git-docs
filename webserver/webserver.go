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

func handler(w http.ResponseWriter, r *http.Request, path string) {
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

func addHandler(path string, fn func(http.ResponseWriter, *http.Request, string)) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len(path):]
		fn(w, r, path)
	}

	http.HandleFunc(path, handler)
}

/* Run runs the Web server... */
func Run(magic *repository.MagicGitRepository) {
	fmt.Println("starting web server")

	addHandler("/webui/", handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
