package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

var userHomeDir string
var docsDir string
var mediaDir string
var exeDir string
var compressedDir string
var codeDir string

func setup() {
	var err error
	log.Println("setup called")
	userHomeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	docsDir = filepath.Join(userHomeDir, "Downloads", "Documents")
	mediaDir = filepath.Join(userHomeDir, "Downloads", "Media")
	exeDir = filepath.Join(userHomeDir, "Downloads", "Executables")
	compressedDir = filepath.Join(userHomeDir, "Downloads", "Compressed")
	codeDir = filepath.Join(userHomeDir, "Downloads", "Code-Files")

	log.Println("Home directory: ", userHomeDir)
	log.Println("-----------------------------------")
	log.Println("Following directories will be created if not already present inside Downloads directory")
	log.Println("-----------------------------------")
	log.Println("Documents Directory", docsDir)
	log.Println("Media Directory", mediaDir)
	log.Println("Exe Directory", exeDir)
	log.Println("Compressed Directory", compressedDir)
	log.Println("Code Directory", codeDir)
	log.Println("-----------------------------------")

	// create directories if not present
	createDirectory(docsDir)
	createDirectory(mediaDir)
	createDirectory(exeDir)
	createDirectory(compressedDir)
	createDirectory(codeDir)
}

func main() {
	// create setup first!
	setup()
	// create a new watcher for main download directory, can be any directory or file
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	// close watcher when the program ends
	defer watcher.Close()

	// create a channel to receive eve
	done := make(chan bool)

	// go routine to check any event coming into the channel
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("Event: ", event)

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					log.Println("Modified a file: ", event.Name)
					// move the file if we get the corresponding path for it
					destPath := getPathByFileType(event.Name)
					if destPath != "" {
						filename := filepath.Base(event.Name)
						newLoc := filepath.Join(destPath, filename)
						if moveFile(event.Name, newLoc) {
							log.Println("file moved to :", newLoc)
						} else {
							log.Println("Unable to move the file.")
						}
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error: ", err)
			}
		}
	}()

	dlPath := filepath.Join(userHomeDir, "Downloads")
	err = watcher.Add(dlPath)
	if err != nil {
		log.Fatal("Oops watcher adding error: ", err)
	}
	<-done
}

// move a file from source to location
func moveFile(source, destination string) bool {
	err := os.Rename(source, destination)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

// based on the file type get the path where the file should be moved to
func getPathByFileType(filename string) string {
	fileExt := strings.ToLower(filepath.Ext(filename))
	log.Println("File extension: ", fileExt)

	var docsExt = []string{".doc", ".docx", ".xlsx", ".xls", ".ppt", ".pptx", ".rtf", ".pdf", ".potx", ".msg", ".csv"}
	var mediaExt = []string{".jpg", ".jpeg", ".png", ".gif", ".mp3", ".mp4", ".wmv", ".wav", ".arf"}
	var exeExt = []string{".exe", ".msi"}
	var compressedExt = []string{".zip", ".tar", ".7zip", ".rar", ".gz"}
	var codeExt = []string{".go", ".js", ".sql", ".pkb", ".pks", ".java", ".c", ".cpp", ".sh", ".xml", ".html", ".xsl", ".xaml", ".json", ".jar", ".prog"}

	if fileExt != "" {
		if contains(docsExt, fileExt) {
			return docsDir
		} else if contains(mediaExt, fileExt) {
			return mediaDir
		} else if contains(exeExt, fileExt) {
			return exeDir
		} else if contains(compressedExt, fileExt) {
			return compressedDir
		} else if contains(codeExt, fileExt) {
			return codeDir
		}
	}

	return ""
}

func contains(extList []string, ext string) bool {
	for _, e := range extList {
		if ext == e {
			return true
		}
	}
	return false
}

// CreateDirectory creates a directory if not exists.
func createDirectory(dirName string) {
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dirName, 0755)
		if errDir != nil {
			log.Fatal("Failed to create directory: ", errDir)
			return
		}
	}
}
