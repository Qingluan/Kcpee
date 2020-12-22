#!/bin/bash
go get github.com/zhangya4548/macapp

mkdir -p /tmp/assets &&
  go build -ldflags="-s -w" &&
	mv Kcpee /tmp/assets/ &&
		macapp -bin Kcpee -dmg Kcpee.dmg -icon "ICON@256x256.png" -identifier "com.Proxy.Kcpee" -name "Kcpee" -o /tmp -assets /tmp/assets
