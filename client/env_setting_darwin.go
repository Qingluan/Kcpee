// +build darwin,!linux,!windows,!js

package client

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gen2brain/dlgs"
)

var RAW_PROFILE = ""

var (
	GlobalStatus = false
)

func IfProxyStart() bool {
	result := os.Getenv("http_proxy")
	if result == "" {
		if os.Getenv("HTTP_PROXY") != "" {
			return true
		}
	} else {
		return true
	}
	return false
}

func InitDesktop() {
	KCPPATH := os.Args[0]
	DEKPATH := filepath.Join(os.Getenv("HOME"), "Desktop")
	if _, err := os.Stat(DEKPATH); err != nil {
		DEKPATH = filepath.Join(os.Getenv("HOME"), "桌面")
	}
	DEKPATH_APP := filepath.Join(DEKPATH, "Kcpee.desktop")

	APP_base := filepath.Join(os.Getenv("HOME"), ".local", "Kcpee")
	if _, err := os.Stat(APP_base); err != nil {
		os.MkdirAll(APP_base, os.ModePerm)
	}
	MAIN_EXE := filepath.Join(APP_base, "Kcpee")
	if data, err := ioutil.ReadFile(KCPPATH); err != nil {
		dlgs.Error("错误", err.Error())
	} else {
		if err := ioutil.WriteFile(MAIN_EXE, data, os.ModePerm); err != nil {
			dlgs.Error("错误", err.Error())
		}
	}

	if _, err := os.Stat(APP_base); err != nil {
		os.MkdirAll(APP_base, os.ModePerm)
	}

	PICPATH := filepath.Join(APP_base, "icon.png")
	dlgs.Info("Initing", "create Desktop app in "+PICPATH)
	cmd := exec.Command("bash", "-c", "wget -c -t 5 'https://gitee.com/dark.H/Kcpee/raw/master/ICON@256x256.png' -O "+PICPATH)
	if aa, err := cmd.Output(); err != nil {
		dlgs.Info("Initing", string(aa)+err.Error())
	}

	INIT := `
[Desktop Entry]
Encoding=UTF-8
Version=1.0
Type=Application
Terminal=false
Exec=sudo ` + MAIN_EXE + `
Name=Kcpee
Icon=` + PICPATH

	if _, err := os.Stat(DEKPATH_APP); err != nil {
		yes, _ := dlgs.Question("Question", "if Create Desktop App icon?", true)
		if yes {
			if err := ioutil.WriteFile(DEKPATH_APP, []byte(INIT), os.ModePerm); err != nil {
				dlgs.Error("错误", err.Error())
			}
		}
	}
}

func ProxySet(localAddr string) {
	// DEKPATH := filepath.Join(os.Getenv("HOME"), "Desktop")
	// if _, err := os.Stat(DEKPATH); err != nil {
	// 	DEKPATH = filepath.Join(os.Getenv("HOME"), "桌面")
	// }
	// DEKPATH_APP := filepath.Join(DEKPATH, "Kcpee.desktop")
	// if _, err := os.Stat(DEKPATH_APP); err != nil {
	// 	dlgs.Info("Init ", DEKPATH_APP)
	// 	InitDesktop()
	// }
	if _, err := os.Stat("/etc/profile"); err != nil {
		return
	}
	RAWB, err := ioutil.ReadFile("/etc/profile")
	if err != nil {
		return
	}
	RAW := string(RAWB)
	if localAddr != "" {
		TMP := `
##### PROXY PROFILE START 
export http_proxy="` + localAddr + `"`
		TMP += `
export https_proxy="` + localAddr + `
##### PROXY PROFILE END
`
		if !strings.Contains(RAW, TMP) {
			RAW += TMP
			//err := ioutil.WriteFile("/etc/profile", []byte(RAW), os.ModePerm)

			_, err := exec.Command("bash", "-c", "sudo cat << EOF > /etc/profile\n"+RAW+"EOF\n").Output()
			if err != nil {
				dlgs.Error("Error Start Global Proxy", err.Error())
			}
			GlobalStatus = true
			dlgs.Info("Infomation", "try to start proxy: "+localAddr)
		}
	} else {
		if strings.Contains(RAW, "##### PROXY PROFILE START") {
			s := strings.SplitN(RAW, "##### PROXY PROFILE START", 2)
			PRE := s[0]
			AFT := strings.TrimSpace(strings.SplitN(s[1], "##### PROXY PROFILE END", 2)[1])
			RAW = PRE + "\n" + AFT
			// err := ioutil.WriteFile("/etc/profile", []byte(RAW), os.ModePerm)
			_, err := exec.Command("bash", "-c", "sudo cat << EOF > /etc/profile\n"+RAW+"EOF\n").Output()
			if err != nil {
				dlgs.Error("Error Start Global Proxy", err.Error())
			}
			//ioutil.WriteFile("/etc/profile", []byte(RAW), os.ModePerm)
			GlobalStatus = false
			dlgs.Info("Infomation", "try to stop proxy: ")
		}
	}

}

func KillKcpee() {
	TMP := `sudo ps aux| grep Kcpee |  egrep -v '(grep|egrep)' | awk '{ print $2}' | xargs kill -9 `
	exec.Command("bash", "-c", TMP).Output()
}
