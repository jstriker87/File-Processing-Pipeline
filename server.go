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
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"
)

var maxUploadSize int

type Desc struct {
	Description string `json:"description"`
	UploadDate  time.Time
	FilePath    string
}

type Desc2 struct {
	Description  string    `json:"description"`
	UploadDate   time.Time `json:"uploaddate"`
	FilePath     string    `json:"filepath"`
	DownloadPath string
}

type Config struct {
	UploadLocation string `yaml:"uploadlocation"`
	MaxFileSize    int    `yaml:"maxfilesize"`
}

var config Config

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/upload", HandleFileUpload).Methods("POST")
	router.HandleFunc("/status", getDirectoryData).Methods("GET")
	router.HandleFunc("/search", Search).Methods("GET")
	router.HandleFunc("/download", DownloadFile).Methods("GET")

	f, err := os.ReadFile("config.yml")

	if err != nil {
		fmt.Println(err.Error())
		//log.Fatal(err)
	}

	if err := yaml.Unmarshal(f, &config); err != nil {
		fmt.Println(err.Error())
	}

	if config.MaxFileSize > 0 {
		maxUploadSize = int(config.MaxFileSize)
	} else {
		maxUploadSize = 1 * 1024 * 1024 // 1 mb
	}
	log.Fatal(http.ListenAndServe(":8080", router))

}

func HandleFileUpload(w http.ResponseWriter, r *http.Request) {

	var data Desc

	wd, _ := os.Getwd()

	var uploadPath string
	if len(config.UploadLocation) > 0 {

		uploadPath = config.UploadLocation

	} else {

		uploadPath = wd + "/uploads/"
	}

	if err := r.ParseMultipartForm(int64(maxUploadSize)); err != nil {
		fmt.Printf("Could not parse multipart form: %v\n", err)
		return
	}

	uploadFile, handler, err := r.FormFile("uploadFile")
	descriptionJson := r.FormValue("descriptionData")
	json.Unmarshal([]byte(descriptionJson), &data)
	data.UploadDate = time.Now()

	if err != nil {
		http.Error(w, "Error with provided file: "+err.Error(), http.StatusBadRequest)
		return
	}

	defer uploadFile.Close()

	if int(handler.Size) > maxUploadSize {
		folderSize := fmt.Sprintf("%.2f", float64(handler.Size)/(1024*1024))
		maxUploadSize_text := fmt.Sprintf("%.2f", float64(maxUploadSize)/(1024*1024))
		http.Error(w, "File size of "+folderSize+" MB is bigger than the allowed size of "+maxUploadSize_text+" MB", http.StatusBadRequest)
		return
	}

	fileBytes, err := io.ReadAll(uploadFile)

	if err != nil {
		http.Error(w, "Error with provided file: "+err.Error(), http.StatusBadRequest)
		return
	}

	fileName_no_extension := strings.Split(handler.Filename, ".")[0]
	fileName_no_extension = regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(fileName_no_extension, "")
	if len(data.Description) == 0 {

		data.Description = handler.Filename

	}

	if _, err := os.Stat(filepath.Join("uploads", fileName_no_extension)); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Join("uploads", fileName_no_extension), 0755)

		if err != nil {
			http.Error(w, "Error creating folder: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	finalPath := filepath.Join(uploadPath, fileName_no_extension+"/"+handler.Filename)
	data.FilePath = finalPath

	_, err = os.Stat(finalPath)
	if os.IsNotExist(err) {
		os.MkdirAll(filepath.Join(uploadPath, fileName_no_extension), 0700)

	} else {
		http.Error(w, "Error: file Exists", http.StatusBadRequest)
		return
	}
	newFile, err := os.Create(finalPath)

	if err != nil {
		http.Error(w, "Error: Unable to create file: "+finalPath+" "+err.Error(), http.StatusBadRequest)
		return
	}
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

func DownloadFile(w http.ResponseWriter, r *http.Request) {

	DownloadPath := r.URL.Query().Get("file")
	wd, _ := os.Getwd()
	Pth := wd + "/uploads/" + DownloadPath
	if _, err := os.Stat(Pth); os.IsNotExist(err) {
		return
	}

	_ = filepath.Walk(Pth, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			http.ServeFile(w, r, path)
		}

		return nil
	})

}

func Search(w http.ResponseWriter, r *http.Request) {

	searchQuery := strings.ToLower(r.URL.Query().Get("sq"))

	var pathList []string
	wd, _ := os.Getwd()
	uploadPath := wd + "/uploads/"
	_ = filepath.Walk(uploadPath, func(path1 string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			pathList = append(pathList, path1)
		}
		return nil
	})

	results := []Desc2{}
	for i := 0; i < len(pathList); i++ {
		var data Desc2
		jsonFile, _ := os.Open(pathList[i])
		fileBytes, _ := io.ReadAll(jsonFile)
		jsonFile.Close()
		if strings.Contains(pathList[i], ".json") {
			err := json.Unmarshal([]byte(fileBytes), &data)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if strings.Contains(strings.ToLower(data.Description), string(searchQuery)) {
				containsItem := slices.Contains(results, data)
				if containsItem == false {
					path2 := strings.Split(pathList[i], "/")
					data.DownloadPath = "http://localhost:8080/download?file=" + path2[len(path2)-2]
					data.FilePath = ""
					results = append(results, data)
				}
			}
		} else {

			fileString := strings.ToLower(string(fileBytes))
			if strings.Contains(fileString, string(searchQuery)) {
				_ = filepath.Walk(path.Dir(pathList[i]), func(path1 string, info os.FileInfo, err error) error {
					if strings.Contains(path1, ".json") {
						jsonFile, err := os.Open(path1)
						if err != nil {
							fmt.Println(err)
						}
						defer jsonFile.Close()
						fileBytes, _ := io.ReadAll(jsonFile)
						_ = json.Unmarshal([]byte(fileBytes), &data)
						containsItem := slices.Contains(results, data)
						if containsItem == false {
							data.FilePath = ""
							path2 := strings.Split(path1, "/")
							data.DownloadPath = "http://localhost:8080/download?file=" + path2[len(path2)-2]
							results = append(results, data)
						}

					}
					return nil
				})
			}

		}
	}

	fmt.Printf("results: %+v\n", results)
	_ = json.NewEncoder(w).Encode(results)
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
