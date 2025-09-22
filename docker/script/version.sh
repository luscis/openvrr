#!/bin/bash

ver=$(git describe --tags --abbrev=0 --match 'v*')
if [ $? -eq 0 ]; then
	echo $ver
	exit 0
fi

echo v$(date +%y%m%d)