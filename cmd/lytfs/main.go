package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/jo3-l/yagfuncdata"
)

var (
	timeout = flag.Duration("timeout", 5*time.Second, "timeout for fetching data.")
	repo    = flag.String("repo", "botlabs-gg/yagpdb", "GitHub repository to fetch data from.")
	branch  = flag.String("branch", "master", "GitHub branch to fetch data from.")
)

func usage() {
	fmt.Fprintln(os.Stderr, `lytfs: list available YAGPDB template function names
	
usage: lytfs [-timeout duration] [-repo owner/repo] [-branch branch]

To authenticate your requests, pass a GitHub personal access token via the LYTFS_GITHUB_TOKEN environment variable.`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	owner := strings.Split(*repo, "/")[0]
	repo := strings.Split(*repo, "/")[1]

	fcp := yagfuncdata.NewGitHubFileProvider(github.NewClient(nil), owner, repo, *branch)

	if token := os.Getenv("LYTFS_GITHUB_TOKEN"); token != "" {
		fcp = yagfuncdata.NewGitHubFileProvider(github.NewClient(nil).WithAuthToken(token), owner, repo, *branch)
	}

	sources := yagfuncdata.DefaultSources(fcp)
	funcs, err := yagfuncdata.Fetch(ctx, sources)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, name := range funcs {
		fmt.Println(name)
	}
}
