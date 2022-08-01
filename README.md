# `yagfuncdata`

> Automatically up-to-date information regarding available YAGPDB template functions.

## Why use `yagfuncdata`?

`yagfuncdata` aims to simplify workflows of projects that require a list of YAGPDB template functions, such as language support extensions for editors. For example, one could imagine setting up a cronjob in CI to run `yagfuncdata` regularly to check for updates and regenerate some asset file if anything changed.

Note, however, that it is not recommended to run `yagfuncdata` directly as part of an application as the command is not infallible. Changes to the structure of relevant files defining template functions may cause the generated list to be incomplete or downright wrong. (We elaborate on why this is the case below.) Thus, its output should be manually vetted by a human against a reliable baseline.

## How does it work?

Instead of hardcoding relevant data (thereby necessitating manual synchronization when upstream sources change), `yagfuncdata` instead automatically generates it on-demand based on the content of files in the YAGPDB project that define template functions. As such, so long as the structure of the files does not change, modifications to the set of available functions show up instantly.

By way of example, the `StandardFuncMap` in [`common/templates/context.go`](https://github.com/botlabs-gg/yagpdb/blob/master/common/templates/context.go) defines the base set of standard functions. It is structured like such:

```go
var (
	StandardFuncMap = map[string]interface{}{
		"name1": func1,
		"name2": func2,
		[...]
	}
)
```

Notice that every key in the map literal corresponds to a function name. Thus, in order to generate a list of standard context functions, it suffices to parse the content of `common/templates/context.go`, look up the declaration of `StandardFuncMap`, and find all keys present in the map literal. This is precisely the approach that `yagfuncdata` takes in `BaseContextFuncSource.collectStandardContextFuncs`.

## Installation

Assuming that Go is installed, run the following on the command-line:

```
go install github.com/jo3-l/yagfuncdata/cmd/lytfs@latest
```

Afterward, running `lytfs` without any arguments will output a newline-delimited list of template function names based on definitions on the master branch of the YAGPDB repository.

## License

`yagfuncdata` is made available under the MIT license.
