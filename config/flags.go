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

func parseFlags() error {
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

	err := parseFlagValuePortToListenOn()
	if err != nil {
		return err
	}

	parseFlagValuePathDataFolder()

	err = parseFlagValuePathTLSCertFile()
	if err != nil {
		return err
	}

	err = parseFlagValuePathTLSKeyFile()
	if err != nil {
		return err
	}

	return nil
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

	pathUploadFolder = filepath.Join(pathDataFolder, DefaultNameUploadFolder)
}

func parseFlagValuePortToListenOn() error {
	if portToListenOn < 1 || portToListenOn > 65535 {
		return fmt.Errorf("The port must be between 1 and 65535 (both ports included).")
	}

	return nil
}

func parseFlagValuePathTLSCertFile() error {
	if pathTLSCertFile == "" {
		if pathTLSKeyFile != "" {
			return fmt.Errorf("Both a certificate and the corresponding key must be provided.")
		}

		return nil
	}

	info, err := os.Stat(pathTLSCertFile)
	if err != nil {
		return fmt.Errorf("Failed to read cert: %v", err)
	}

	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Cert must be a file\n")
		return fmt.Errorf("Cert must be a valid file.")
	}

	return nil
}

func parseFlagValuePathTLSKeyFile() error {
	if pathTLSKeyFile == "" {
		if pathTLSCertFile != "" {
			return fmt.Errorf("Both a certificate and the corresponding key must be provided.")
		}

		return nil
	}

	info, err := os.Stat(pathTLSKeyFile)
	if err != nil {
		return fmt.Errorf("Failed to read key: %v", err)
	}

	if info.IsDir() {
		return fmt.Errorf("Key must be a valid file.")
	}

	return nil
}
