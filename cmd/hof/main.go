package main

import (
	"os"

	"github.com/hofstadter-io/hof/cmd/hof/cmd"
)

func main() {
	os.Setenv("CUE_EXPERIMENT", "evalv3,modules=0")

	cmd.RunExit()
}
