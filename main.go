package main

import (
	"os"

	"github.com/StatusCakeDev/cakectl/internal/cmd"
	cliruntime "github.com/tomasbasham/cli-runtime"
)

func main() {
	command := cmd.NewRootCommand()
	if code := cliruntime.Run(command); code != 0 {
		os.Exit(code)
	}
}
