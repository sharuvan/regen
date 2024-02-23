/*
Regen uses data redundancy techniques to generate redundant data
on archive files and uses the produced regen file to restore the
integrity of the original archive file when verification fails
*/

package main

import "github.com/sharuvan/regen/cmd"

func main() {
	cmd.Execute()
}
