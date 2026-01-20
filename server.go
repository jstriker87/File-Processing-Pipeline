package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	//"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const maxUploadSize = 1 * 1024 * 1024 // 1 mb

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type Desc struct {
	Description string `json:"description"`
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

	//var data Desc
	//err := json.NewDecoder(r.Body).Decode(&data)
	//if err != nil {
	//	log.Println("ererer" + err.Error())
	//	http.Error(w, "Invalid JSON "+err.Error(), http.StatusBadRequest)
	//	return
	//}

	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		return
	}

	fileType := r.PostFormValue("type")
	fmt.Println(fileType)
	uploadFile, handler, err := r.FormFile("uploadFile")
	//descriptionJson := r.FormValue("description")

	if err != nil {
		http.Error(w, "Error with provided file: "+err.Error(), http.StatusBadRequest)
		return
	}

	defer uploadFile.Close()

	if handler.Size > maxUploadSize {
		folderSize := fmt.Sprintf("%.2f", float64(handler.Size)/(1024*1024))
		http.Error(w, "File size of "+folderSize+" MB is bigger than the allowed size of "+strconv.Itoa(maxUploadSize/(1024*1024))+" MB", http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(uploadFile)

	if err != nil {
		http.Error(w, "Error with provided file: "+err.Error(), http.StatusBadRequest)
		return
	}

	//FileType := http.DetectContentType(fileBytes)
	//fileEx, err := mime.ExtensionsByType(FileType)

	//fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	//fmt.Printf("File Size: %+v\n", handler.Size)
	//fmt.Printf("MIME Header: %+v\n", handler.Header)
	//fmt.Printf("File type: %+v\n", fileEx[0])

	fileName_no_extension := strings.Split(handler.Filename, ".")[0]
	fileName_no_extension = regexp.MustCompile(`[^a-zA-Z0-9 ]+`).ReplaceAllString(fileName_no_extension, "")

	if _, err := os.Stat(filepath.Join("uploads",fileName_no_extension)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Join("uploads", fileName_no_extension), 0755)

		if err != nil {
			http.Error(w, "Error creating folder: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	finalPath := filepath.Join(uploadPath, fileName_no_extension + "/" + handler.Filename)

	_, err = os.Stat(finalPath)
	if os.IsNotExist(err) {
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
	var totalSize int64
	err := filepath.Walk(uploadPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		json.NewEncoder(w).Encode(files)
		return
	}

	folderSize := fmt.Sprintf("Total folder: %.1f MB", float64(totalSize)/(1024*1024))
	files = append(files, folderSize)

	json.NewEncoder(w).Encode(files)

}

func AddDescription(w http.ResponseWriter, r *http.Request) {
	var data Desc
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON "+err.Error(), http.StatusBadRequest)
		return
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
