package regen

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Generates a regen file and SHA256 hash file out of given archive file
func Generate(filename string, percentage int, checksumBlockLength int,
	verbose bool) error {
	if verbose {
		defer timer("Generate")()
	}
	if filename == "" {
		return errors.New("archive file name not specified")
	}

	archiveFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer archiveFile.Close()
	hash, err := calculateSHA256(archiveFile)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Println("Saving SHA256 hash:", hash)
	}
	hashFile, err := os.Create(filename + ".sha256")
	if err != nil {
		return err
	}
	defer hashFile.Close()
	fileNameBase := filepath.Base(archiveFile.Name())
	hashFile.WriteString(hash + "  " + fileNameBase)

	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	parityBlocks := int64(math.Round(float64(100 / percentage)))
	parityBlockLength := int64(math.Floor(float64(fileSize / parityBlocks)))
	if verbose {
		fmt.Println("File Size:", fileSize)
		fmt.Println("Parity Blocks:", parityBlocks)
		fmt.Println("Parity Block Length:", parityBlockLength)
		fmt.Println("Checksum Block Length:", checksumBlockLength)
		fmt.Println("Generating regen file")
	}
	regenFile, err := os.Create(filename + ".regen")
	if err != nil {
		return err
	}
	defer regenFile.Close()

	header := make([]byte, 11)
	header[0] = 'R'
	header[1] = 'E'
	header[2] = 'G'
	header[3] = 'E'
	header[4] = 'N'
	binary.BigEndian.PutUint16(header[5:], uint16(1)) // version
	binary.BigEndian.PutUint16(header[7:], uint16(checksumBlockLength))
	binary.BigEndian.PutUint16(header[9:], uint16(parityBlocks))
	regenFile.Write(header)

	// write checksum data
	if verbose {
		fmt.Println("Writing checksum data")
	}
	checksumBlocks := int(math.Ceil(float64(parityBlockLength) /
		float64(checksumBlockLength)))
	lastChecksumBlockLength := int(parityBlockLength -
		(int64(checksumBlocks-1) * int64(checksumBlockLength)))
	checksumBlockBuffer := make([]byte, checksumBlockLength)
	var sum uint16
	for i := 0; i < int(parityBlocks); i++ {
		for j := 0; j < checksumBlocks; j++ {
			offset := int64(i*int(parityBlockLength) + j*checksumBlockLength)
			archiveFile.ReadAt(checksumBlockBuffer, offset)
			if j == checksumBlocks-1 {
				// truncate and compute checksum for the final checksum block
				if lastChecksumBlockLength == 0 {
					continue
				}
				sum = fletcher16(checksumBlockBuffer[:lastChecksumBlockLength])
			} else {
				sum = fletcher16(checksumBlockBuffer)
			}
			sumBytes := make([]byte, 2)
			binary.BigEndian.PutUint16(sumBytes, sum)
			regenFile.Write(sumBytes)
		}
	}

	// write parity data
	if verbose {
		fmt.Println("Writing parity data")
	}
	// initialize buffer for data blocks
	parityBuffer := make([][]byte, parityBlocks)
	for i := range parityBuffer {
		parityBuffer[i] = make([]byte, BUFFER_SIZE)
	}
	parityOutBuffer := make([]byte, BUFFER_SIZE)
	bufferChunks := int(math.Ceil(float64(parityBlockLength) / float64(BUFFER_SIZE)))
	lastChunkSize := parityBlockLength - (int64(BUFFER_SIZE) * int64(bufferChunks-1))
	for i := 0; i < bufferChunks; i++ {
		for j := 0; j < int(parityBlocks); j++ {
			offset := (j * int(parityBlockLength)) + (i * BUFFER_SIZE)
			archiveFile.ReadAt(parityBuffer[j], int64(offset))
		}
		parityOutBuffer = xorParity(&parityBuffer)
		if i == bufferChunks-1 {
			regenFile.Write(parityOutBuffer[:lastChunkSize])
		} else {
			regenFile.Write(parityOutBuffer)
		}
	}
	return nil
}
