package main

import (
	"encoding/json"
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
	AuditContent []byte
}

func (ad *AuditData) auditFileCheck() []byte {
	ad.AuditContent, _ = ioutil.ReadFile(filePath)
	return ad.AuditContent
}

type LogEntry struct {
	Time         string `json:"time"`
	Auth         struct {
		DisplayName string `json:"display_name"`
	} `json:"auth"`
	Request struct {
		Operation string `json:"operation"`
	} `json:"request"`
	Response struct {
		Data map[string]interface{} `json:"data"`
	} `json:"response"`
}

func logHandler() {
	audit := &AuditData{}
	audit.auditFileCheck()

	var logEntry LogEntry
	if err := json.Unmarshal(audit.AuditContent, &logEntry); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return
	}

	// Log the filtered data
	log.Printf("Time: %s", logEntry.Time)
	log.Printf("Display Name: %s", logEntry.Auth.DisplayName)
	log.Printf("Operation: %s", logEntry.Request.Operation)
	log.Printf("Response Data: %v", logEntry.Response.Data)

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
