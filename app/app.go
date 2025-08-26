package app

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"git.0x0001f346.de/andreas/ablage/config"
	"github.com/julienschmidt/httprouter"
)

//go:embed assets/index.html
var assetIndexHTML []byte

//go:embed assets/favicon.svg
var assetFaviconSVG []byte

//go:embed assets/script.js
var assetScriptJS []byte

//go:embed assets/style.css
var assetStyleCSS []byte

func Init() {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	router.GET(httpPathRoot, httpGetRoot)
	router.GET(httpPathConfig, httpGetConfig)
	router.GET(httpPathFaviconICO, httpGetFaviconICO)
	router.GET(httpPathFaviconSVG, httpGetFaviconSVG)
	router.GET(httpPathFiles, httpGetFiles)
	router.GET(httpPathFilesDeleteFilename, httpGetFilesDeleteFilename)
	router.GET(httpPathFilesGetFilename, httpGetFilesGetFilename)
	router.GET(httpPathScriptJS, httpGetScriptJS)
	router.GET(httpPathStyleCSS, httpGetStyleCSS)
	router.POST(httpPathUpload, httpPostUpload)

	var handler http.Handler = router

	if config.GetBasicAuthMode() {
		handler = basicAuthMiddleware(handler, config.GetBasicAuthUsername(), config.GetBasicAuthPassword())
	}

	if config.GetHttpMode() {
		config.PrintStartupBanner()
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.GetPortToListenOn()), handler)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ablage exited with error:\n%v\n", err)
			os.Exit(1)
		}
		return
	}

	tlsCert, err := tls.X509KeyPair(config.GetTLSCertificate(), config.GetTLSKey())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Faild to parse PEM encoded public/private key pair:\n%v\n", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:     fmt.Sprintf(":%d", config.GetPortToListenOn()),
		ErrorLog: log.New(io.Discard, "", 0),
		Handler:  handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	config.PrintStartupBanner()

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ablage exited with error:\n%v\n", err)
		os.Exit(1)
	}
}

func getClientIP(r *http.Request) string {
	if r.Header.Get("X-Forwarded-For") != "" {
		return r.Header.Get("X-Forwarded-For")
	}

	return r.RemoteAddr
}

func isBrowserDisplayableFileType(extension string) bool {
	browserDisplayableFileTypes := map[string]struct{}{
		// audio
		".mp3": {}, ".ogg": {}, ".wav": {},
		// pictures
		".bmp": {}, ".gif": {}, ".ico": {}, ".jpg": {}, ".jpeg": {},
		".png": {}, ".svg": {}, ".webp": {},
		// programming
		".bat": {}, ".cmd": {}, ".c": {}, ".cpp": {}, ".go": {},
		".h": {}, ".hpp": {}, ".java": {}, ".kt": {}, ".lua": {},
		".php": {}, ".pl": {}, ".ps1": {}, ".py": {}, ".rb": {},
		".rs": {}, ".sh": {}, ".swift": {}, ".ts": {}, ".tsx": {},
		// text
		".csv": {}, ".log": {}, ".md": {}, ".pdf": {}, ".txt": {},
		// video
		".mp4": {}, ".webm": {},
		// web
		".css": {}, ".js": {}, ".html": {},
	}

	_, isBrowserDisplayableFileType := browserDisplayableFileTypes[extension]

	return isBrowserDisplayableFileType
}
