#!/bin/bash
set -e

# Build the binary
echo "Building cwgo..."
go build -o cwgo cwgo.go

# Determine GOBIN
GOBIN=$(go env GOBIN)
if [ -z "$GOBIN" ]; then
    GOPATH=$(go env GOPATH)
    if [ -z "$GOPATH" ]; then
        GOPATH=$HOME/go
    fi
    GOBIN=$GOPATH/bin
fi

# Ensure GOBIN exists
mkdir -p "$GOBIN"

# Move and replace
echo "Installing to $GOBIN/cwgo..."
mv cwgo "$GOBIN/cwgo"

echo "Success!"

