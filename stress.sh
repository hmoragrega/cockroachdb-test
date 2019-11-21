#!/usr/bin/env bash

echo "Press [CTRL+C] to stop... (before it's too late)"

INSTANCES=100

while :
do
	echo "Launching an app instance"
	./goapp/main &

	sleep 20

	INSTANCES=$(( $INSTANCES - 1 ))
	if [ $INSTANCES -eq 0 ]; then
    break
	fi
done

echo "Stress test finished"

trap "trap - SIGTERM && kill -- -$$" SIGINT SIGTERM EXIT
