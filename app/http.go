package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.0x0001f346.de/andreas/ablage/config"
	"git.0x0001f346.de/andreas/ablage/filesystem"
	"github.com/julienschmidt/httprouter"
)

const httpPathRoot string = "/"
const httpPathConfig string = "/config/"
const httpPathFaviconICO string = "/favicon.ico"
const httpPathFaviconSVG string = "/favicon.svg"
const httpPathFiles string = "/files/"
const httpPathFilesDeleteFilename string = "/files/delete/:filename"
const httpPathFilesGetFilename string = "/files/get/:filename"
const httpPathScriptJS string = "/script.js"
const httpPathStyleCSS string = "/style.css"
const httpPathUpload string = "/upload/"

func httpGetConfig(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type Endpoints struct {
		Files       string `json:"Files"`
		FilesDelete string `json:"FilesDelete"`
		FilesGet    string `json:"FilesGet"`
		Upload      string `json:"Upload"`
	}

	type Modes struct {
		Readonly bool `json:"Readonly"`
		Sinkhole bool `json:"Sinkhole"`
	}

	type Config struct {
		Endpoints Endpoints `json:"Endpoints"`
		Modes     Modes     `json:"Modes"`
	}

	var config Config = Config{
		Endpoints: Endpoints{
			Files:       httpPathFiles,
			FilesDelete: httpPathFilesDeleteFilename,
			FilesGet:    httpPathFilesGetFilename,
			Upload:      httpPathUpload,
		},
		Modes: Modes{
			Readonly: config.GetReadonlyMode(),
			Sinkhole: config.GetSinkholeMode(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func httpGetFaviconICO(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.Redirect(w, r, "/favicon.svg", http.StatusSeeOther)
}

func httpGetFaviconSVG(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(assetFaviconSVG)
}

func httpGetFiles(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type FileInfo struct {
		Name string `json:"Name"`
		Size int64  `json:"Size"`
	}

	if config.GetSinkholeMode() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]FileInfo{})
	}

	entries, err := os.ReadDir(config.GetPathDataFolder())
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	files := make([]FileInfo, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name: info.Name(),
			Size: info.Size(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func httpGetFilesDeleteFilename(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if config.GetReadonlyMode() {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	if config.GetSinkholeMode() {
		http.Error(w, "404 File Not Found", http.StatusNotFound)
		return
	}

	entries, err := os.ReadDir(config.GetPathDataFolder())
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	files := map[string]int64{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files[info.Name()] = info.Size()
	}

	filename := ps.ByName("filename")
	sizeInBytes, fileExists := files[filename]
	if !fileExists {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	fullPath := filepath.Join(config.GetPathDataFolder(), filename)
	err = os.Remove(fullPath)
	if err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("| Delete   | %-21s | %-10s | %s\n", getClientIP(r), filesystem.GetHumanReadableSize(sizeInBytes), filename)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func httpGetFilesGetFilename(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if config.GetSinkholeMode() {
		http.Error(w, "404 File Not Found", http.StatusNotFound)
		return
	}

	filename := ps.ByName("filename")
	filePath := filepath.Join(config.GetPathDataFolder(), filename)

	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		http.Error(w, "404 File Not Found", http.StatusNotFound)
		return
	}

	log.Printf("| Download | %-21s | %-10s | %s\n", getClientIP(r), filesystem.GetHumanReadableSize(info.Size()), filename)

	extension := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(extension)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)

	if isBrowserDisplayableFileType(extension) {
		w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	}

	http.ServeFile(w, r, filePath)
}

func httpGetRoot(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(assetIndexHTML)
}

func httpGetScriptJS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "text/javascript")
	w.Write(assetScriptJS)
}

func httpGetStyleCSS(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Type", "text/css")
	w.Write(assetStyleCSS)
}

func httpPostUpload(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if config.GetReadonlyMode() {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not get multipart reader: %v", err), http.StatusBadRequest)
		return
	}

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading part: %v", err), http.StatusInternalServerError)
			return
		}
		defer part.Close()

		if part.FileName() == "" {
			continue
		}

		safeFilename := filesystem.SanitizeFilename(part.FileName())
		pathToFileInDataFolder := filepath.Join(config.GetPathDataFolder(), safeFilename)
		pathToFileInUploadFolder := filepath.Join(config.GetPathUploadFolder(), safeFilename)

		if _, err = os.Stat(pathToFileInDataFolder); err == nil {
			http.Error(w, "File already exists", http.StatusConflict)
			return
		}
		if _, err = os.Stat(pathToFileInUploadFolder); err == nil {
			http.Error(w, "File already exists", http.StatusConflict)
			return
		}

		uploadFile, err := os.Create(pathToFileInUploadFolder)
		if err != nil {
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		bytesWritten, err := io.Copy(uploadFile, part)
		uploadFile.Close()
		if err != nil {
			_ = os.Remove(pathToFileInUploadFolder)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err = os.Rename(pathToFileInUploadFolder, pathToFileInDataFolder); err != nil {
			_ = os.Remove(pathToFileInDataFolder)
			_ = os.Remove(pathToFileInUploadFolder)
			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
			return
		}

		log.Printf("| Upload   | %-21s | %-10s | %s\n",
			getClientIP(r), filesystem.GetHumanReadableSize(bytesWritten), safeFilename)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
