package main

import (
	"flag"
	"fmt"
	"path"
	"path/filepath"

	"git-docs/repository"
	"git-docs/tools"
	"git-docs/webserver"
)

func detectGitRootdirectory(dir string) *string {
	for cur := dir; cur != "/"; {
		maybeGitDir := path.Join(cur, ".git")
		if tools.ExistsFile(maybeGitDir) {
			return &cur
		}

		cur = path.Dir(cur)
	}

	return nil
}

func main() {
	fmt.Printf("\ngit-docs, welcome\n\n")

	var printUsage = false
	var help = flag.Bool("help", false, "show this help")
	var port = flag.Int("port", 8080, "webserver listening port")

	flag.Parse()

	if *help {
		printUsage = true
	}

	verbs := flag.Args()

	if len(verbs) == 0 {
		fmt.Println("not enough parameters, use '-help' !")
		printUsage = true
	}

	printHelp := func() {
		fmt.Printf("\ngit-docs usage :\n\n  git-docs [OPTIONS] verbs...\n\nOPTIONS :\n\n")
		flag.PrintDefaults()
	}

	if printUsage {
		printHelp()
		return
	}

	// execute the verb
	switch verbs[0] {
	case "serve":
		relativeWorkdir := ".git-docs"
		if len(verbs) > 1 {
			relativeWorkdir = verbs[1]
		}

		workingDir, err := filepath.Abs(relativeWorkdir)
		if err != nil {
			fmt.Printf("cannot find working directory, abort (%v)\n", err)
			return
		}

		fmt.Printf("content directory: %s\n", workingDir)

		gitRepositoryDir := detectGitRootdirectory(workingDir)
		if gitRepositoryDir == nil {
			fmt.Println("not working with git repository")
		} else {
			fmt.Printf("working with git repository: %s\n", *gitRepositoryDir)
		}

		fmt.Println()

		repo := repository.NewGitDocsRepository(gitRepositoryDir, workingDir)
		webserver.Start(repo, *port)
		break

	default:
		printHelp()
	}

	// parse command line for those actions :
	// * serve -port 8098 -insecure
	// 		=> future options : multi repositories
	// * document list -remoteUri=local
	// * document create  => opens a file with a template, creates the files, commit the changes, optionnally push
	// * document update DOCUMENT_ID   => same flow : file, modify the document files, commit the changes, optionnally push

	// option '-remoteUri=local' can be used to talk through the REST api of another git-docs server (http://...)
}
