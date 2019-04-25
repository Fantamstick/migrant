#!/usr/bin/env bash
while true; do echo -e "HTTP/1.1 200 OK\n\n $(date)" | nc -l -N 1234; done