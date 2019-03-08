#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
mandir="$workspace/src/github.com/MatrixAINetwork"
if [ ! -L "$mandir/go-matrix" ]; then
    mkdir -p "$mandir"
    cd "$mandir"
    ln -s ../../../../../. go-matrix
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$mandir/go-matrix"
PWD="$mandir/go-matrix"

# Launch the arguments with the configured environment.
exec "$@"
