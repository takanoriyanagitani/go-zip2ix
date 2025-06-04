#!/bin/sh

tinygo \
	build \
	-o ./zip2ix2asn1.wasm \
	-target=wasip1 \
	-opt=z \
	-no-debug \
	./main.go
