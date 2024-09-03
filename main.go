package main

import (
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

var filePath = os.Getenv("AUDIT_FILE_PATH")

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type AuditData struct {
	auditContent []byte
}

func (ad *AuditData) auditFileCheck() []byte {
	ad.auditContent, _ = ioutil.ReadFile(filePath)
	return ad.auditContent
}

func logHandler() {
	audit := &AuditData{}
	audit.auditFileCheck()

	// Log the Audit Data as STDOUT
	log.Println(string(audit.auditContent))

	// Truncate the file content after it is being alerted
	if err := os.Truncate(filePath, 0); err != nil {
		check(err)
	}
}

// watchFile monitors changes on the filePath, logs them as STDOUT, and truncates the file content
func watchFile() {
	watcher, err := fsnotify.NewWatcher()
	check(err)
	defer watcher.Close()

	var debounceTimer *time.Timer

	// starting to listen on events
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					// Debounce with a 1-second delay before calling logHandler
					debounceTimer = time.AfterFunc(1*time.Second, logHandler)
				}
			case err := <-watcher.Errors:
				check(err)
			}
		}
	}()

	err = watcher.Add(filePath)
	check(err)
	<-make(chan struct{}) // Block forever
}

func main() {
	watchFile()
}
