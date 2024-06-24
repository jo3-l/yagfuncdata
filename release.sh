#!/bin/bash

set -euo pipefail

readonly VERSION="$1"
readonly OLD_VERSION="$(git describe --tags)"
readonly CUR_DIR="$(pwd)"

refresh-go-cache() {
    echo "Refreshing pkg.go.dev cache..."
    OLD_GOPROXY="$GOPROXY"
    OLD_GO111MODULE="$GO111MODULE"
    unset GOPROXY
    unset GO111MODULE

    export GOPROXY="https://proxy.golang.org"
    export GO111MODULE=on

    mkdir "/tmp/dummy"
    cd "/tmp/dummy"

    go mod init dummy
    go get github.com/jo3-l/yagfuncdata/cmd/lytfs@"$VERSION"

    cd "$CUR_DIR"
    rm -rf "/tmp/dummy"

    export GOPROXY="$OLD_GOPROXY"
    export GO111MODULE="$OLD_GO111MODULE"
}

make_tag() {
    echo "Creating tag $VERSION..."
    git tag "$VERSION"
    git push origin "$VERSION"
}

make_gh_release() {
    if ! command -v gh &> /dev/null; then
        echo "gh-cli is not installed; either install it or create the release manually on the web interface."
        return
    fi

    echo "Creating GitHub release using gh-cli..."
    gh release create "$VERSION" -t "$VERSION" --notes-start-tag "$OLD_VERSION"
}

usage() {
    echo "Usage: "$0" <version>"
    exit 1
}

main() {
    if [ -z "$VERSION" ]; then
        usage
    fi

    make_tag
    refresh-go-cache
    make_gh_release
}

main
