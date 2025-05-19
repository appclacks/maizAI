#!/bin/bash

set -e

contextName=$1
filePath=$2

maizai context message add --context-name $contextName --message-from-file "user:$filePath"
