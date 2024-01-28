package regen

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
)

type correctedBlockType struct {
	offset                         int64
	checksumBlock                  int
	blockLength                    int
	checksumBlockCombinationBuffer []byte
}

func Regenerate(filename string, bruteforceLimit int, verbose bool) error {
	if verbose {
		defer timer("Regenerate")()
	}
	archiveFile, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer archiveFile.Close()
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return err
	}
	regenFile, err := os.Open(filename + ".regen")
	if err != nil {
		return err
	}
	defer regenFile.Close()

	// read header info
	header := make([]byte, 11)
	regenFile.Read(header)
	if header[0] != 'R' ||
		header[1] != 'E' ||
		header[2] != 'G' ||
		header[3] != 'E' ||
		header[4] != 'N' {
		return errors.New("not a valid regen file")
	}
	version := binary.BigEndian.Uint16(header[5:7])
	checksumBlockLength := int(binary.BigEndian.Uint16(header[7:9]))
	parityBlocks := binary.BigEndian.Uint16(header[9:11])
	archiveFileSize := fileInfo.Size()
	parityBlockLength := int64(math.Floor(float64(archiveFileSize /
		int64(parityBlocks))))
	if verbose {
		fmt.Println("Version:", version)
		fmt.Println("Checksum Block Length:", checksumBlockLength)
		fmt.Println("Parity Blocks:", parityBlocks)
		fmt.Println("Archive File Size:", archiveFileSize)
		fmt.Println("Parity Block Length:", parityBlockLength)
	}
	if int(version) > VERSION {
		return fmt.Errorf("regen file version %d not supported. Update the program", version)
	}

	// iterate checksum blocks
	checksumBlocks := int(math.Ceil(float64(parityBlockLength) /
		float64(checksumBlockLength)))
	lastChecksumBlockLength := int(parityBlockLength -
		(int64(checksumBlocks-1) * int64(checksumBlockLength)))
	checksumBlockBuffer := make([]byte, checksumBlockLength)
	checksumBuffer := make([]byte, 2)
	var sum uint16
	parityBuffer := make([][]byte, parityBlocks)
	for i := range parityBuffer {
		parityBuffer[i] = make([]byte, checksumBlockLength)
	}
	parityOutBuffer := make([]byte, checksumBlockLength)
	regenParityBuffer := make([]byte, checksumBlockLength)
	correctedBlocks := []correctedBlockType{}
	failedBlocks := []int{}
	for i := 0; i < int(parityBlocks); i++ {
		for j := 0; j < checksumBlocks; j++ {
			offset := int64(i*int(parityBlockLength) + j*checksumBlockLength)
			archiveFile.ReadAt(checksumBlockBuffer, offset)
			lastChecksumBlock := j == checksumBlocks-1
			if lastChecksumBlock {
				// truncate and compute checksum for the final checksum block
				if lastChecksumBlockLength == 0 {
					continue
				}
				sum = fletcher16(checksumBlockBuffer[:lastChecksumBlockLength])
			} else {
				sum = fletcher16(checksumBlockBuffer)
			}
			regenFile.ReadAt(checksumBuffer, int64(11+(i*(checksumBlocks*2))+j*2))
			regenSum := binary.BigEndian.Uint16(checksumBuffer)
			if sum != regenSum {
				if verbose {
					fmt.Println("Checksum error found:", i, j*2)
				}

				// get parity data for checksum block
				regenParityOffset := 11 + (checksumBlocks * 2 * int(parityBlocks))
				regenFile.ReadAt(regenParityBuffer,
					int64(regenParityOffset+j*checksumBlockLength))
				for k := 0; k < int(parityBlocks); k++ {
					archiveFile.ReadAt(parityBuffer[k],
						int64(k*int(parityBlockLength)+j*checksumBlockLength))
				}
				parityOutBuffer = xorParity(&parityBuffer)

				// identify bad bits
				var blockLength int
				if lastChecksumBlock {
					blockLength = lastChecksumBlockLength
				} else {
					blockLength = checksumBlockLength
				}
				badBits := make([]int, 0)
				for k := 0; k < blockLength; k++ {
					parityByte := parityOutBuffer[k]
					regenByte := regenParityBuffer[k]
					for l := 0; l < 8; l++ {
						bitMask := byte(1 << l)
						parityBitByte := parityByte & bitMask
						regenBitByte := regenByte & bitMask
						if parityBitByte != regenBitByte {
							badBits = append(badBits, (k*8)+l)
						}
					}
				}
				if len(badBits) > 1 {
					if verbose {
						fmt.Println("Bad bits:", badBits)
					}
				}
				// if len(badBits) > 5 {
				// 	fmt.Println("HIGH BAD BITS:", len(badBits))
				// }

				// brute force combinations
				combinations := generateCombinations(badBits, bruteforceLimit)
				checksumBlockCombinationBuffer := make([]byte, checksumBlockLength)
				found := false
				for k, combination := range combinations {
					copy(checksumBlockCombinationBuffer, checksumBlockBuffer)
					for _, index := range combination {
						flipByte := int(math.Floor(float64(index / 8)))
						flipMask := byte(1 << (index % 8))
						checksumBlockCombinationBuffer[flipByte] =
							checksumBlockCombinationBuffer[flipByte] ^ flipMask
					}
					// calculate checksum
					sum = fletcher16(checksumBlockCombinationBuffer)
					if sum == binary.BigEndian.Uint16(checksumBuffer) {
						found = true
						if len(combinations) > 1 {
							if verbose {
								fmt.Println("Combination found:", k, combination, sum)
							}
						}
						break
					}
				}
				if found {
					if lastChecksumBlock {
						correctedBlock := correctedBlockType{
							offset,
							j,
							lastChecksumBlockLength,
							checksumBlockCombinationBuffer}
						correctedBlocks = append(correctedBlocks, correctedBlock)
					} else {
						correctedBlock := correctedBlockType{
							offset,
							j,
							checksumBlockLength,
							checksumBlockCombinationBuffer}
						correctedBlocks = append(correctedBlocks, correctedBlock)
					}
				} else {
					failedBlocks = append(failedBlocks, j)
					if verbose {
						fmt.Println("Could not find the correct combination")
					}
				}
			}
		}
	}

	// write corrected blocks
	for _, block := range correctedBlocks {
		failedBlock := false
		for i := range failedBlocks {
			if i == block.checksumBlock {
				failedBlock = true
				break
			}
		}
		if !failedBlock {
			if verbose {
				fmt.Println("Writing corrected block", block.checksumBlock,
					block.offset)
			}
			buffer := make([]byte, block.blockLength)
			copy(buffer, block.checksumBlockCombinationBuffer)
			_, err := archiveFile.WriteAt(buffer, block.offset)
			if err != nil {
				return err
			}
		} else {
			if verbose {
				fmt.Println("Not writing corrected block", block.checksumBlock,
					block.offset)
			}
		}
	}
	if len(failedBlocks) != 0 {
		return fmt.Errorf("failed blocks: %v", len(failedBlocks))
	}
	return nil
}
