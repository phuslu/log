#!/bin/bash

case $1 in
	verify )
		go test -v ./verify -args -useWarnings
		;;
	bench )
		go test -bench=. ./bench
		;;
	* )
		echo "Usage: ./run.sh verify|bench"
		;;
esac

