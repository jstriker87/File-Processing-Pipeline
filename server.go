package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

const maxUploadSize = 1 * 1024 * 1024 // 0.5 mb

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

var users []User

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/users", getUsers).Methods("GET")
	router.HandleFunc("/users", createUser).Methods("POST")
	router.HandleFunc("/upload", HandleFileUpload).Methods("POST")
	router.HandleFunc("/status", getDirectoryData).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func HandleFileUpload(w http.ResponseWriter, r *http.Request) {

	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		return
	}


	fileType := r.PostFormValue("type")
	fmt.Println(fileType)
	file, handler, err := r.FormFile("uploadFile")
	if err != nil {
		http.Error(w, "Error with provided file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if handler.Size > maxUploadSize {
		http.Error(w, "File size of " + strconv.FormatInt(handler.Size,10)  + " bytes is bigger than the allowed size of " + strconv.Itoa(maxUploadSize) + " bytes", http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(file)

	if err != nil {
		http.Error(w, "Error with provided file: "+err.Error(), http.StatusBadRequest)
		return
	}

	FileType := http.DetectContentType(fileBytes)
	fileEx, err := mime.ExtensionsByType(FileType)

	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	fmt.Printf("File type: %+v\n", fileEx[0])

	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	finalPath := filepath.Join(uploadPath + handler.Filename)

	_ , staterr := os.Stat(finalPath)
	if os.IsNotExist(staterr) {
	} else {
		http.Error(w, "Error: file Exists", http.StatusBadRequest)
		return
	}
	newFile, err := os.Create(finalPath)

	_, err = newFile.Write(fileBytes)

	if err != nil {
		http.Error(w, "Error: Unable to save file: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("SUCCESS"))

}


func getDirectoryData(w http.ResponseWriter, r *http.Request) {
	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"
	root := os.DirFS(uploadPath)
	allFiles, _ := fs.Glob(root, "*.*")

	var files []string
    for _, f := range allFiles {
       files = append(files, path.Join(uploadPath, f))
	   fmt.Println(files)
    }


}

func getUsers(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON "+err.Error(), http.StatusBadRequest)
		return
	}

	if newUser.ID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if newUser.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	if newUser.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	users = append(users, newUser)
	err = json.NewEncoder(w).Encode(newUser)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
