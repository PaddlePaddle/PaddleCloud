#!/bin/sh
#
# Copyright 2021 The Kubeflow Authors
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

set -ex

if ! [ -f go-licenses.yaml ]; then
    echo "go-licenses.yaml not found"
    exit 1
fi

if [ ! -f bin/go-licenses ] || [ ! -d bin/licenses ]; then

    DOWNLOAD_URL="https://github.com/Bobgy/go-licenses/releases/download/v0.0.0-2021-06-27/go-licenses-linux.tar.gz"
    if which wget; then
    	wget "${DOWNLOAD_URL}"
    else
    	curl -LO "${DOWNLOAD_URL}"
    fi
    
    tar xvf go-licenses-linux.tar.gz
    mv go-licenses bin/
    mv licenses bin/
    rm -f go-licenses-linux.tar.gz

fi

if [ -d third_party/NOTICES ]; then
    rm -rf third_party/NOTICES 
fi

bin/go-licenses save third_party/licenses/licenses.csv --save_path third_party/NOTICES 
