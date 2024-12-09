#!/usr/bin/env bash
# Copyright 2023 buildkit-syft-scanner authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -eu

GENERATOR=$1

for example in "${@:2}"; do
  example=$(basename "$example")
  echo "[-] Building example ${example}..."

  if [[ ${example} == "npm-lock" ]]; then
     docker buildx build "./examples/${example}" --sbom="generator=${GENERATOR},SELECT_CATALOGERS=+javascript-lock-cataloger" --output="./examples/${example}/build"
  else
     docker buildx build "./examples/${example}" --sbom="generator=${GENERATOR}" --output="./examples/${example}/build"
  fi


  echo "[-] Checking example ${example}..."
  for file in "./examples/${example}"/checks/*.json; do
    echo "  [-] Checking schema ${file}..."
    go run ./cmd/check "$PWD/$file" $PWD/examples/${example}/build/${file#"./examples/${example}/checks/"}
  done
  
  echo ""
done
