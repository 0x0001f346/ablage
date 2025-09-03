package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git.0x0001f346.de/andreas/ablage/config"
)

func Init() error {
	err := prepareDataFolder()
	if err != nil {
		return err
	}

	err = prepareUploadDir()
	if err != nil {
		return err
	}

	return nil
}

func GetHumanReadableSize(bytes int64) string {
	const unit int64 = 1024

	if bytes < unit {
		return fmt.Sprintf("%d Bytes", bytes)
	}

	div, exp := int64(unit), 0

	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func SanitizeFilename(dirtyFilename string) string {
	if dirtyFilename == "" {
		return "upload.bin"
	}

	filenameWithoutPath := filepath.Base(dirtyFilename)

	extension := filepath.Ext(filenameWithoutPath)
	filenameWithoutPathAndExtension := filenameWithoutPath[:len(filenameWithoutPath)-len(extension)]

	cleanedFilename := strings.ReplaceAll(filenameWithoutPathAndExtension, " ", "_")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "Ä", "Ae")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "ä", "äe")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "Ö", "Oe")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "ö", "oe")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "Ü", "Ue")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "ü", "ue")
	cleanedFilename = strings.ReplaceAll(cleanedFilename, "ß", "ss")

	var safeNameRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	cleanedFilename = safeNameRegex.ReplaceAllString(cleanedFilename, "_")

	for strings.Contains(cleanedFilename, "__") {
		cleanedFilename = strings.ReplaceAll(cleanedFilename, "__", "_")
	}

	cleanedFilename = strings.Trim(cleanedFilename, "_")

	const maxLenFilename int = 128
	if len(cleanedFilename) > maxLenFilename {
		cleanedFilename = cleanedFilename[:maxLenFilename]
	}

	return cleanedFilename + extension
}

func prepareDataFolder() error {
	info, err := os.Stat(config.GetPathDataFolder())
	if os.IsNotExist(err) {
		if err := os.Mkdir(config.GetPathDataFolder(), 0755); err != nil {
			return fmt.Errorf("Could not create folder '%s': %v", config.GetPathDataFolder(), err)
		}
	} else if err != nil {
		return fmt.Errorf("Could not access '%s': %v", config.GetPathDataFolder(), err)
	} else if !info.IsDir() {
		return fmt.Errorf("'%s' exists but is not a directory", config.GetPathDataFolder())
	}

	pathTestFile := filepath.Join(config.GetPathDataFolder(), ".write_test")
	err = os.WriteFile(pathTestFile, []byte("test"), 0644)
	if err != nil {
		return fmt.Errorf("Could not create test file '%s': %v", pathTestFile, err)
	}

	err = os.Remove(pathTestFile)
	if err != nil {
		return fmt.Errorf("Could not delete test file '%s': %v", pathTestFile, err)
	}

	return nil
}

func prepareUploadDir() error {
	info, err := os.Stat(config.GetPathUploadFolder())
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists, but is not a folder", config.GetPathUploadFolder())
		}

		err = os.RemoveAll(config.GetPathUploadFolder())
		if err != nil {
			return fmt.Errorf("Could not delete upload folder '%s': %v", config.GetPathUploadFolder(), err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("'%s' exists but is somewhat broken", config.GetPathUploadFolder())
	}

	if err := os.MkdirAll(config.GetPathUploadFolder(), 0755); err != nil {
		return fmt.Errorf("Could not create upload folder '%s': %v", config.GetPathUploadFolder(), err)
	}

	pathTestFile := filepath.Join(config.GetPathUploadFolder(), ".write_test")
	err = os.WriteFile(pathTestFile, []byte("test"), 0644)
	if err != nil {
		return fmt.Errorf("Could not create test file '%s': %v", pathTestFile, err)
	}

	err = os.Remove(pathTestFile)
	if err != nil {
		return fmt.Errorf("Could not delete test file '%s': %v", pathTestFile, err)
	}

	return nil
}
