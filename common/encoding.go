package common

import (
	"unicode/utf8"

	"github.com/suapapa/go_hangul/encoding/cp949"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

func Convert_BytetoStringForEUCKR(str []byte) string { // 성공 실패 여부 상관없이 문자열 반환
	retnBuff, _, _ := ConvertEncoding(str, "UTF-8")
	return string(retnBuff)
}

func ConvertEncoding(str []byte, targetEncoding string) ([]byte, bool, string) {
	// 타겟 인코딩 지정
	var targetDecoder transform.Transformer
	charset := utf8.Valid(str)
	switch targetEncoding {
	case "UTF-8", "UTF_8", "UTF8", "utf-8", "utf_8", "utf8":
		if charset {
			return str, true, "This String is Already UTF-8"
		}
		targetDecoder = korean.EUCKR.NewDecoder()
	case "cp949", "CP949", "windows-1252", "EUC-KR", "EUC_KR", "EUCKR", "euc-kr", "euc_kr", "euckr":
		if !charset {
			return str, true, "This String is Already CP949"
		}
		cp949Bytes, err := cp949.To(str)
		if err != nil {
			return str, false, err.Error()
		}
		return cp949Bytes, true, ""
	default:
		return str, false, "unsupported encoding"
	}

	// 인코딩 변환
	newBytes, _, err := transform.Bytes(targetDecoder, str)
	if err != nil {
		return str, false, err.Error()
	}

	// 변환된 문자열 반환
	return newBytes, true, ""
}

/*
******************************************************************************************

	Function    : Exception
	Description : Panic 시  Stack Trace 기록 및 프로세스 recovery
	Argumet     :
	Return      :

******************************************************************************************
*/

/*
func Exception() {
	//executablePath := fmt.Sprintf("%s/%04d%02d%02d/", Logging.FilePath, time.Now().Year(), time.Now().Month(), time.Now().Day())
	if r := recover(); r != nil {
		Logging.Write(LogERROR, "Exception is thrown %s", r)
		stackTrace := debug.Stack()
		errmnt, errline := PrintErrorLine(r)
		absolutePath := filepath.Dir(Logging.FilePath)
		filePth, err := os.Stat(absolutePath)
		if err != nil {
			if os.IsNotExist(err) {
				Logging.Write(LogERROR, "Not Found Directory.. %s", err)
			} else {
				Logging.Write(LogERROR, "Searching Problem Directory .. %s", err)
			}
			return
		}
		if !filePth.IsDir() {
			Logging.Write(LogERROR, "Not Useful Directory Path..")
		}
		numbering := fmt.Sprintf("%02d", S_Atoi(os.Args[1]))
		file, err := os.OpenFile(absolutePath+"/"+filepath.Base(BIN_NAME)+"Core_"+numbering+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			Logging.Write(LogERROR, "File Open Fail.. %s", err)
			return
		}
		defer file.Close()

		content := stackTrace
		writer := bufio.NewWriter(file)
		currentTime := time.Now()
		formattedTime := currentTime.Format("2006-01-02 15:04:05")
		if errline != -1 {
			_, err = writer.WriteString("\n[" + formattedTime + "] > " + "[Build Date : " + BUILD_DATE + " ]" + "\n" + "Err Func : " + errmnt + " -> Line :" + I_ItoA(errline) + "\n")
			if err != nil {
				Logging.Write(LogERROR, "File Write Fail.. %s", err)
				return
			}
		} else {
			_, err = writer.WriteString("[" + formattedTime + "] > " + "[Build Date : " + BUILD_DATE + " ]" + "\n" + string(content) + "\n")
			if err != nil {
				Logging.Write(LogERROR, "File Write Fail.. %s", err)
				return
			}
		}

		Logging.Fd.WriteString("[" + formattedTime + "] > " + "[Build Date : " + BUILD_DATE + " ]" + "\n" + string(content) + "\n")

		err = writer.Flush()
		if err != nil {
			Logging.Write(LogERROR, "Flush Buff Fail.. %s", err)
			return
		}
	}

}

*/
