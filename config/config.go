package config

import (
	"fmt"
	"os"
)

const DefaultBasicAuthUsername string = "ablage"
const DefaultNameDataFolder string = "data"
const DefaultNameUploadFolder string = "upload"
const DefaultPortToListenOn int = 13692
const LengthOfRandomBasicAuthPassword int = 16
const VersionString string = "1.0"

var randomBasicAuthPassword string = generateRandomPassword()

func Init() {
	err := gatherDefaultPaths()
	if err != nil {
		panic(err)
	}

	parseFlags()

	if GetReadonlyMode() && GetSinkholeMode() {
		fmt.Println("Error: Cannot enable both readonly and sinkhole modes at the same time.")
		os.Exit(1)
	}

	err = loadOrGenerateTLSCertificate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func PrintStartupBanner() {
	fmt.Println("****************************************")
	fmt.Println("*                Ablage                *")
	fmt.Println("****************************************")
	fmt.Printf("Version        : %s\n", VersionString)
	fmt.Printf("Basic Auth mode: %v\n", GetBasicAuthMode())
	fmt.Printf("HTTP mode      : %v\n", GetHttpMode())
	fmt.Printf("Readonly mode  : %v\n", GetReadonlyMode())
	fmt.Printf("Sinkhole mode  : %v\n", GetSinkholeMode())
	fmt.Printf("Path           : %s\n", GetPathDataFolder())

	if GetBasicAuthMode() {
		fmt.Printf("Username       : %s\n", GetBasicAuthUsername())
		fmt.Printf("Password       : %s\n", GetBasicAuthPassword())
	}

	if GetHttpMode() {
		fmt.Printf("Listening on   : http://0.0.0.0:%d\n", GetPortToListenOn())
	} else {
		if pathTLSCertFile == "" || pathTLSKeyFile == "" {
			fmt.Printf("TLS cert       : self-signed\n")
			fmt.Printf("TLS key        : self-signed\n")
		} else {
			fmt.Printf("TLS cert       : %s\n", pathTLSCertFile)
			fmt.Printf("TLS key        : %s\n", pathTLSKeyFile)
		}

		fmt.Printf("Listening on   : https://0.0.0.0:%d\n", GetPortToListenOn())
	}

	fmt.Println("")
}
