package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/jo3-l/yagfuncdata"
)

var (
	timeout = flag.Duration("timeout", 5*time.Second, "timeout for fetching data.")
)

func usage() {
	fmt.Fprintln(os.Stderr, `lytfs: list available YAGPDB template function names
	
usage: lytfs [owner/repo@branch] [-timeout duration]

To authenticate your requests, pass a GitHub personal access token via the LYTFS_GITHUB_TOKEN environment variable.`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	owner := "botlabs-gg"
	repo := "yagpdb"
	branch := "master"

	target := flag.Arg(0)
	if target != "" {
		var targetRegex = regexp.MustCompile(`^([^/]+)/([^@]+)@(.+)$`)
		matches := targetRegex.FindStringSubmatch(target)
		if len(matches) != 4 {
			fmt.Fprintln(os.Stderr, "invalid target format")
			os.Exit(1)
		}

		owner = matches[1]
		repo = matches[2]
		branch = matches[3]
	}

	fcp := yagfuncdata.NewGitHubFileProvider(github.NewClient(nil), owner, repo, branch)

	if token := os.Getenv("LYTFS_GITHUB_TOKEN"); token != "" {
		fcp = yagfuncdata.NewGitHubFileProvider(github.NewClient(nil).WithAuthToken(token), owner, repo, branch)
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
