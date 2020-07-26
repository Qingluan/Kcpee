package utils

// // Copyright 2015 Daniel Theophanes.
// // Use of this source code is governed by a zlib-style
// // license that can be found in the LICENSE file.

// // simple does nothing except block while running the service.

// import (
// 	"errors"

// 	"github.com/CodyGuo/win"
// 	"github.com/kardianos/service"
// )

// var logger service.Logger

// func execWindowsBackground(cmd string) error {
// 	lpCmdLine := win.StringToBytePtr(cmd)
// 	ret := win.WinExec(lpCmdLine, win.SW_HIDE)
// 	if ret <= 31 {
// 		return errors.New("run windows cmd error")
// 	}
// 	return nil
// }
