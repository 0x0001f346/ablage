package config

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var basicAuthMode bool = false
var basicAuthPassword string = ""
var httpMode bool = false
var pathDataFolder string = ""
var pathTLSCertFile string = ""
var pathTLSKeyFile string = ""
var pathUploadFolder string = ""
var portToListenOn int = DefaultPortToListenOn
var readonlyMode bool = false
var sinkholeMode bool = false

func GetBasicAuthMode() bool {
	return basicAuthMode
}

func GetBasicAuthPassword() string {
	return basicAuthPassword
}

func GetBasicAuthUsername() string {
	return DefaultBasicAuthUsername
}

func GetHttpMode() bool {
	return httpMode
}

func GetPathDataFolder() string {
	return pathDataFolder
}

func GetPathTLSCertFile() string {
	return pathTLSCertFile
}

func GetPathTLSKeyFile() string {
	return pathTLSKeyFile
}

func GetPathUploadFolder() string {
	return pathUploadFolder
}

func GetPortToListenOn() int {
	return portToListenOn
}

func GetReadonlyMode() bool {
	return readonlyMode
}

func GetSinkholeMode() bool {
	return sinkholeMode
}

func generateRandomPassword() string {
	b := make([]byte, LengthOfRandomBasicAuthPassword)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)[:LengthOfRandomBasicAuthPassword]
}

func parseFlags() {
	flag.BoolVar(&basicAuthMode, "auth", false, "Enable basic authentication.")
	flag.BoolVar(&httpMode, "http", false, "Enable http mode. Nothing will be encrypted.")
	flag.BoolVar(&readonlyMode, "readonly", false, "Enable readonly mode. No files can be uploaded or deleted.")
	flag.BoolVar(&sinkholeMode, "sinkhole", false, "Enable sinkhole mode. Existing files won't be visible.")
	flag.IntVar(&portToListenOn, "port", DefaultPortToListenOn, "Set Port to listen on.")
	flag.StringVar(&basicAuthPassword, "password", "", "Set password for basic authentication (or let ablage generate a random one).")
	flag.StringVar(&pathDataFolder, "path", "", "Set path to data folder (default is 'data' in the same directory as ablage).")
	flag.StringVar(&pathTLSCertFile, "cert", "", "TLS cert file")
	flag.StringVar(&pathTLSKeyFile, "key", "", "TLS key file")

	flag.Parse()

	parseFlagValueBasicAuthPassword()
	parseFlagValuePortToListenOn()
	parseFlagValuePathDataFolder()
	parseFlagValuePathTLSCertFile()
	parseFlagValuePathTLSKeyFile()
}

func parseFlagValueBasicAuthPassword() {
	if len(basicAuthPassword) < 1 || len(basicAuthPassword) > 128 {
		basicAuthPassword = generateRandomPassword()
	}
}

func parseFlagValuePathDataFolder() {
	if pathDataFolder == "" {
		pathDataFolder = defaultPathDataFolder
		pathUploadFolder = defaultPathUploadFolder
		return
	}

	info, err := os.Stat(pathDataFolder)
	if err != nil {
		pathDataFolder = defaultPathDataFolder
		pathUploadFolder = defaultPathUploadFolder
		return
	}

	if !info.IsDir() {
		pathDataFolder = defaultPathDataFolder
		pathUploadFolder = defaultPathUploadFolder
		return
	}

	pathUploadFolder = filepath.Join(pathDataFolder, DefaultNameUploadFolder)
}

func parseFlagValuePortToListenOn() {
	if portToListenOn < 1 || portToListenOn > 65535 {
		portToListenOn = DefaultPortToListenOn
	}
}

func parseFlagValuePathTLSCertFile() {
	if pathTLSCertFile == "" {
		pathTLSKeyFile = ""
		return
	}

	info, err := os.Stat(pathTLSCertFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read cert: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Cert must be a file\n")
		os.Exit(1)
	}
}

func parseFlagValuePathTLSKeyFile() {
	if pathTLSKeyFile == "" {
		pathTLSCertFile = ""
		return
	}

	info, err := os.Stat(pathTLSKeyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to read key: %v\n", err)
		os.Exit(1)
	}

	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Key must be a file\n")
		os.Exit(1)
	}
}
