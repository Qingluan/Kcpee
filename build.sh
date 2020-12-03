#!/bin/bash

which apt 2>/dev/null
if [ $? -eq 0 ] ; then
	
	sudo apt-get install -y gcc libgtk-3-dev libappindicator3-dev
	sudo apt install -y zenity
fi

go build -ldflags="-s -w "
mv Kcpee $GOPATH/bin/
mkdir -p $GOPATH/Imgs
cp ICON@256x256.png $GOPATH/Imgs/Kcpee.png 
cp Kcpee.desktop $HOME/Desktop/
