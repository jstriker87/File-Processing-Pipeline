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

- A future parameter (-upload-path, -up) will allow users to specify the location of the uploads folder (the folder name can also be changed as desired)

## Operations 

- The backend system exposes four endpoints

    - /status - Shows a basic list of the files in the ‘uploads’ folder within the backend (you can also change the folder location)
    - /upload - Allows users to upload a file of up to a certain size ( in the future users will be  able set the maximum file size using the parameters  '--max-filesize,-mf')
        - Alongside the file you can provide a description as a json with the key of 'description'. Foe example {"description": "My current CV 2026"}

    - /search - Allows users to search for words in the description provided with the file. You can also search inside the files themselves if they are encoded using UTF-8. Example files are:
        - An example search of 'Apples' is http://localhost:8080/search?sq=Apples
            - The 'sq' field provides the search words to use
            - Note that searches are not case sensitive. Searching for 'apples' or 'Apples' will yield the same results if either the description or file data contains either 'apples' or 'Apples'
        - Example files that use UTF-8 encoding are:
            - Web Files: .html, .css, .js, .xml
            - Text & Data Files: .txt, .csv, .json, .md (Markdown)
            - Configuration & Scripts: .py (Python), .ini, .yaml
        - The search results provided will provide you with a direct link from the /download endpoint to download the files in the search results

    - /download - Allows users to download a file. TBC
## Usage

- This script has currently only been tested on Linux, but it will also be tested on Windows & Mac, and once completed this section will be updated to confirm full compatibility

### Save Location / Upload path

-  (TBC) - The program will provide a parameter when running (--upload-path,-up), to allow the user to specify where the upload folder is located
    - There will also be an option to use a basic yaml config file to specify the upload path


### Running program

TBC


## License 

TBC


## Dependencies / Acknowledgements

TBC
