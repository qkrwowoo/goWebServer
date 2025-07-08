package main

import (
	"fmt"
	"grpcWeb/db"
	"grpcWeb/gRPCsrc"
	pb "grpcWeb/proto" // 생성된 pb.go 경로에 맞게 import
	c "local/common"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pborman/getopt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var b_version *bool
var s_port *string
var s_ini *string
var end_Signal chan os.Signal

func init() {
	b_version = getopt.BoolLong("version", 'v', "Print Version")
	s_port = getopt.StringLong("port", 'p', "", "Listen Port")
	s_ini = getopt.StringLong("file", 'f', "", "Configure File")
	getopt.Parse()

	end_Signal = make(chan os.Signal, 2)
	signal.Notify(end_Signal, syscall.SIGTERM)
}

// MAIN  -------------------------------------------------------
func main() {
	checkArg()
	time.Sleep(500 * time.Millisecond) // DB접속 완료까지 잠깐 대기

	c.Logging.Write(c.LogALL, "=================================================")
	c.Logging.Write(c.LogALL, "\t [%s] START [%s]", os.Args[0], *s_port)
	c.Logging.Write(c.LogALL, "=================================================")
	lis, err := net.Listen("tcp", *s_port)
	if err != nil {
		clearMem()
		log.Fatal("Failed to connect to gRPC server:", err)
		return
	}

	// gRPC Server 기동
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(gRPCsrc.UnaryInterceptor)) // interceptor 등록
	pb.RegisterUserServiceServer(grpcServer, &gRPCsrc.UserServer{})
	reflection.Register(grpcServer)
	c.Logging.Write(c.LogALL, "\t ### Open gRPC Server")

	var wg sync.WaitGroup
	wg.Add(1)
	if *s_ini != "" {
		go check_Config(&wg)
	} else {
		wg.Done()
	}
	if err := grpcServer.Serve(lis); err != nil {
		clearMem()
	}
	wg.Wait()
}

/*
*******************************************************************************************
  - function	: checkArg
  - Description	: 프로세스 초기 입력 검사.
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         : 환경파일 옵션 존재시 환경파일 조회

*******************************************************************************************
*/
func checkArg() {
	var wg sync.WaitGroup
	wg.Add(1)
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
		fmt.Printf(" grpcWeb Version [%s] --build with go [%s]\n", version, goVersion)
		fmt.Printf("-------------------------------------------------------------------------------\n")
		clearMem()
		os.Exit(0) // 버전 확인 후 종료
	}

	if *s_ini != "" {
		err := c.Load_Config(*s_ini)
		if err != nil {
			fmt.Printf(" Load Config Failed [%s]\n", *s_ini)
			clearMem()
			os.Exit(0)
		}
		init_config()
		db.RDB.LoadConfig()
		db.REDIS.LoadConfig()
		*s_port = ":" + c.CFG["COMMON"]["PORT"].(string)
	} else {
		db.RDB.Default("mysql")
		db.REDIS.Default()
	}

	if *s_port == "" {
		*s_port = ":50051"
	} else if (*s_port)[0] != ':' {
		*s_port = ":" + *s_port
	}
	wg.Done()
}

/*
*******************************************************************************************
  - function	: init_config
  - Description	: 환경파일 조회
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         :

*******************************************************************************************
*/
func init_config() {
	var encyn bool
	if strings.ToLower(c.CFG["LOG"]["ENCYN"].(string)) == "y" {
		encyn = true
	} else {
		encyn = false
	}

	if c.CFG["LOG"]["DIR"].(string) == "" || c.CFG["LOG"]["NAME"].(string) == "" {
		fmt.Println("Not Found Log Section in configure file", c.CFG["LOG"])
		clearMem()
		os.Exit(0)
	}
	c.Logging.InitLog(c.TypeL,
		c.CFG["LOG"]["DIR"].(string),
		c.CFG["LOG"]["NAME"].(string),
		encyn,
		c.CFG["LOG"]["LEVEL"].(string))
	c.Logging.Print_Config(c.CFG)
}

/*
*******************************************************************************************
  - function	: check_Config
  - Description	: 환경파일 변경 감지 확인 후 리로드
  - Argument	: [ (*sync.WaitGroup) 고루틴 작업 대기 ]
  - Return		: [ ]
  - Etc         : 환경파일 옵션 존재시 환경파일 읽기

*******************************************************************************************
*/
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

/*
*******************************************************************************************
  - function	: clearMem
  - Description	: 메모리 자원 동적 해제
  - Argument	: [ ]
  - Return		: [ ]
  - Etc         : 가비지컬렉션 있기에 하지않아도 되지만 명확하게 해제

*******************************************************************************************
*/
func clearMem() {
	db.RDB.AllClose()
	db.REDIS.AllClose()
	end_Signal <- syscall.SIGTERM
}
