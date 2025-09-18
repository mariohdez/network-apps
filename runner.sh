#!/bin/bash

if [ "$#" -ne 1 ]; then
	echo "usage: $0 {updclient|updserver}"
	exit 1
fi


PROJECT_NAME="$1"

case "$PROJECT_NAME" in	
	updclient)
		;;
	updserver)
		;;
	*)
		echo "usage: $0 {updclient|updserver}"
		exit 1
		;;
esac


BINARY_NAME="$PROJECT_NAME.exe"

if [ -f "$BINARY_NAME" ]; then
	echo "removing old binary"
	rm "$BINARY_NAME"
fi

go build -o "$BINARY_NAME" "/Users/mariohernandez/development/network-apps/cmd/$PROJECT_NAME"

if [ $? -eq 0 ]; then
	./"$BINARY_NAME"
	EXIT_CODE=$?
	exit $EXIT_CODE
else
	echo "Build failed. Not running." >&2
	exit 1
fi

