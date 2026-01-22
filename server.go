package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/mux"
)

const maxUploadSize = 1 * 1024 * 1024 // 1 mb

type Desc struct {
	Description string `json:"description"`
	Upload_date string
	Is_utf8     bool
}

type Desc2 struct {
	Description string    `json:"description"`
	UploadDate  time.Time `json:"upload_date"`
	IsUTF8      bool      `json:"is_utf8"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/upload", HandleFileUpload).Methods("POST")
	router.HandleFunc("/status", getDirectoryData).Methods("GET")
	router.HandleFunc("/search", Search).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func HandleFileUpload(w http.ResponseWriter, r *http.Request) {

	var data Desc

	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		return
	}

	uploadFile, handler, err := r.FormFile("uploadFile")
	buffer := make([]byte, 1024)
	_, err = uploadFile.Read(buffer)
	if utf8.Valid(buffer) {

		data.Is_utf8 = true

	}
	descriptionJson := r.FormValue("descriptionData")
	json.Unmarshal([]byte(descriptionJson), &data)
	data.Upload_date = time.Now().String()

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

	if _, err := os.Stat(filepath.Join("uploads", fileName_no_extension)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Join("uploads", fileName_no_extension), 0755)

		if err != nil {
			http.Error(w, "Error creating folder: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	finalPath := filepath.Join(uploadPath, fileName_no_extension+"/"+handler.Filename)

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
	dat, err := json.Marshal(data)
	_, err = os.Create(filepath.Join(uploadPath, fileName_no_extension+"/"+"."+fileName_no_extension+".json"))
	os.WriteFile(filepath.Join(uploadPath, fileName_no_extension+"/"+"."+fileName_no_extension+".json"), dat, 0644)

	w.Write([]byte("SUCCESS"))

}

func Search(w http.ResponseWriter, r *http.Request) {

	searchQuery := r.URL.Query().Get("sq")
	fmt.Println(searchQuery)

	var pathList []string
	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"
	_ = filepath.Walk(uploadPath, func(path1 string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.Contains(info.Name(), ".json") {
				pathList = append(pathList, path1)
				// TO DO - Check if the actual file in the folder (not the json) is UTF-8. If so also add it to the pathList array
			}
		}
		return nil
	})
	var files []string

	for i := 0; i < len(pathList); i++ {
		var data Desc2
		jsonFile, _ := os.Open(pathList[i])
		defer jsonFile.Close()
		fileBytes, _ := io.ReadAll(jsonFile)
		json.Unmarshal([]byte(fileBytes), &data)
		if strings.Contains(data.Description, string(searchQuery)) {
			files = append(files, pathList[i])
			// TO-DO - Fix the append above so it just returns the folder not the actual file
		}
	}
	// TO-DO - Once the append above has been fixed remove any duplicate folders and return it back to the requester
	fmt.Println(files)
}

func getDirectoryData(w http.ResponseWriter, r *http.Request) {

	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"
	root := os.DirFS(uploadPath)
	allFiles, _ := fs.Glob(root, "*.*")

	var files []string
	for _, f := range allFiles {
		files = append(files, path.Join(uploadPath, f))
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
