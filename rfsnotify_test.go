package rfsnotify

import (
	"log"
	"os"
	"testing"

	"github.com/fsnotify/fsnotify"

	"github.com/stretchr/testify/assert"
)

func Test_Watcher_CreateAndRemove(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		log.Fatalf("Could not create the watcher : %v", err)
	}
	defer watcher.Close()

	err = watcher.AddRecursive(".", nil)
	if err != nil {
		log.Fatalf("Could not add the folder : %v", err)
	}

	// Test creation

	if file, err := os.Create("temp.txt"); err != nil {
		log.Fatalf("Could not create the file 'temp.txt' : %v", err)
	} else {
		defer file.Close()
	}

	event := <-watcher.Events

	assert.Equal(t, fsnotify.Create, event.Op, "Invalid event type")

	// Test Remove

	if err = os.Remove("temp.txt"); err != nil {
		log.Fatalf("Could not remove the file 'temp.txt' : %v", err)
	}

	event = <-watcher.Events

	assert.Equal(t, fsnotify.Remove, event.Op, "Invalid event type")

	if err := os.Mkdir("a", os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func Test_Watcher_CreateAndRemoveFolder(t *testing.T) {
	watcher, err := NewWatcher()
	if err != nil {
		log.Fatalf("Could not create the watcher : %v", err)
	}
	defer watcher.Close()

	err = watcher.AddRecursive(".", nil)
	if err != nil {
		log.Fatalf("Could not add the folder : %v", err)
	}

	// Folder creation
	if err := os.Mkdir("test", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	event := <-watcher.Events

	assert.Equal(t, fsnotify.Create, event.Op, "Invalid event type")
	assert.Equal(t, "./test", event.Name, "Invalid event path")

	// Test creation

	if file, err := os.Create("test/temp.txt"); err != nil {
		log.Fatalf("Could not create the file 'test/temp.txt' : %v", err)
	} else {
		defer file.Close()
	}

	event = <-watcher.Events

	assert.Equal(t, fsnotify.Create, event.Op, "Invalid event type")
	assert.Equal(t, "test/temp.txt", event.Name, "Invalid event path")

	// Test Remove

	if err = os.Remove("test/temp.txt"); err != nil {
		log.Fatalf("Could not remove the file 'test/temp.txt' : %v", err)
	}

	event = <-watcher.Events

	assert.Equal(t, fsnotify.Remove, event.Op, "Invalid event type")
	assert.Equal(t, "test/temp.txt", event.Name, "Invalid event path")

	// Folder remove

	if err = os.Remove("test"); err != nil {
		log.Fatalf("Could not remove the folder 'test' : %v", err)
	}

	event = <-watcher.Events

	assert.Equal(t, fsnotify.Remove, event.Op, "Invalid event type")
	assert.Equal(t, "./test", event.Name, "Invalid event path")

}
