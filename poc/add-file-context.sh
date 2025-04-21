#!/bin/bash

set -e

contextName=$1
filePath=$2
fileContent=$(cat $filePath)

contextID=$(maizai context get --name $contextName | jq -r .id)
maizai context message add --id $contextID --message "user:$fileContent"
