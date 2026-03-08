#!/bin/sh
trap 'echo "Ignoring SIGTERM"' TERM
echo "Starting stubborn server"
sleep 30
