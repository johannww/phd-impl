#!/usr/bin/env bash

mkdir -p ./docs/diagrams

dirs="go/auction,go/api,go/cmd/auction,go/vendor/github.com/johannww/phd-impl/chaincodes/carbon/auction"

go-plantuml generate \
    -d "${dirs}" \
    -o ./docs/diagrams/diagram.puml

echo "plantuml file generated ./docs/diagrams/diagram.puml"
sed -i 's/github\.com/github-com/g' ./docs/diagrams/diagram.puml

plantuml -d ./docs/diagrams/diagram.puml -tsvg

echo "diagrams generated in ./docs/diagrams/diagram.svg"

