#!/bin/bash
set -e

targets=`go list ./...`
cwd=`pwd`
v=$1

[ -z $v ] && echo "No version specified" && exit 1

for target in $targets; do
	cd $cwd
	
	#target="$cwd${target/_$cwd\///}"
	target=${target/github.com\/stampzilla\/stampzilla-go\//}

	if [ -e "$target/.goxc.json" ] 
	then
		echo "- Building $target"
		cd $target
		goxc -d "$cwd/build" -pv=$v
	else
		echo "- Skipping $target (no .goxc.json file)"
	fi
done


rename 's/(.*)/$1-arm/' $cwd/build/$v/linux_arm/*
rename 's/(.*)/$1-amd64/' $cwd/build/$v/linux_amd64/*
mkdir $cwd/build/$v/prepare
mv $cwd/build/$v/linux_arm/* $cwd/build/$v/prepare
mv $cwd/build/$v/linux_amd64/* $cwd/build/$v/prepare
sha512sum $cwd/build/$v/prepare/* > $cwd/build/$v/prepare/checksum 

