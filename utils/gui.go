package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gen2brain/dlgs"
)

func testIfStart() bool {
	cmd := exec.Command(os.Args[0], "-book.ls")
	cmd.Env = os.Environ()
	data, err := cmd.Output()
	if err != nil {
		log.Println(err)
		// dlgs.Info("Pid", fmt.Sprintln(err))
		return false
	}
	println(string(data))
	// dlgs.Info("Pid", string(data))

	if strings.Contains(string(data), "json:unexpected") {
		return false
	}
	return true
}

func execs(cmds string, std bool) (output string) {
	cmd := exec.Command("bash", "-c", cmds)
	cmd.Env = os.Environ()
	if strings.HasPrefix(cmds, "Kcpee") {
		msg := strings.Split(cmds, " ")
		// dlgs.Info("show", cmds)
		cmd = exec.Command(os.Args[0], msg[1:]...)
	}
	if std {
		var stdout bytes.Buffer
		// cmd.Stdout = &stdout
		err := cmd.Start()
		if err != nil {
			// dlgs.Info("Pid", fmt.Sprintln(err))
		}
		// time.Sleep(1 * time.Second)
		return fmt.Sprintf("%s", string(stdout.Bytes()))
	}
	data, err := cmd.Output()
	if err != nil {
		log.Println(err)
		// dlgs.Info("Pid", fmt.Sprintln(err))
	}
	output = strings.TrimSpace(string(data))
	println("output:", output)
	return

}

func RunGui(global func()) {
	if testIfStart() {
		items := []string{"Global Mode", "Stop Kcp", "Auto Mode", "Flow Mode"}
		items_2 := strings.Split(execs("Kcpee -book.ls", false), "\n")
		items = append(items, items_2...)

		item, s, err := dlgs.List("Kcpee", "Select route:", items)
		if item == "Stop Kcp" {
			execs("Kcpee -book.stop ", false)
		} else if item == "Global Mode" {
			global()
		} else if item == "Flow Mode" {
			execs("Kcpee -book.flow ", false)
		} else if item != "Auto Mode" && s {
			execs("Kcpee -book.single "+strings.SplitN(item, " ", 2)[0], false)
		} else {
			execs("Kcpee -book.auto", false)
		}
		if err != nil {
			panic(err)
		}
		if !s {
			os.Exit(0)
		}
	} else {

		name, use, err := dlgs.Entry("Name", "Enter your name:", "dark.H/kcpconfig")
		if !use {
			os.Exit(0)
		}
		passwd, _, err := dlgs.Password("Password", "Enter your password:")
		if err != nil {
			panic(err)
		}
		if runtime.GOOS == "windows" {
			execs("Kcpee -Auth -name "+name+" -pwd "+passwd, true)
			// execWindowsBackground("Kcpee -Auth -name " + name + " -pwd " + passwd)
		} else {
			execs("Kcpee -Auth -name "+name+" -pwd "+passwd, true)
		}
		// dlgs.Info("Pid", msg)

		// time.Sleep(5 * time.Second)
	}
	return
}
