#!/bin/bash

targets=`go list ./...`
cwd=`pwd`
v=$1

[ -z $v ] && exit 1

for target in $targets; do
	cd $cwd
	
	#target="$cwd${target/_$cwd\///}"
	target=${target/github.com\/stampzilla\/stampzilla-go\//}

	echo $target
	if [ -e "$target/.goxc.json" ] 
	then
		cd $target
		goxc -d "$cwd/build" -pv=$v
	fi
done


rename 's/(.*)/$1-arm/' $cwd/build/$v/linux_arm/*
rename 's/(.*)/$1-amd64/' $cwd/build/$v/linux_amd64/*
