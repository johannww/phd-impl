#!/usr/bin/env bash

mkdir -p ./docs/diagrams

# TODO: add more folders
dirs="bids,credits,properties"

go-plantuml generate \
    -d "${dirs}" \
    -o ./docs/diagrams/diagram.puml

plantuml -d ./docs/diagrams/diagram.puml

