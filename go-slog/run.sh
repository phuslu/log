#!/bin/bash

case $1 in
	verify )
		go test -v -args -useWarnings creator.go verify_test.go
		;;
	bench )
		go test -v -bench=. creator.go bench_test.go
		;;
	* )
		echo "Usage: ./run.sh verify|bench"
		;;
esac

