#!/bin/bash

# Run slog verification tests with -useWarnings.
#  go_slog/verify.sh
# Save verification output to /tmp/go-slog/verify.txt.

clear
mkdir -p /tmp/go-slog
go test -v ./go_slog/verify -args -useWarnings | # Run verification tests  \
   tee /tmp/go-slog/verify.txt            # Save benchmark output
