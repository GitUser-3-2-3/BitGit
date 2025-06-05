package main

import (
	objects "BitGit/internal"
	"flag"
	"fmt"
	"os"
)

func main() {
	var choice string
	flag.StringVar(&choice,
		"Initialize .git File", "init", "initializes a git repository")
	flag.Parse()

	switch choice {
	case "init":
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
}
