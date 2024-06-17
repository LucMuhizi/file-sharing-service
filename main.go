package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	ensureStorageDir("data")
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/files/", downloadHandler)
	http.HandleFunc("/files", listFilesHandler)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// Log message to indicate where the server is listening
	log.Println("Server starting on port 8080...")

	// Start the server on port 8080 and log if there's an error
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// sendErrorResponse writes a standard error response with a given status code and message.
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	jsonResponse, _ := json.Marshal(map[string]string{"error": message})
	w.Write(jsonResponse)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendErrorResponse(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		sendErrorResponse(w, "Error parsing upload form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		sendErrorResponse(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := saveFile(handler.Filename, file); err != nil {
		sendErrorResponse(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully: " + handler.Filename))
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[len("/files/"):]
	if filename == "" || strings.Contains(filename, "..") {
		sendErrorResponse(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filepath := "data/" + filename
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			sendErrorResponse(w, "File not found", http.StatusNotFound)
		} else {
			sendErrorResponse(w, "Error accessing file", http.StatusInternalServerError)
		}
		return
	}

	if fileInfo.IsDir() {
		sendErrorResponse(w, "Requested resource is a directory, not a file", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filepath)
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	storageDir := "data/"
	files, err := ioutil.ReadDir(storageDir)
	if err != nil {
		sendErrorResponse(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	jsonResponse, err := json.Marshal(fileNames)
	if err != nil {
		sendErrorResponse(w, "Error generating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func ensureStorageDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0755)
	}
}

func saveFile(filename string, file io.Reader) error {
	filePath := filepath.Join("data", filename)
	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the file content to the new location
	_, err = io.Copy(dst, file)
	return err
}
