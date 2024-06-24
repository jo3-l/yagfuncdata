package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/jo3-l/yagfuncdata"
)

var timeout = flag.Duration("timeout", 5*time.Second, "timeout for fetching data")
var githubRepoRe = regexp.MustCompile("^(.+)/(.+)@(.+)$")

func usage() {
	fmt.Fprintln(os.Stderr, `lytfs: list available YAGPDB template function names
	
usage: lytfs [owner/repo@branch] [-timeout duration]

To authenticate your requests, pass a GitHub personal access token via the LYTFS_GITHUB_TOKEN environment variable.`)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("lytfs: ")

	flag.Usage = usage
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var (
		owner  = "botlabs-gg"
		repo   = "yagpdb"
		branch = "master"
	)

	if flag.NArg() > 0 {
		repoArg := flag.Arg(0)
		matches := githubRepoRe.FindStringSubmatch(repoArg)
		if matches == nil {
			log.Fatalf("invalid source repository %q (format: owner/repo@branch)\n", repoArg)
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
		log.Fatalln(err)
	}

	for _, name := range funcs {
		fmt.Println(name)
	}
}
