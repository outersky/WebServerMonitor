package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"crypto/tls"
	"time"
	"path/filepath"
	"os/exec"
	"fmt"
	"runtime"
)

var ini *Ini

var dir string     // work dir , where exe and config files are stored
var cmdFilePath string  // cmd file which will be executed while max error times occurs
var logFilePath string  // log file path
var count = 0      // error count

var logFile *os.File

var LF string

type InitError struct {
	Message string
	Err  error
}

func (e *InitError) Error() string {
	m := e.Message
	if e.Err!=nil {
		m += " " + e.Err.Error()
	}
	return m
}

func doInit() error {
	if runtime.GOOS == "windows" {
		LF = "\r\n"
	}else{
		LF = "\n"
	}

//	println(runtime.GOOS)
/*
	for _,s := range os.Environ() {
		println(s)
	}

	println(os.Environ())
*/
	exeFile := os.Args[0]
	dir = filepath.Dir(exeFile)
	cfgFile := filepath.Join(dir,"app.json")
	logFilePath = filepath.Join(dir,"log.txt")

	file,err := openLogFile(logFilePath)
	if err!= nil {
		println(err)
		return err
	}
	logFile = file

	info("exeFile: %s",exeFile)

	// defer logFile.Close()

	info("cfgFilePath: %s",cfgFile)
	ini,err = openConfig(cfgFile)
	if err!=nil {
		return err
	}
	cmdFilePath = filepath.Join(dir,ini.Cmd)
	info("url: %s",ini.Url)
	info("scanIntervals: %d",ini.ScanIntervals)
	info("maxErrorCount: %d",ini.MaxErrorCount)
	info("cmdFilePath: %s",cmdFilePath)

	info("================")

	if ini.ScanIntervals<=0 {
		info("ScanIntervals can NOT be zero .")
		return &InitError{
			Message: "ScanIntervals can NOT be zero .",
		}
	}
	return nil
	// startTicker()
}

type Ini struct {
	Url string `json:"url,omitempty"`
	ScanIntervals int `json:"scanIntervals,omitempty"`
	MaxErrorCount int `json:"maxErrorCount,omitempty"`
	Cmd string `json:"cmd,omitempty"`
}

func openConfig(filename string) (ini *Ini,err error) {
	ini = &Ini{}
	bytes,err := ioutil.ReadFile(filename)
	if err!= nil {
		return
	}
	json.Unmarshal(bytes,ini)
	return
}

func openUrl(url string) (content []byte, err error){
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err!=nil {
		return nil,err
	}
	return ioutil.ReadAll( resp.Body )
}

func doWork() {
	log.Info("I'm Running!")
	ticker := time.NewTicker(time.Duration(ini.ScanIntervals) * time.Second)
	for {
		select {
		case <-ticker.C:
			monitor()
		case <-exit:
			ticker.Stop()
			return
		}
	}
}

/*
func startTicker(){
	c := time.Tick( time.Duration(ini.ScanIntervals) * time.Second  )
	for now := range c {
		monitor(now)
	}
}
*/

func monitor(/*now time.Time*/){
//	info("open url: %s", ini.Url)
	_,err := openUrl(ini.Url)

	if err == nil {
		return
	}

	count ++
	info("error times : %d", count)
	if count < ini.MaxErrorCount {
		return
	}

	restart()
	count = 0

}

func restart(){
	info("restart by : %s", cmdFilePath)
	out, err := execute(cmdFilePath)
	if err != nil {
		panic(err)
	}
	info("The output data is %s", out)
}

func execute(cmd string) ([]byte, error) {
	if runtime.GOOS == "windows" {
		return exec.Command(cmd).Output()
	}else{
		return exec.Command("/bin/bash",cmd).Output()
	}
}

func openLogFile(filename string) (file *os.File,err error) {
	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	return
}

func info(format string, a ...interface{}) {
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("%s : ", now)
	fmt.Printf(format + LF, a...)

	fmt.Fprintf(logFile,"%s : ", now)
	fmt.Fprintf(logFile, format + LF, a...)
}

