#!/usr/bin/env bash


ALL_OS=(darwin linux windows dragonfly freebsd netbsd openbsd)
ALL_ARCH=(amd64 386 arm arm64)

DIST_DIR='./dist'


for OS in "${ALL_OS[@]}"
do
	for ARCH in "${ALL_ARCH[@]}"
	do
		OUTFILE="arpwatch-$OS-$ARCH"
		if [[ "$OS" == "windows" ]]
		then
			OUTFILE="$OUTFILE".exe
		fi
		export GOOS="$OS"
		export GOARCH="$ARCH"
		go build -o "$DIST_DIR"/"$OUTFILE" arpwatch.go
	done
done
