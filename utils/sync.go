package utils

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/howeyc/gopass"
)

var (
	Tmp    = os.TempDir()
	dst    = filepath.Join(Tmp, "tmp.list")
	cdst   = filepath.Join(Tmp, "config.en")
	cdst_z = filepath.Join(Tmp, "config.temp.en.zip")
	cdst_k = filepath.Join(Tmp, "Kcpconfig")
)

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	if _, err := os.Stat(filepath); err == nil {
		os.Remove(filepath)
	}
	startAt := time.Now()
	defer ColorL("Download config "+url+" used:", time.Now().Sub(startAt), "=>", filepath)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.

func zipit(source, target string) error {
	if _, err := os.Stat(target); err == nil {
		os.Remove(target)
	}
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {
	startAt := time.Now()
	defer ColorL("parse config used:", time.Now().Sub(startAt))
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		log.Println("open err:", src)
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

// will donwload file from your git://github.com/${name}/pro/config.en  name : name/pro
// password will decrypt config
func Credient(username, password string) (routeDir string, err error) {
	startAt := time.Now()
	defer ColorL("load config from Network used:", time.Now().Sub(startAt))
	if !PathExists(cdst) {
		if username == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter Username(git_username/gitproject_name): ")
			username, _ = reader.ReadString('\n')
			username = strings.TrimSpace(username)
		}
	} else {
		ColorL("found crypted config file", "config.en")
	}
	if password == "" {
		fmt.Print("Enter Password: ")
		passwd, err := gopass.GetPasswd()

		password = strings.TrimSpace(string(passwd))
		if err != nil {
			os.Exit(0)
		}
	}
	// fmt.Print("Enter Method[aes/tea/xtea]: ")
	// reader = bufio.NewReader(os.Stdin)
	// method, _ := reader.ReadString('\n')

	method := "aes-256-cfb"
	if method == "" || password == "" {
		ColorL("must specify , user , pwd, method[aes/tea/xtea]")
		os.Exit(0)
	}

	config := Config{Server: username, Method: method, Password: password}
	en := config.GeneratePassword()
	if strings.Contains(username, " ") {
		us := []string{}
		for _, v := range strings.Split(username, " ") {
			if v != "" {
				us = append(us, v)
			}
		}
		username = strings.Join(us, "/")
	}
	if !PathExists(cdst) {
		pts := strings.SplitN(username, ":", 2)
		username = pts[0]
		choiceConfigKey := "default"
		if len(pts) > 1 {
			choiceConfigKey = pts[1]
		}
		baseURL := fmt.Sprintf("https://gitee.com/%s/raw/master/", username)

		if strings.Contains(username, "://") {
			parts := strings.SplitN(username, "://", 2)

			head, content := parts[0], parts[1]
			ColorL(head, content)
			switch head {
			case "git":
				baseURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/master/", content)
			case "http":
				baseURL = username
			case "https":
				baseURL = username
			}
		}
		indexURL := baseURL + "list"
		if err = DownloadFile(dst, indexURL); err != nil {
			return
		}
		defer os.Remove(dst)
		if db, ierr := os.Open(dst); ierr == nil {
			defer db.Close()

			if d, ierr := ioutil.ReadAll(db); ierr == nil {
				mapJSON := make(map[string]string)
				if err = json.Unmarshal(d, &mapJSON); err == nil {
					if v, ok := mapJSON[choiceConfigKey]; ok {
						url := baseURL + v
						ColorL("sync from :", url)
						if err = DownloadFile(cdst, url); err != nil {
							log.Println("err sync:", err)
							return
						}
					} else {
						ColorL("No such config in :", baseURL)
						os.Exit(0)
					}
				}
			}
		}

	}

	if fb, err := os.Open(cdst); err == nil {
		defer os.Remove(cdst)
		defer fb.Close()
		if datas, err := ioutil.ReadAll(fb); err == nil {
			deData := make([]byte, len(datas))
			en.Decrypt(deData, datas)
			if err = ioutil.WriteFile(cdst_z, deData, 0644); err == nil {
				defer os.Remove(cdst_z)
				if _, err := Unzip(cdst_z, cdst_k); err == nil {
					routeDir = cdst_k
				} else {
					log.Println("unzip error:", err)
				}
			} else {
				log.Println("open config.en error:", err)
			}
		}
	} else {
		log.Fatal(err)
	}

	return
}

func Sync(configRoot string) (file string, err error) {
	startAt := time.Now()
	defer ColorL("gen config.en:", time.Now().Sub(startAt))
	// var files []string
	fmt.Println("Enter Password: ")
	passwd, err := gopass.GetPasswd()
	if err != nil {
		os.Exit(0)
	}
	method := "aes-256-cfb"
	password := strings.TrimSpace(string(passwd))
	if method == "" || password == "" {
		ColorL("must specify , user , pwd, method[aes/tea/xtea]")
		os.Exit(0)
	}
	config := Config{Method: method, Password: password}
	en := config.GeneratePassword()

	// err = filepath.Walk(configRoot, func(path string, info os.FileInfo, err error) error {
	// 	ColorL("pack", path)
	// 	if strings.HasSuffix(path, ".json") {
	// 		files = append(files, path)
	// 	}
	// 	return nil
	// })
	// if PathExists(route) {
	// 	files = append(files, route)
	// }
	if err := zipit(configRoot, "Kcpconfig.zip"); err == nil {
		defer os.Remove("Kcpconfig.zip")
		if fb, err := os.Open("Kcpconfig.zip"); err == nil {
			defer fb.Close()
			if datas, err := ioutil.ReadAll(fb); err == nil {
				deData := make([]byte, len(datas))
				en.Encrypt(deData, datas)
				if err = ioutil.WriteFile("config.en", deData, 0644); err == nil {
					file = "config.en"
					home := filepath.Join(HOME, "Desktop", "config.en")
					os.Rename(file, home)
					file = home
				}
			}
		}

	}
	return
}
