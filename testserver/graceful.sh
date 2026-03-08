#!/bin/sh
trap 'echo "Received SIGTERM, exiting"; exit 0' TERM
echo "Starting graceful server"
sleep 30
