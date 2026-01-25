# [File-Processing-Pipeline]


A file processing and storage backend  with metadata, written in Go. 

## Background

- Sometimes when you have an important file, it takes time to search for it in the future
    - There is no additional metadata to search for a document, other than knowing certain keywords in the file

- This is where the File Processing process can be used to help
    - The backend system provides an easy to upload a file, but also provide a description of the file, which is stored in the file

- The default location of the files is the 'uploads' folder inside of the location that you run the script (see example below)

    - Home
        - Tom
            - Documents
                - File-Processing-Pipeline
                    - Uploads

- A config file can be used to specify the location of the uploads folder (see Configuration section)

## Operations 

- The backend system exposes four endpoints

    - /status - (GET) - Shows the current storage usage of your ‘uploads’ folder
    - /upload - (POST) Allows users to upload a file of up to a certain size (The maximum filesize of files can be set in the configuration (see Configuration section)
        - Alongside the file you can provide a description as a json with the key of 'description'. Foe example {"description": "My current CV 2026"}

    - /search - (GET) - Allows users to search for words in the description provided with the file. You can also search inside the files themselves if they are encoded using UTF-8. Example files are:
        - An example search of 'Apples' is http://localhost:8080/search?sq=Apples
            - The 'sq' field provides the search words to use
            - Note that searches are not case sensitive. Searching for 'apples' or 'Apples' will yield the same results if either the description or file data contains either 'apples' or 'Apples'
        - Example files that use UTF-8 encoding and therefore its file contents can also be searched) are:
            - Web Files: .html, .css, .js, .xml
            - Text & Data Files: .txt, .csv, .json, .md (Markdown)
            - Configuration & Scripts: .py (Python), .ini, .yaml
        - The search results provided will provide you with a direct link from the /download endpoint to download the files in the search results

    - /download - (GET) - Allows users to download a file. The endpoint uses the name of the folder where your file is located
    - If for example you have a file called accounts-2026.xlsx and your files stored in /home/johnsmith/Documents/files and there you upload a folder created called accounts-2026
        - If your server is located on the standard port 8080 then by making a GET request to the endpoint http://localhost:8080/download?file=accounts-2026 you will be returned the file

## Usage

- This script has currently only been tested on Linux, but it will also be tested on Windows & Mac, and once completed this section will be updated to confirm full compatibility

### Running program

- The program can be run using the command 'go run server.go'
- The backend server will then be available on port 8000. You can also specify a port using the 'port' option in the configuration file (see Configuration section)

### Configuration
- The config file 'config.yml' should be located in the root folder of File-Processing-Pipeline. 
- An example config file can be used by renaming it

#### Save Location / Upload path
- The 'uploadlocation' field in the config file provides the user the option to set a location of where they would like their files to be stored

#### Maximum filesize
- The 'maxfilesize' field in the config file provides the user the option to set the maximum filesize that can be accepted by the backend. The default maximum size is 5MB
    - Please note that this value needs to be in megabytes (MB). The formula to convert MB to bytes is :- megabytes=bytes/1024/1024 

## License 

- This program is provided under the GNU General Public License v3.0
