package main

import (
	objects "BitGit/internal"
	"fmt"
	"os"
)

const (
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: git <command> [args...]")
		fmt.Printf(`
Commands: 
	-init 
	-add 
	-commit 
	-log 
	-checkout 
	-status 
	-branch 
	-switch
`)
		return
	}
	command := os.Args[1]

	switch command {
	case "init":
		invokeInit()
	case "load-object":
		invokeLoadObject()
	default:
		fmt.Printf("%sNot a recognised command%s\n", colorRed, colorReset)
		fmt.Println("Usage: git <command> [args...]")
		fmt.Printf(`
Commands: 
	-init 
	-add 
	-commit 
	-log 
	-checkout 
	-status 
	-branch 
	-switch
`)
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
	fmt.Printf("Object Content: \n\n%s\n", obj.Content())
}
