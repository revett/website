package main

import (
	"fmt"
	"os"

	"github.com/revett/website/generator/site"
)

func main() {
	cmd := "build"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	var err error
	switch cmd {
	case "build":
		err = site.Build()
	case "serve":
		err = site.Serve(":8080")
	case "check":
		err = site.Check()
	default:
		err = fmt.Errorf("unknown command %q (want build, serve, or check)", cmd)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
