rsrc -manifest .\Kcpee.exe.manifest -ico .\ICON@256x256.ico -o .\Kcpee.exe.syso
go build -ldflags="-s -w -H windowsgui"
del Kcpee.exe.syso
