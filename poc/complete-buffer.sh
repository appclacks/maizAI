#!/bin/bash

set -e
system=$(cat <<EOF
You are an IDE helper tool. You should only provide code in your answer, nothing else.
The code will be used directly in user's IDE. You should use common programming patterns and best practices to answer.
Don't ask additional information, don't explain the code, just answer with code that can be immediately used. NEVER give an output with quotes or backticks for the file extension, for example don't add \`\`\` at the top of bottom of code blocks.

EOF
)

prompt=$1
fileName=$2
args=()
args+=(--provider $PROVIDER)
args+=(--model $MODEL)

echo "source context: '${SOURCE_CONTEXT}'"
if [ ! -z "${SOURCE_CONTEXT}" ]; then
    sourceContextID=$(maizai context get --name "$SOURCE_CONTEXT" | jq -r '.id')
    args+=(--source-context $sourceContextID)
fi



echo "pass file content: '${PASS_FILE_CONTENT}'"
if [ "${PASS_FILE_CONTENT}" == "true" ]; then
    fileContent=$(cat $fileName)
    args+=(--message "assistant:$fileContent")
fi

echo $(date)
#echo "args: ${args[@]}"
echo
echo "Updating file $fileName..."
start=`date +%s`

maizai conversation \
       --system "${system}" \
       --message "user:$prompt" \
       "${args[@]}" | jq -r '.result.[0].text' >> $fileName

echo
end=`date +%s`
duration=$((end-start))
echo "File updated in $duration seconds"
