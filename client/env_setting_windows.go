// +build windows,!linux,!darwin,!js

package client

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"

	"github.com/gen2brain/dlgs"
)

var (
	GlobalStatus = false
)

func IfProxyStart() bool {
	TMP := `
	Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings\" -Name ProxyServer
	`
	TMP2 := `
	Get-ItemProperty -Path "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings\" -Name ProxyEnable
	`
	result, err := exec.Command("powershell.exe", "-command", TMP2).Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(result), "ProxyEnable  : 1") {
		return false
	}

	exe := exec.Command("powershell.exe", "-windowstyle", "hidden", "-command", TMP)
	exe.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	result, _ = exe.Output()
	if strings.Contains(string(result), ":10091") {
		return true
	}
	return false
}

func ProxySet(localAddr string) {
	TMP := `
	# Set-Proxy command
	Function SetProxy() {
		Param(
			$Addr = $null,
			[switch]$ApplyToSystem
		)
		
		$env:HTTP_PROXY = $Addr;
		$env:HTTPS_PROXY = $Addr; 
		$env:http_proxy = $Addr;
		$env:https_proxy = $Addr;
	  
		if ($addr -eq $null) {
			[Net.WebRequest]::DefaultWebProxy = New-Object Net.WebProxy;
			if ($ApplyToSystem) { SetSystemProxy $null; }
			Write-Output "Successful unset all proxy variable";
		}
		else {
			[Net.WebRequest]::DefaultWebProxy = New-Object Net.WebProxy $Addr;
			if ($ApplyToSystem) {
				$matchedResult = ValidHttpProxyFormat $Addr;
				# Matched result: [URL Without Protocol][Input String]
				if (-not ($matchedResult -eq $null)) {
					SetSystemProxy $matchedResult.1;
				}
			}
			Write-Output "Successful set proxy as $Addr";
		}
	}
	Function SetSystemProxy($Addr = $null) {
		Write-Output $Addr
		$proxyReg = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Internet Settings";
	
		if ($Addr -eq $null) {
			Set-ItemProperty -Path $proxyReg -Name ProxyEnable -Value 0;
			Set-ItemProperty -Path $proxyReg -Name ProxyOverride -Value "*.local";
			return;
		}
		
		Set-ItemProperty -Path $proxyReg -Name ProxyServer -Value $Addr;
		Set-ItemProperty -Path $proxyReg -Name ProxyEnable -Value 1;
		Set-ItemProperty -Path $proxyReg -Name ProxyOverride -Value "*.local;localhost;192.168.*.*;127.0.0.1";
		
		if ($Addr -eq ""){
			Set-ItemProperty -Path $proxyReg -Name ProxyEnable -Value 0;
			Set-ItemProperty -Path $proxyReg -Name ProxyOverride -Value "*.local";
			return;
		}
	}
	Function ValidHttpProxyFormat ($Addr) {
		$regex = "(?:https?:\/\/)(\w+(?:.\w+)*(?::\d+)?)";
		$result = $Addr -match $regex;
		if ($result -eq $false) {
			throw [System.ArgumentException]"The input $Addr is not a valid HTTP proxy URI.";
		}
	
		return $Matches;
	}
	Set-Alias set-proxy SetProxy;
	` + fmt.Sprintf("SetSystemProxy '%s'", localAddr)
	exe := exec.Command("powershell.exe", "-windowstyle", "hidden", "-Command", TMP)
	exe.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	exe.Output()
	if localAddr != "" {
		GlobalStatus = true
		dlgs.Info("消息", "代理到:"+localAddr)
	} else {
		GlobalStatus = false
		dlgs.Info("Close Proxy", "reset internet http proxy")
	}

}

func KillKcpee() {
	TMP := `
ps | ? { $_.Name -Contains "Kcpee"} | kill
	`
	cmd := exec.Command("powershell.exe", "-windowstyle", "hidden", "-Command", TMP)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	cmd.Output()
}
