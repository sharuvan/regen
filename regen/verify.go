package regen

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Verifies archive's data integrity using saved SHA256 hash
func Verify(filename string, verbose bool) error {
	if verbose {
		defer timer("Verify")()
	}
	if filename == "" {
		return errors.New("archive file name not specified")
	}
	fileContent, err := os.ReadFile(filename + ".sha256")
	if err != nil {
		return errors.New("error reading .sha256 hash file")
	}
	hash := string(fileContent)
	if len(hash) < 64 {
		return errors.New("invalid hash file")
	}
	hash = hash[:64]
	archiveFile, err := os.Open(filename)
	if err != nil {
		return errors.New("error reading archive file")
	}
	defer archiveFile.Close()
	archiveHash, err := calculateSHA256(archiveFile)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Println("Saved    SHA256 hash:", hash)
		fmt.Println("Computed SHA256 hash:", archiveHash)
	}
	if strings.Compare(hash, archiveHash) == 0 {
		if verbose {
			fmt.Println("no errors in data integrity")
		}
	} else {
		return errors.New("data is corrupt. Use regenerate")
	}
	return nil
}
