#!/bin/bash

if [ "$#" -ne 1 ]; then
	echo "usage: $0 {udpclient|udpserver}"
	exit 1
fi


CMD="$1"
case "$CMD" in	
	udpclient)
		;;
	udpserver)
		;;
	*)
		echo "usage: $0 {udpclient|udpserver}"
		exit 1
		;;
esac


BINARY_NAME="$CMD.exe"
if [ -f "$BINARY_NAME" ]; then
	rm "$BINARY_NAME"
fi


go build -o "$BINARY_NAME" "/Users/mariohernandez/development/network-apps/cmd/$CMD"
if [ $? -eq 0 ]; then
	./"$BINARY_NAME"
	exit $?
else
	echo "Build failed. Not running." >&2
	exit 1
fi

