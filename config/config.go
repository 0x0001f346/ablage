package config

import (
	"fmt"
)

const DefaultBasicAuthUsername string = "ablage"
const DefaultNameDataFolder string = "data"
const DefaultNameUploadFolder string = ".upload"
const DefaultPortToListenOn int = 13692
const LengthOfRandomBasicAuthPassword int = 16
const VersionString string = "1.1"

var randomBasicAuthPassword string = generateRandomPassword()

func Init() error {
	err := gatherDefaultPaths()
	if err != nil {
		return err
	}

	parseFlags()

	if GetReadonlyMode() && GetSinkholeMode() {
		return fmt.Errorf("Cannot enable both readonly and sinkhole modes at the same time.")
	}

	err = loadOrGenerateTLSCertificate()
	if err != nil {
		return err
	}

	return nil
}
