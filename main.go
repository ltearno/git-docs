package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"./repository"
	"./tools"
	"./webserver"
)

func detectGitRootdirectory() *string {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}

	for cur := dir; cur != ""; {
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

	flag.Parse()

	if *help {
		printUsage = true
	}

	verbs := flag.Args()

	//for i, verb := range verbs {
	//	fmt.Printf("%d %s/", i, verb)
	//}

	if len(verbs) == 0 {
		fmt.Println("not enough parameters, sue '-help' !")
		printUsage = true
	}

	gitRepositoryDir := detectGitRootdirectory()
	if gitRepositoryDir == nil {
		fmt.Println("not in a git repository !")
		printUsage = true
	}

	if printUsage {
		fmt.Printf("\nmgit usage :\n\n  mgit [OPTIONS] verbs...\n\nOPTIONS :\n\n")
		flag.PrintDefaults()
		return
	}

	fmt.Printf(" working in %s\n", *gitRepositoryDir)
	fmt.Println()

	repo := repository.NewGitDocsRepository(gitRepositoryDir, path.Join(*gitRepositoryDir, ".git-docs"))
	ok := repo.EnsureWorkingSpaceReady()
	if !ok {
		fmt.Printf("ERROR cannot prepare working directory !\n")
		return
	}
	// TODO ensure we have a working git repo and data directory is ready

	// execute the verb
	switch verbs[0] {
	case "serve":
		webserver.Run(repo)
		break

	case "document":
		fmt.Println("documents management")
		break

	case "documents":
		fmt.Println("documents management")
		break
	}

	// parse command line for those actions :
	// * serve -port 8098 -insecure
	// 		=> future options : multi repositories
	// * document list -remoteUri=local
	// * document create  => opens a file with a template, creates the files, commit the changes, optionnally push
	// * document update DOCUMENT_ID   => same flow : file, modify the document files, commit the changes, optionnally push

	// option '-remoteUri=local' can be used to talk through the REST api of another git-docs server (http://...)
}
