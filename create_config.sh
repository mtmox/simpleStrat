#!/bin/bash

# Check if all arguments are provided
if [ "$#" -ne 6 ]; then
    echo "Usage: $0 <pair> <market> <exchange> <entry_signal> <max_position> <testnet>"
    exit 1
fi

# Assign arguments to variables
pair=$1
market=$2
exchange=$3
entry_signal=$4
max_position=$5
testnte=$6

# Create the config directory if it doesn't exist
mkdir -p config

# Generate a filename based on the pair and market, with an underscore between them
filename="config/${pair}_${market}.json"

# Create the JSON content
json_content=$(cat <<EOF
{
    "pair": "$pair",
    "market": "$market",
    "exchange": "$exchange",
    "entry_signal": "$entry_signal",
    "max_position": $max_position,
    "use_testnet": "$testnet",
}
EOF
)

# Write the JSON content to the file
echo "$json_content" > "$filename"

echo "Configuration file created: $filename"
