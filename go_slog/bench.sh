#!/bin/bash

# Run slog benchmark tests.
#  go_slog/bench.sh
# Save benchmark output to /tmp/go-slog/bench.txt.
# Use madkins23/go-slog/cmd/tabular or cmd/server to see pretty results.

clear
mkdir -p /tmp/go-slog
go test -bench=. ./go_slog/bench       | # Run benchmark tests                \
   tee /dev/tty                  | # Show progress to user in real time \
   tee /tmp/go-slog/bench.txt      # Save benchmark output