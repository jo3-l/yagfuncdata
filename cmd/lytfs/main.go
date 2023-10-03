package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/jo3-l/yagfuncdata"
)

var (
	timeout = flag.Duration("timeout", 5*time.Second, "timeout for fetching data; default: 5s")
)

func usage() {
	fmt.Fprintln(os.Stderr, `lytfs: list available YAGPDB template function names
	
usage: lytfs [-timeout duration]

To authenticate your requests, pass a GitHub personal access token via the YAGFUNCDATA_AUTH_TOKEN environment variable.`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var sources []yagfuncdata.Source
	if token := os.Getenv("YAGFUNCDATA_AUTH_TOKEN"); token != "" {
		fcp := yagfuncdata.NewGitHubFileProvider(github.NewClient(nil).WithAuthToken(token), "botlabs-gg", "yagpdb", "master")
		sources = yagfuncdata.DefaultSources(fcp)
	} else {
		sources = yagfuncdata.DefaultSources(yagfuncdata.DefaultFileContentProvider)
	}
	funcs, err := yagfuncdata.Fetch(ctx, sources)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, name := range funcs {
		fmt.Println(name)
	}
}
