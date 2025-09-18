#!/bin/bash

BINARY_NAME="network"

if [ -f "$BINARY_NAME" ]; then
	echo "removing old binary"
	rm "$BINARY_NAME"
fi

go build -o "$BINARY_NAME" .

if [ $? -eq 0 ]; then
	./"$BINARY_NAME"
	EXIT_CODE=$?
	exit $EXIT_CODE
else
	echo "Build failed. Not running." >&2
	exit 1
fi

