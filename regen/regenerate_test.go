package regen

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
)

const FILE = "../testdata/cats.zip"
const N = 100
const ERRORS = 1000
const BURSTS = 1
const PARITY_PERCENTAGE = 10
const CHECKSUM_BLOCK_LENGTH = 64
const VERBOSE = false

func TestRegenerate(t *testing.T) {
	insertRandomBitErrors(FILE, 10)
	err := Regenerate(FILE, 1023, VERBOSE)
	if err != nil {
		t.Error(err)
	}
}

func BenchmarkRandomBurst(b *testing.B) {
	fails := 0
	for i := 0; i < N; i++ {
		fmt.Println("running test", i+1)
		Generate(FILE, PARITY_PERCENTAGE, CHECKSUM_BLOCK_LENGTH, VERBOSE)
		insertRandomBurstErrors(FILE, ERRORS, BURSTS)
		err := Regenerate(FILE, 1023, VERBOSE)
		if err != nil {
			fails += 1
		} else {
			err = Verify(FILE, VERBOSE)
			if err != nil {
				fails += 1
			}
		}
	}
	fmt.Printf("%d out of %d burst error benchmark tests failed\n", fails, N)
}

func insertRandomBurstErrors(filename string, n int, bursts int) {
	archiveFile, _ := os.OpenFile(filename, os.O_RDWR, 0666)
	defer archiveFile.Close()
	fileInfo, _ := os.Stat(filename)
	bitLength := fileInfo.Size() * 8
	errorBits := generateUniqueRandomNumbers(bursts, int(bitLength))
	burstLengths := distributeNumbers(n, bursts)
	for i, index := range errorBits {
		for j := 0; j < burstLengths[i]; j++ {
			flipByte := int(float64((index + j) / 8))
			flipMask := byte(1 << ((index + j) % 8))
			buffer := make([]byte, 1)
			archiveFile.ReadAt(buffer, int64(flipByte))
			buffer[0] = buffer[0] ^ flipMask
			archiveFile.WriteAt(buffer, int64(flipByte))
		}
	}
}

func distributeNumbers(n, m int) map[int]int {
	result := make(map[int]int)
	for i := 0; i < n; i++ {
		groupIndex := rand.Intn(m)
		_, exists := result[groupIndex]
		if !exists {
			result[groupIndex] = 1
		} else {
			result[groupIndex] = result[groupIndex] + 1
		}
	}
	return result
}

func BenchmarkRandomBit(b *testing.B) {
	fails := 0
	for i := 0; i < N; i++ {
		fmt.Println("running test", i+1)
		Generate(FILE, PARITY_PERCENTAGE, CHECKSUM_BLOCK_LENGTH, VERBOSE)
		insertRandomBitErrors(FILE, ERRORS)
		err := Regenerate(FILE, 1023, VERBOSE)
		if err != nil {
			fails += 1
		} else {
			err = Verify(FILE, VERBOSE)
			if err != nil {
				fails += 1
			}
		}
	}
	fmt.Printf("%d out of %d bit error benchmark tests failed\n", fails, N)
}

func insertRandomBitErrors(filename string, n int) {
	archiveFile, _ := os.OpenFile(filename, os.O_RDWR, 0666)
	defer archiveFile.Close()
	fileInfo, _ := os.Stat(filename)
	bitLength := fileInfo.Size() * 8
	errorBits := generateUniqueRandomNumbers(n, int(bitLength))
	for _, index := range errorBits {
		flipByte := int(float64(index / 8))
		flipMask := byte(1 << (index % 8))
		buffer := make([]byte, 1)
		archiveFile.ReadAt(buffer, int64(flipByte))
		buffer[0] = buffer[0] ^ flipMask
		archiveFile.WriteAt(buffer, int64(flipByte))
	}
}

func generateUniqueRandomNumbers(n, upperLimit int) []int {
	uniqueNumbers := make(map[int]struct{})
	for len(uniqueNumbers) < n {
		randomNumber := rand.Intn(upperLimit)
		uniqueNumbers[randomNumber] = struct{}{}
	}
	result := make([]int, 0, n)
	for num := range uniqueNumbers {
		result = append(result, num)
	}
	return result
}
