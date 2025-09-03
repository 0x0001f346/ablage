package config

import (
	"fmt"
	"os"
	"path/filepath"
)

var defaultPathDataFolder string = ""
var defaultPathUploadFolder string = ""

func gatherDefaultPaths() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Could not determine binary path: %v", err)
	}

	defaultPathDataFolder = filepath.Join(filepath.Dir(execPath), DefaultNameDataFolder)
	defaultPathUploadFolder = filepath.Join(defaultPathDataFolder, DefaultNameUploadFolder)

	return nil
}
