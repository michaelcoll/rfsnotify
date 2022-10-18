package main

import (
	"fmt"
	"log"
	"os"

	"github.com/michaelcoll/rfsnotify"
)

func main() {

	fmt.Println("Creating watcher ...")
	watcher, err := rfsnotify.NewBufferedWatcher(100)
	if err != nil {
		log.Fatalf("Could not create the watcher : %v", err)
	}
	defer watcher.Close()

	fmt.Println("Adding a folder ...")
	err = watcher.AddRecursive(".", dirFilter)
	if err != nil {
		log.Fatalf("Could not add the folder : %v", err)
	}

	fmt.Println("Watching for file changes ...")
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Println("event:", event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

// dirFilter permits watch on all folder except the `.git` folder
func dirFilter(path string, info os.FileInfo) bool {
	return path != ".git"
}
