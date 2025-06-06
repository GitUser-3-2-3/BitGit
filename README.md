# BitGit

A minimal Git-like version control system, implemented in Go.

## Overview

BitGit is a simple command-line tool that mimics core Git functionalities, including repository initialization, staging,
committing, and viewing commit logs. It uses a custom repository and object storage format, supporting blobs, trees, and
commits. I might extend the feature list in the future but no current plans as of yet.

## Features

- **Repository Initialization**:  
  Initialize a new repository with `git init`, creating the necessary `.git` directory structure.

- **Staging Files**:  
  Add files to the staging area using `git add <file>`. Tracks file metadata and prepares them for commit.

- **Committing Changes**:  
  Commit staged changes with `git commit -m <message>`. Stores commit objects with author, message, and timestamp.

- **Viewing Commit Logs**:  
  Display recent commits using `git log`, showing commit hash, author, date, and message.

- **Object Storage**:  
  Stores blobs (file contents), trees (directory structure), and commits as compressed objects in `.git/objects`.

- **Loading Objects**:  
  Inspect raw objects by hash with `git load-object <hash>`.

- **Branch and HEAD Management**:  
  Maintains branch references and HEAD pointer, supporting basic branch tracking.

## Preview

![BigGit Preview](preview%201.png)
![BigGit Preview](preview%202.png)

## Commands

- `git init [path]`  
  Initialize a new repository at the given path (default: current directory).

- `git add <file>`  
  Stage a file for commit.

- `git commit -m <message>`  
  Commit staged changes with a message.

- `git log`  
  Show the latest 10 commits.

- `git load-object <hash>`  
  Display the type and content of a stored object by its hash.

## Project Structure

- `cmd/main.go`  
  CLI entry point and command dispatch.

- `internal/repository.go`  
  Repository logic: initialization, object storage, staging, commit, log, and branch management.

- `internal/object.go`  
  Object definitions: blobs, trees, commits, and their serialization.

## Getting Started

1. **Clone the repository**
2. **Open in the ide of your choice or navigate to the project directory with the terminal**
3. **Run application with `go run ./cmd/ <command>` from the project directory**
