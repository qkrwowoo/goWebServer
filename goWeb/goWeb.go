package main

import (
	"fmt"
	"goWeb/grpcHandler"
	c "local/common"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pborman/getopt"
)

var b_version *bool
var s_ip *string
var s_port *string
var s_listen *string
var s_ini *string
var end_Signal chan os.Signal

var releaseMode bool

func init() {
	b_version = getopt.BoolLong("version", 'v', "", "Print Version")
	s_ip = getopt.StringLong("ip", 'i', "", "Dest gRPC IP Address")
	s_port = getopt.StringLong("port", 'p', "", "Dest gRPC Port")
	s_ini = getopt.StringLong("file", 'f', "", "Configure File")
	s_listen = getopt.StringLong("listen_port", 'l', "", "web listen port")
	getopt.Parse()

	end_Signal = make(chan os.Signal, 2)
	signal.Notify(end_Signal, syscall.SIGTERM)
}

func main() {
	checkArg()

	grpcHandler.IPADDR = *s_ip + ":" + *s_port
	c.Logging.Write(c.LogALL, "=================================================")
	c.Logging.Write(c.LogALL, "\t [%s] START [%s]", os.Args[0], *s_listen)
	c.Logging.Write(c.LogALL, "=================================================")
	r := gin.Default()
	if releaseMode {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := grpcHandler.Open_gRPC_Session(); err != nil {
		log.Fatal("Failed to connect to gRPC server:", err)
		return
	}
	defer grpcHandler.Close_gRPC_Session()
	time.Sleep(500 * time.Millisecond)
	go grpcHandler.Register(r)
	go grpcHandler.Login(r)
	go grpcHandler.UserInfo(r)

	var wg sync.WaitGroup
	wg.Add(1)
	if *s_ini != "" {
		go check_Config(&wg)
	} else {
		wg.Done()
	}
	r.Run(*s_listen)
	c.Logging.Write(c.LogALL, "\t Open Web Server")
	wg.Wait()
}

func checkArg() string {
	if *b_version {
		var goVersion string

		info, ok := debug.ReadBuildInfo()
		if ok {
			goVersion = info.GoVersion
		} else {
			goVersion = "unknown"
		}
		version := "v.0.0.0"
		fmt.Println()
		fmt.Printf("-------------------------------------------------------------------------------\n")
		fmt.Printf(" goWebServer Version [%s] --build with go [%s]\n", version, goVersion)
		fmt.Printf("-------------------------------------------------------------------------------\n")
		os.Exit(0) // 버전 확인 후 종료
	}

	if *s_ini != "" {
		err := c.Load_Config(*s_ini)
		if err != nil {
			log.Fatalln("configure file open failed", *s_ini, err)
			os.Exit(0)
		}
		init_config()
	}

	if *s_ip == "" {
		*s_ip = "127.0.0.1"
	}

	if *s_port == "" {
		*s_port = "50051"
	}

	if *s_listen == "" {
		*s_listen = ":8080"
	} else if (*s_listen)[0] != ':' {
		*s_listen = ":" + *s_listen
	}

	return ""
}

func init_config() {
	if mode := c.CFG["COMMON"]["PORT"].(string); mode == "DEV" {
		releaseMode = false
	} else {
		releaseMode = true
	}

	if lport := c.CFG["COMMON"]["PORT"].(string); lport[0] != ':' {
		*s_listen = ":" + lport
	} else {
		*s_listen = lport
	}

	*s_ip = c.CFG["gRPC"]["IP"].(string)
	*s_port = c.CFG["gRPC"]["PORT"].(string)

	var encyn bool
	if strings.ToLower(c.CFG["LOG"]["ENCYN"].(string)) == "y" {
		encyn = true
	} else {
		encyn = false
	}

	c.Logging.InitLog(c.TypeL,
		c.CFG["LOG"]["DIR"].(string),
		c.CFG["LOG"]["NAME"].(string),
		encyn,
		c.CFG["LOG"]["LEVEL"].(string))
	c.Logging.Print_Config(c.CFG)
}

func check_Config(wg *sync.WaitGroup) {
	ConfigInfo, _ := os.Stat(*s_ini)
	lastCheckTime := ConfigInfo.ModTime()

	readTimer := time.NewTimer(10 * time.Second)
	for {
		select {
		case <-end_Signal:
			wg.Done()
			return
		case <-readTimer.C:
			ConfigInfo, err := os.Stat(*s_ini)
			if os.IsNotExist(err) || ConfigInfo.ModTime() != lastCheckTime { // 변경 감지
				if ConfigInfo.ModTime() != lastCheckTime {
					c.Load_Config(*s_ini)
					lastCheckTime = ConfigInfo.ModTime()
					c.Logging.Print_Config(c.CFG)
				}
			}
			readTimer.Reset(10 * time.Second)
		}
		time.Sleep(1 * time.Second)
	}
}
