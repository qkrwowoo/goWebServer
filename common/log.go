package common

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// 조합하려면 더하기 (ex DEBUG+TRACE = 3, DEBUG+TRACE+ERROR = 11, ALL = ff)
const LogDEBUG byte = 0x01  // 1
const LogTRACE byte = 0x02  // 2
const LogERROR byte = 0x04  // 4
const LogWARN byte = 0x08   // 8
const LogSTDOUT byte = 0x10 // 16
const LogLv5 byte = 0x20    // 32
const LogLv6 byte = 0x40    // 64
const LogStream byte = 0x80 // 128
const LogALL byte = 0xff    // 255

const SEND_SUCC byte = 0x01 // 1
const SEND_FAIL byte = 0x02 // 2
const RECV_SUCC byte = 0x04 // 4
const RECV_FAIL byte = 0x08 // 8

const TypeL byte = 0x01 // 1 (로그)
const TypeS byte = 0x02 // 2 (성공)
const TypeF byte = 0x04 // 4 (실패)
const TypeC byte = 0x08 // 8 (완료)

type LogFile struct {
	LogType     byte // 로그 타입
	LogLevel    byte // MODE
	LogLevels   string
	EncType     string   // 암호화 타입
	EncKey      string   // 암호화 키
	DefaultPath string   // 기본 경로
	DefaultName string   // 기본 이름
	FileDate    string   // 파일 생성 일자 (디렉토리 변경용도)
	FilePath    string   // 디렉토리 경로
	FileName    string   // 파일 이름
	FullName    string   // 경로/파일
	FileTime    int      // 파일 접근 시간
	FileIdx     int      // Adapter Index
	FileType    bool     // 파일 타입
	MaskYN      bool     // Mask
	Fd          *os.File // FD
}

var Logging LogFile
var FailLog LogFile
var SuccLog LogFile
var DoneLog LogFile

func init() {
}

func (L *LogFile) InitLog(LogType byte, dir string, name string, encYN bool, levels string) error {
	if L.Fd != nil {
		L.Fd.Close()
		L.Fd = nil
	}

	L.LogLevels = levels
	l := strings.Split(levels, ",")
	for _, level := range l {
		switch level {
		case "DEBUG":
			L.LogLevel |= LogDEBUG
		case "TRACE":
			L.LogLevel |= LogTRACE
		case "ERROR":
			L.LogLevel |= LogERROR
		case "WARN":
			L.LogLevel |= LogWARN
		case "STDOUT":
			L.LogLevel |= LogSTDOUT
		}
	}

	L.DefaultName = name
	L.DefaultPath = dir
	L.LogType = LogType
	L.FileDate = GetDateTime8()
	switch L.LogType {
	case TypeL:
		L.FilePath = dir + "/" + L.FileDate
		L.FileName = name + "_" + L.FileDate + fmt.Sprintf("_%02d", time.Now().Hour()) + ".log"
	case TypeS:
		L.FilePath = dir + "/" + L.FileDate + "/succ"
		L.FileName = name + "_" + L.FileDate + fmt.Sprintf("_%02d", time.Now().Hour()) + ".succ"
	case TypeF:
		L.FilePath = dir + "/" + L.FileDate + "/fail"
		L.FileName = name + "_" + L.FileDate + fmt.Sprintf("_%02d", time.Now().Hour()) + ".fail"
	case TypeC:
		L.FilePath = dir + "/" + L.FileDate + "/done"
		L.FileName = name + "_" + L.FileDate + fmt.Sprintf("_%02d", time.Now().Hour()) + ".done"
	}
	L.FullName = L.FilePath + "/" + L.FileName
	L.MaskYN = encYN
	L.FileTime = 99
	L.FileIdx = 0

	// 디렉토리 생성
	if _, err := os.Stat(string(L.FilePath)); os.IsNotExist(err) {
		err = os.Mkdir(string(L.FilePath), os.FileMode(0755))
		if err != nil {
			//fmt.Printf("\n  \033[41mDirectory access denied. Check your permissions [%s] [%s]\033[0m \n", L.FilePath, err)
			//fmt.Printf("\n  \033[41m dir[%s/%s]\033[0m \n", dir, L.FilePath)
			return err
		}
	}
	return nil
}

/*
*******************************************************************************************

	Function    : Logging.Write
	Description : tbalog 로그 기록 :: TBAdapter 흐름 로그
	Argumet     : 1. (byte) Log 레벨
	            : 2. (string) 로그 포맷
	            : 3. (interface) 인자 값
	Return      : 1. X
	Etc         : %. : 로그함축화(...)

******************************************************************************************
*/
func (L *LogFile) Write(lv byte, format string, data ...interface{}) {
	if err := L.Check_filetime(); err != nil {
		return
	}
	var tempBuff []byte = make([]byte, 0)
	var buffer bytes.Buffer
	var argIndex int

	if lv != LogStream {
		// 로그 기록 X
		if L.LogLevel == 0 {
			return
		}

		if lv&L.LogLevel <= 0 {
			return
		}

		temp := fmt.Sprintf("[%s] > ", GetTime9())
		tempBuff = append(tempBuff, []byte(temp)...)

		switch lv {
		case LogDEBUG:
			tempBuff = append(tempBuff, []byte("[DEBUG] > ")...)
		case LogTRACE:
			tempBuff = append(tempBuff, []byte("[TRACE] > ")...)
		case LogERROR:
			tempBuff = append(tempBuff, []byte("\033[41m[ERROR]\033[0m > ")...)
		case LogWARN:
			tempBuff = append(tempBuff, []byte("\033[43m[WARN ]\033[0m > ")...)
		case LogALL:
			tempBuff = append(tempBuff, []byte("[ ALL ] > ")...)
		}
	}

	//slev := fmt.Sprintf("[LV-%d] > ", L.LogLevel)
	//tempBuff = append(tempBuff, []byte(slev)...)
	for i := 0; i < len(format); i++ {
		if format[i] == '%' {
			argIndex++
			if argIndex > len(data) {
				break
			}
			switch format[i+1] {
			case 0:
			case '.':
				if L.MaskYN {
					fmt.Fprintf(&buffer, "...")
				} else {
					fmt.Fprintf(&buffer, "%s", data[argIndex-1])
				}
			case '0':
				var temp string
				switch format[i+3] {
				default:
					temp = fmt.Sprintf("%0*d", S_Atoi(format[i+2:i+3]), data[argIndex-1])
				case 'd', 'i':
					temp = fmt.Sprintf("%0*d", S_Atoi(format[i+2:i+3]), data[argIndex-1])
				}
				i += 2
				fmt.Fprintf(&buffer, "%s", temp)
			case 'd', 'i':
				fmt.Fprintf(&buffer, "%d", data[argIndex-1])
			case 'u':
				fmt.Fprintf(&buffer, "%d", uint(data[argIndex-1].(int)))
			case 'o':
				fmt.Fprintf(&buffer, "%o", data[argIndex-1])
			case 'x':
				fmt.Fprintf(&buffer, "%x", data[argIndex-1])
			case 'X':
				fmt.Fprintf(&buffer, "%X", data[argIndex-1])
			case 'f', 'F':
				fmt.Fprintf(&buffer, "%f", data[argIndex-1])
			case 'e', 'E':
				fmt.Fprintf(&buffer, "%e", data[argIndex-1])
			case 'g', 'G':
				fmt.Fprintf(&buffer, "%g", data[argIndex-1])
			case 's':
				fmt.Fprintf(&buffer, "%s", data[argIndex-1])
			case 'c':
				fmt.Fprintf(&buffer, "%c", data[argIndex-1])
			case 'v':
				fmt.Fprintf(&buffer, "%v", data[argIndex-1])
			case 'p':
				fmt.Fprintf(&buffer, "%p", data[argIndex-1])
			case '%':
				fmt.Fprintf(&buffer, "%%")
			default:
				//panic("unknown format specifier")
			}
			i++
		} else {
			buffer.WriteByte(format[i])
		}
	}

	tempBuff = append(tempBuff, buffer.String()...)
	tempBuff = append(tempBuff, '\n')

	L.Fd.WriteString(string(tempBuff))
	if L.LogLevel&LogSTDOUT == LogSTDOUT {
		fmt.Print(string(tempBuff))
	}
}

func (L *LogFile) Print_Config(cfgMap map[string]map[string]interface{}) {
	for k, v := range cfgMap {
		L.Write(LogDEBUG, " [ %s ] =================================================", k)
		for key, val := range v {
			switch value := val.(type) {
			case int:
				L.Write(LogDEBUG, "\t%s = [%d]", key, value)
			case string:
				L.Write(LogDEBUG, "\t%s = [%s]", key, value)
			default:
				L.Write(LogDEBUG, "\t%s = [%s]", key, value)
			}
		}
	}
}

/*
*******************************************************************************************

	Function    : Check_filetime
	Description : 시간을 체크하여 로그이름 및 경로 변경
	Argumet     : 1.
	Return      : 1. X

******************************************************************************************
*/
func (L *LogFile) Check_filetime() error {
	var err error

	if _, err := os.Stat(L.FullName); !os.IsNotExist(err) && L.FileTime == time.Now().Hour() {
		return err
	}
	if L.Fd != nil {
		L.Fd.Close()
		L.Fd = nil
	}
	L.InitLog(L.LogType, L.DefaultPath, L.DefaultName, L.MaskYN, L.LogLevels)
	L.Fd, err = OpenLogFile(L.FilePath, L.FullName)
	if err != nil {
		return err
	}
	L.FileTime = time.Now().Hour()
	return nil
}

func PrintErrorLine(r interface{}) (string, int) {
	pc, _, line, ok := runtime.Caller(4)
	if !ok {
		Logging.Write(LogERROR, "Failed to get caller information")
		return "", -1
	}
	fn := runtime.FuncForPC(pc)
	fnName := fn.Name()
	return fnName, line
}

func OpenLogFile(path, FullName string) (*os.File, error) {
	Fd, err := os.OpenFile(FullName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0666))
	if err != nil {
		// 일자 디렉토리 있는지 확인 및 생성
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.Mkdir(path, os.FileMode(0755))
			if err != nil {
				return nil, err
			}
		}
		Fd, err = os.OpenFile(FullName, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0666))
		if err != nil {
			return nil, err
		}
	}
	return Fd, nil
}
