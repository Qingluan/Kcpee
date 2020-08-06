package utils

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type DebugLog bool

var Debug DebugLog
var dbgLog = log.New(os.Stdout, "[DEBUG] ", log.Ltime)
var mypid = os.Getpid()
var SpeedMsgCHANNEL = make(chan Statistic, 30)

var FGCOLORS = []func(a ...interface{}) string{
	color.New(color.FgYellow, color.Bold).SprintFunc(),
	color.New(color.FgRed, color.Bold).SprintFunc(),
	color.New(color.FgGreen, color.Bold).SprintFunc(),
	color.New(color.FgBlue, color.Bold).SprintFunc(),
}
var BGCOLORS = []func(a ...interface{}) string{
	color.New(color.BgYellow, color.Bold).SprintFunc(),
	color.New(color.BgRed, color.Bold).SprintFunc(),
	color.New(color.BgGreen, color.Bold).SprintFunc(),
	color.New(color.BgBlue, color.Bold).SprintFunc(),
}

func KillProcess() (err error) {
	ColorL("Killing browser process")
	if runtime.GOOS == "windows" {
		kill := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(mypid))
		err = kill.Run()
	} else {
		kill := exec.Command("kill", "/T", "/F", "/PID", strconv.Itoa(mypid))
		err = kill.Run()
	}

	if err != nil {
		ColorL("Error killing chromium process")
	}

	return err
}
func Stat(cmd string, s bool) {

	if s {
		f := color.New(color.FgGreen, color.BlinkSlow).SprintFunc()
		fmt.Println("[", f("T"), "]", cmd)
	} else {
		f := color.New(color.FgRed, color.BlinkSlow).SprintFunc()
		fmt.Println("[", f("T"), "]", cmd)
	}

}

func Md5Str(buf []byte) string {
	// return "No md5"
	h := md5.New()
	h.Write(buf)
	return hex.EncodeToString(h.Sum(nil))
}

func (d DebugLog) Printf(format string, args ...interface{}) {
	if d {
		dbgLog.Printf(format, args...)
	}
}

func (d DebugLog) Println(args ...interface{}) {
	if d {
		dbgLog.Println(args...)
	}
}

func ColorD(args interface{}, join ...string) {

	if b, err := json.Marshal(args); err == nil {
		var data map[string]interface{}
		// yellow := FGCOLORS[0]
		if err := json.Unmarshal(b, &data); err == nil {
			var S []string
			c := 0
			for k, v := range data {
				// ColorD(data)
				S = append(S, fmt.Sprint(k, ": ", FGCOLORS[c](v)))
				c++
				c %= len(BGCOLORS)
			}
			if len(join) == 0 {
				fmt.Println(strings.Join(S, "\n"))
			}

		}
	}
}

func ColorM(args ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	blue := color.New(color.FgBlue).SprintFunc()
	var S string
	useR := false
	if args[len(args)-1] == "\r" {
		useR = true
		args = args[:len(args)-1]
	}
	for i, arg := range args {
		switch i % 4 {
		case 0:
			S += green(fmt.Sprint(arg))
		case 1:
			S += yellow(fmt.Sprint(arg))
		case 2:
			S += blue(fmt.Sprint(arg))
		case 3:
			S += red(fmt.Sprint(arg))
		}
	}
	if useR {
		fmt.Print(S, "                   \r")
	} else {
		fmt.Println(S)
	}

}

func ColorL(args ...interface{}) {
	// yellow := color.New(color.FgYellow).SprintFunc()
	// red := color.New(color.FgRed).SprintFunc()
	// green := color.New(color.FgGreen).SprintFunc()
	// blue := color.New(color.FgBlue).SprintFunc()
	// var S string
	// useR := false
	// if args[len(args)-1] == "\r" {
	// 	useR = true
	// 	args = args[:len(args)-1]
	// }
	// for i, arg := range args {
	// 	switch i % 4 {
	// 	case 0:
	// 		S += green(fmt.Sprint("[", arg, "]"))
	// 	case 1:
	// 		S += yellow(fmt.Sprint("[", arg, "]"))
	// 	case 2:
	// 		S += blue(fmt.Sprint("[", arg, "]"))
	// 	case 3:
	// 		S += red(fmt.Sprint("[", arg, "]"))
	// 	}
	// }
	// if useR {
	// 	fmt.Print(S, "                   \r")
	// } else {
	// 	fmt.Println(S)
	// }

}

func ColorE(args ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	var S string
	for _, arg := range args {
		S += yellow(fmt.Sprint(arg))
	}
	log.Print(red("[Err]"), S)
}

type Statistic struct {
	ip     string
	speed  float64
	passed int64
}

type SpeedBar struct {
	last  time.Time
	usage int64
	ip    string
	speed float64
	bar   *widgets.Gauge
}

func SendSpeedMsg(ip string, usage int64, speed float64) {
	SpeedMsgCHANNEL <- Statistic{ip: ip, passed: usage, speed: speed}
}

func (speedbar *SpeedBar) updateWithMsg(msg Statistic) (err error) {

	speedbar.last = time.Now()
	// var thisUsage int64 = 0
	speedbar.usage += msg.passed
	nowData := float64(speedbar.usage) / 1024.0
	usageStr := fmt.Sprintf("%f Kb", nowData)
	if nowData > 1024 {
		nowData = nowData / 1024.0
		usageStr = fmt.Sprintf("%f Mb", nowData)
	}

	if nowData > 1024 {
		nowData = nowData / 1024.0
		usageStr = fmt.Sprintf("%f Gb", nowData)
	}

	// speedKbt := float64(thisUsage) / secs
	// var speedKb int64 = 0
	// if speedKbt < 1 {
	// 	speedKb = 1
	// } else {
	// 	speedKb = int64(speedKbt)
	// }

	// speedbar.bar.SetTotal(speedbar.usage, false)
	// usageStr = fmt.Sprintf("%f b", float64(speedbar.usage))
	speedF := msg.speed
	speedStr := fmt.Sprintf("%f b/s", speedF)

	if speedF > 1024 {
		speedF = speedF / 1024.0
		speedStr = fmt.Sprintf("%f Kb/s", speedF/1024.0)
		speedbar.bar.Percent = 59
	} else {
		speedbar.bar.Percent = int(float32(speedF) / 1024.0 * 60)
	}
	usageStr = speedStr + "/ " + usageStr

	speedbar.bar.Label = usageStr
	// speedbar.bar.Title = parts[1]
	return

}

func SpeedShow() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	var Bars = make(map[string]*SpeedBar)
	start, width, height, interval := 0, 60, 3, 1
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Close()
				os.Exit(0)
			}
		case msg := <-SpeedMsgCHANNEL:

			ip := msg.ip
			now_bars := len(Bars)
			if bar, ok := Bars[strings.TrimSpace(ip)]; ok {
				bar.updateWithMsg(msg)
				ui.Render(bar.bar)
			} else {
				g3 := widgets.NewGauge()
				g3.Title = ip + " speed show"
				g3.SetRect(start, 1+now_bars*(interval+height), width, (now_bars+1)*(interval+height))
				g3.Percent = width
				g3.Label = fmt.Sprintf("%v%% (100MBs free)", g3.Percent)
				g3.BarColor = ui.ColorGreen
				g3.LabelStyle = ui.NewStyle(ui.ColorYellow)
				g3.TitleStyle.Fg = ui.ColorMagenta
				g3.BorderStyle.Fg = ui.ColorWhite

				Bars[strings.TrimSpace(ip)] = &SpeedBar{
					last:  time.Now(),
					ip:    FGCOLORS[0](ip),
					speed: 0.0,
					bar:   g3,
					usage: int64(msg.passed),
				}

				ui.Render(g3)
			}

		default:
			continue
		}

	}

}
