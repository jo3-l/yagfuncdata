package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jo3-l/yagfuncdata"
)

var (
	timeout = flag.Duration("timeout", 5*time.Second, "timeout for fetching data; default: 5s")
)

func usage() {
	fmt.Fprintln(os.Stderr, `lytfs: list available YAGPDB template function names
	
usage: lytfs [-timeout duration]`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	sources := yagfuncdata.DefaultSources(yagfuncdata.DefaultFileContentProvider)
	funcs, err := yagfuncdata.Fetch(ctx, sources)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for _, name := range funcs {
		fmt.Println(name)
	}
}
