package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

func existsFile(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func detectGitRootdirectory() *string {
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}

	for cur := dir; cur != ""; {
		maybeGitDir := path.Join(cur, ".git")
		if existsFile(maybeGitDir) {
			return &cur
		}

		cur = path.Dir(cur)
	}

	return nil
}

func ensureWorkingSpaceReady(repositoryDir string) bool {
	workingSpaceRoot := path.Join(repositoryDir, ".magic-git")
	if !existsFile(workingSpaceRoot) {
		var err = os.Mkdir(workingSpaceRoot, 0755)
		if err != nil {
			return false
		}

		err = os.Mkdir(path.Join(workingSpaceRoot, "issues"), 0755)
		if err != nil {
			fmt.Printf("ERROR %v\n!\n", err)
			return false
		}

		err = os.Mkdir(path.Join(workingSpaceRoot, "tmp"), 0755)
		if err != nil {
			return false
		}
	}

	return true
}

func main() {
	fmt.Printf("\nmagic-git, welcome\n\n")

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

	ok := ensureWorkingSpaceReady(*gitRepositoryDir)
	if !ok {
		fmt.Printf("ERROR cannot prepare working directory !\n")
		return
	}
	// TODO ensure we have a working git repo and data directory is ready

	// execute the verb
	switch verbs[0] {
	case "serve":
		fmt.Println("starting web server")
		break

	case "issue":
		fmt.Println("issues management")
		break

	case "issues":
		fmt.Println("issues management")
		break
	}

	// parse command line for those actions :
	// * serve -port 8098 -insecure
	// 		=> future options : multi repositories
	// * issue list -remoteUri=local
	// * issue create  => opens a file with a template, creates the files, commit the changes, optionnally push
	// * issue update ISSUE_ID   => same flow : file, modify the issue files, commit the changes, optionnally push

	// option '-remoteUri=local' can be used to talk through the REST api of another magic-git server (http://...)

	// parse command line
	// detect being in a git repository
	// detect the git repository root path
	// ensure .magic-git/ is initialized (issues/, tmp/)
	// instantiate a remote or local (depending on the options) main application code
	// execute action for each mode (serve, issues)
	// - serve listens on TCP
	// - issues and other commands alike use either direct local filesystem or the REST api, if specified in options
	// quit
}
