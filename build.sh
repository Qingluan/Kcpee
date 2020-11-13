#!/bin/bash

which apt 2>/dev/null
if [ $? -eq 0 ] ; then
	
	sudo apt-get install -y gcc libgtk-3-dev libappindicator3-dev
	
fi

go build -ldflags="-s -w "
