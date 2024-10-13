package regen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

// utility functions

func xorParity(data *[][]byte) []byte {
	parity := make([]byte, len((*data)[0]))
	for _, row := range *data {
		for i, val := range row {
			parity[i] ^= val
		}
	}
	return parity
}

func fletcher16(data []byte) uint16 {
	var sum1, sum2 uint16
	for _, value := range data {
		sum1 = (sum1 + uint16(value)) % 255
		sum2 = (sum2 + sum1) % 255
	}
	return (sum2 << 8) | sum1
}

func calculateSHA256(file *os.File) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)
	return hashString, nil
}

func generateCombinations(input []int, limit int) [][]int {
	var result [][]int
	for length := len(input); length > 0; length-- {
		generateRecursive(&result, input, []int{}, 0, length, limit)
	}
	return result
}

func generateRecursive(result *[][]int, input []int, current []int, start int,
	length int, limit int) {
	if len(*result) >= limit {
		return
	}
	if length == 0 {
		combination := make([]int, len(current))
		copy(combination, current)
		*result = append(*result, combination)
		return
	}
	for i := start; i <= len(input)-length; i++ {
		current = append(current, input[i])
		generateRecursive(result, input, current, i+1, length-1, limit)
		// remove the last element for backtracking
		current = current[:len(current)-1]
	}
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
