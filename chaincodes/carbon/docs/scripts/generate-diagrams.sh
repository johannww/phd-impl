#!/usr/bin/env bash

mkdir -p ./docs/diagrams

# TODO: add more folders
dirs="auction,bids,companies,credits,data,identities,payment,properties,state,tee,vegetation"

go-plantuml generate \
    -d "${dirs}" \
    -o ./docs/diagrams/diagram.puml

echo "plantuml file generated ./docs/diagrams/diagram.puml"

plantuml -d ./docs/diagrams/diagram.puml -tsvg

echo "diagrams generated in ./docs/diagrams/diagram.svg"

