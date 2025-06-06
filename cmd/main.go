package main

import (
	objects "BitGit/internal"
	"fmt"
	"os"
)

var commands = `
Commands:
	-init 
	-add 
	-commit 
	-log 
	-checkout 
	-status 
	-branch 
	-switch
`

const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: git <command> [args...]")
		fmt.Printf(commands)
		return
	}
	command := os.Args[1]

	switch command {
	case "init":
		invokeInit()
	case "load-object":
		invokeLoadObject()
	case "add":
		invokeAdd()
	case "commit":
		invokeCommit()
	case "log":
		invokeLog()
	default:
		fmt.Printf("%sNot a recognised command%s\n", colorRed, colorReset)
		fmt.Println("Usage: git <command> [args...]")
		fmt.Printf(commands)
		return
	}
}

func invokeInit() {
	path := "."
	if len(os.Args) > 2 {
		path = os.Args[2]
	}
	repo, err := objects.InitRepo(path)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		return
	}
	fmt.Printf("Initialized an empty git repository in %s\n", repo.GitDir)
}

func invokeLoadObject() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: git load-object <hash>")
		return
	}
	hash := os.Args[2]

	repo, err := objects.LoadRepo(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	obj, err := repo.LoadObject(hash)
	if err != nil {
		fmt.Printf("Error loading object: %v\n", err)
		return
	}
	fmt.Printf("Object Type: %s\n\n", obj.Type())
	bytes, err := obj.Content()
	if err != nil {
		return
	}
	fmt.Printf("Object Content: \n\n%s\n", bytes)
}

func invokeAdd() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: git add <file>")
		return
	}
	repo, err := objects.LoadRepo(".")
	if err != nil {
		fmt.Printf("%sError:%s %v\n", colorRed, colorReset, err)
		return
	}
	if err = repo.Add(os.Args[2]); err != nil {
		fmt.Printf("%sError:%s %v\n", colorRed, colorReset, err)
		return
	}
	fmt.Printf("Added %s\n", os.Args[2])
}

func invokeCommit() {
	if len(os.Args) < 4 || os.Args[2] != "-m" {
		fmt.Println("Usage: git commit -m <message>")
		return
	}
	repo, err := objects.LoadRepo(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	message := os.Args[3]
	author := "User <parth@gmail.com>"

	hash, err := repo.Commit(message, author)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Created commit %s\n", hash[:8])
}

func invokeLog() {
	repo, err := objects.LoadRepo(".")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	commits, err := repo.Log(10)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	for _, commit := range commits {
		commitHash, err := commit.Hash()
		if err != nil {
			return
		}
		fmt.Printf("commit %s\n", commitHash)
		fmt.Printf("Author: %s\n", commit.Author)
		fmt.Printf("Date: %s\n", commit.Timestamp.Format("Mon Jan 2 15:04:05 2006"))
		fmt.Printf("\n    %s\n\n", commit.Message)
	}
}
