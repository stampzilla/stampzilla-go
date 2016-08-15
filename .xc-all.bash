#!/bin/bash

targets=`go list ./...`
cwd=`pwd`

for target in $targets; do
	cd $cwd

	target=${target/github.com\/stampzilla\/stampzilla-go\//}
	if [ -e "$target/.goxc.json" ] 
	then
		cd $target
		goxc -d "$cwd/build"
	fi
done
