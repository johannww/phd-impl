#/bin/env bash

DIRECTORY=${1:-.}

pushd "$DIRECTORY" || exit 1

# fetch and print all struct definitions
gopls workspace_symbol . | rg Struct | cut -d' ' -f1 | while read sym; do
    gopls definition "$sym" | sed -n '/type .* struct/,/^}/p' &
done

wait

popd
