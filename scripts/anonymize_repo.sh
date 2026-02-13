#!/bin/env bash
# rm -rf blind_repo
mkdir blind_repo

COPIABLE=(
    Makefile
    README.md
    chaincodes
    scripts
    tee_auction
)

cp -r ${COPIABLE[@]} blind_repo

pushd blind_repo
touch .nosync # prevent syncing to cloud storage

delete_dirs=$(fd --type d --no-ignore vendor)
delete_dirs+=(
    ./tee_auction/LEARNING.md
    ./chaincodes/carbon/tests/fabric-samples/
    ./scripts/anonymize_repo.sh
)
for dir in ${delete_dirs[@]}; do
    echo "Removing $dir"
    rm -rf "$dir"
done

echo "Cloning modified fabric-chaincode and gateway"
repo_names=(
    fabric-chaincode-go
    fabric-gateway
)
repo_commits=(
    d967d6ea1875
    c4252b48f5f1
)
for ((i=0; i<${#repo_names[@]}; i++)); do
    repo=${repo_names[i]}
    commit=${repo_commits[i]}
    if [ -d "$repo" ]; then
        echo "Directory $repo already exists, skipping clone"
        continue
    fi
    echo "Cloning $repo"
    git clone https://github.com/johannww/"$repo".git
    pushd ./"$repo" || continue
    git checkout "$commit"
    popd || continue
    rm -rf ./"$repo"/.git
done

rg --files-with-matches "github.com/johannww/phd-impl" | while read file; do
    echo "Anonymizing $file"
    sed -i 's/github\.com\/johannww\/phd-impl/github-com\/anonymized-repo/g' "$file"
done

rg --files-with-matches "ghcr.io/johannww/phd-impl" | while read file; do
    echo "Anonymizing $file"
    sed -i 's/ghcr\.io\/johannww\/phd-impl/ghcr\.io\/anonymized-repo/g' "$file"
done

echo "Anonymizing johannww forks in go.mod files"
mod_files=(
    ./chaincodes/carbon/go.mod
    ./chaincodes/interop/go.mod
    ./tee_auction/go/go.mod
    )
mod_relative_paths=(
    "../../"
    "../../"
    "../../"
)
FABRIC_CC_PATH=./fabric-chaincode-go
FABRIC_GATEWAY_PATH=./fabric-gateway

for ((i=0; i<${#mod_files[@]}; i++)); do
    mod_file=${mod_files[i]}
    base_path=${mod_relative_paths[i]}
    cc_path=$(echo "${base_path}$FABRIC_CC_PATH" | sed 's/\//\\\//g')
    gateway_path=$(echo "${base_path}$FABRIC_GATEWAY_PATH" | sed 's/\//\\\//g')
    echo "Anonymizing $mod_file"
    sed -i "s/github\.com\/johannww\/fabric-chaincode-go.*/${cc_path}/g" "$mod_file"
    sed -i "s/github\.com\/johannww\/fabric-gateway.*/${gateway_path}/g" "$mod_file"
    sed -i 's/github\.com\/johannww\/confidential-sidecar-containers.*//g' "$mod_file"
done

name_in_comments=$(rg -ui --files-with-matches "//.*johann")
for file in ${name_in_comments[@]}; do
    echo "Anonymizing comments in $file"
    sed -i '/\(\/\/.*\)johann/Id' "$file"
done

revendor_dirs=(
    chaincodes/carbon
    chaincodes/interop
    tee_auction/go
)

for dir in "${revendor_dirs[@]}"; do
    echo "Revendoring $dir"
    pushd "./$dir" || continue
    go mod tidy
    go mod vendor
    popd || continue
done

echo "Generate diagrams without johann's name"
make diagrams

echo "Testing that the modified repo can be built and tested"

pushd ./chaincodes/carbon
make test-no-cache &
make cc-docker &
make app &
make test-network &
wait
popd

pushd ./chaincodes/interop
make test-no-cache &
make cc-docker &
wait
popd

pushd ./tee_auction
make docker
popd

rg -ui "johann"

popd

zip -r blind_repo.zip blind_repo
