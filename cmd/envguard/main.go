package main

import (
	"errors"
	"os"

	"github.com/envguard/envguard/internal/cli"
)

const version = "0.1.6"

func main() {
	if err := cli.Execute(version); err != nil {
		if errors.Is(err, cli.ErrValidationFailed) {
			os.Exit(1)
		}
		if errors.Is(err, cli.ErrIO) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}
