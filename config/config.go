package config

import (
	"fmt"
	"os"
)

const DefaultBasicAuthUsername string = "ablage"
const DefaultNameDataFolder string = "data"
const DefaultNameUploadFolder string = ".upload"
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
