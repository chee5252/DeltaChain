#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
dchdir="$workspace/src/github.com/deltachaineum"
if [ ! -L "$dchdir/go-deltachaineum" ]; then
    mkdir -p "$dchdir"
    cd "$dchdir"
    ln -s ../../../../../. go-deltachaineum
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$dchdir/go-deltachaineum"
PWD="$dchdir/go-deltachaineum"

# Launch the arguments with the configured environment.
exec "$@"
