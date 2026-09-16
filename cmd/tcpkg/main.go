package main

import (
	"fmt"
	"os"

	"github.com/secnova-ai/ai-siem-toolkit/internal/cli"
)

var version = "dev"

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout, version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
