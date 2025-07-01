package common

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"
)

/*
*******************************************************************************************

	Function    : M_Atoi
	Description : string -> int 변환
	Argumet     : 1. (string) 문자열
	Return      : 1. (int) 변환 정수형
	ETC         :

******************************************************************************************
*/
func S_Atoi(text string) int {
	var ret int
	ret, _ = strconv.Atoi(text)
	return ret
}

/*
*******************************************************************************************

	Function    : B_Atoi
	Description : []byte -> int 변환
	Argumet     : 1. ([]byte) 문자열
	Return      : 1. (int) 변환 정수형
	ETC         :

******************************************************************************************
*/
func B_Atoi(text []byte) int {
	var ret int
	var temp string = string(text)
	ret, _ = strconv.Atoi(temp)
	return ret
}

/*
*******************************************************************************************

	Function    : I_ItoA
	Description : int -> string 변환
	Argumet     : 1. (int) 정수형
	Return      : 1. (string) 변환 문자열
	ETC         :

******************************************************************************************
*/
func I_ItoA(num int) string {
	ret := strconv.Itoa(num)
	return ret
}

/*
******************************************************************************************

	Function    : GetTime6
	Description : 현재 시간 정보 표시 ( HHMMSS )
	Argumet     : 1.
	Return      : 1. (string) 현재 시간
	ETC         :

******************************************************************************************
*/
func GetTime6() string {
	now := time.Now()
	dateStr := fmt.Sprintf("%02d%02d%02d", now.Hour(), now.Minute(), now.Second())

	return dateStr
}

/*
******************************************************************************************

	Function    : GetTime9
	Description : 현재 시간 정보 표시 ( HH:MM:SS.mls )
	Argumet     : 1.
	Return      : 1. (string) 현재 시간
	ETC         :

******************************************************************************************
*/
func GetTime9() string {
	now := time.Now()
	millisecond := now.Nanosecond() / 1e6 // Convert nanoseconds to milliseconds
	dateStr := fmt.Sprintf("%02d:%02d:%02d.%03d", now.Hour(), now.Minute(), now.Second(), millisecond)

	return dateStr
}

/*
******************************************************************************************

	Function    : GetTimeRequired
	Description : 소요 시간 계산
	Argumet     : 1. (string) 시작 시간
	            : 2. (string) 종료 시간
	Return      : 1. (int) 소요 시간
		        : 2. (error) 에러
	ETC         :

******************************************************************************************
*/
func GetTimeRequired(t9_1, t9_2 string) int {
	// string 형식을 time.Time 형식으로 변환
	layout := "15:04:05.000"

	t1, err1 := time.Parse(layout, t9_1)
	if err1 != nil {
		return -1
	}
	t2, err2 := time.Parse(layout, t9_2)
	if err2 != nil {
		return -2
	}

	// 시간 차이 계산
	duration := t2.Sub(t1)
	return int(duration.Seconds())
}

/*
******************************************************************************************

	Function    : GetDateTime8
	Description : 현재 날짜 정보 표시 ( YYYYMMDD )
	Argumet     : 1.
	Return      : 1. (string) 현재 날짜
	ETC         :

******************************************************************************************
*/
func GetDateTime8() string {
	now := time.Now()
	dateStr := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())

	return dateStr
}

/*
******************************************************************************************

	Function    : GetDateTime14
	Description : 현재 날짜 정보 표시 ( YYYYMMDDhhmmss )
	Argumet     : 1.
	Return      : 1. (string) 현재 날짜
	ETC         :

******************************************************************************************
*/
func GetDateTime14() string {
	now := time.Now()
	dateStr := fmt.Sprintf("%04d%02d%02d%02d%02d%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	return dateStr
}

/*
******************************************************************************************

	Function    : GetDateTime16
	Description : 현재 날짜 정보 표시 ( YYYYMMDDhhmmssms )
	Argumet     : 1.
	Return      : 1. (string) 현재 날짜
	ETC         :

******************************************************************************************
*/
func GetDateTime16() string {
	now := time.Now()
	millisecond := now.Nanosecond() / 1e6 // Convert nanoseconds to milliseconds

	dateStr := fmt.Sprintf("%04d%02d%02d%02d%02d%02d%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), millisecond)

	return dateStr
}

/*
******************************************************************************************

	Function    : GetDateTime17
	Description : 현재 날짜 정보 표시 ( YYYYMMDDhhmmssmls )
	Argumet     : 1.
	Return      : 1. (string) 현재 날짜
	ETC         :

******************************************************************************************
*/
func GetDateTime17() string {
	now := time.Now()
	millisecond := now.Nanosecond() / 1e6 // Convert nanoseconds to milliseconds

	dateStr := fmt.Sprintf("%04d%02d%02d%02d%02d%02d%03d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), millisecond)

	return dateStr
}

/*
******************************************************************************************

	Function    : CheckKeyValue
	Description : 다수 문자열에서([]string) delimiter 를 구분하여 key, value 분리
	Argumet     : 1. ([]string) 문자열
	            : 2. (string) key
	            : 3. (string) delimiter
	Return      : 1. ([]string) key, value
	ETC         :

******************************************************************************************
*/
func CheckKeyValue(nStr []string, key string, dil string) []string {
	var keyValue []string
	for _, str := range nStr {
		if strings.Contains(str, key+dil) {
			parts := strings.SplitN(str, dil, 2)
			if len(parts) == 2 {
				keyValue = append(keyValue, strings.TrimSpace(parts[0]))
				keyValue = append(keyValue, strings.TrimSpace(parts[1]))
			}
		}
	}
	return keyValue
}

/*
******************************************************************************************

	Function    : CheckFileEncoding
	Description : 파일 인코딩 형식 확인
	Argumet     : 1. (string) 파일명
	Return      : 1. (string) 인코딩 형식
	            : 2. (error) 에러
	ETC         :

******************************************************************************************
*/
func CheckFileEncoding(filename string) (string, error) {
	cmd := exec.Command("file", "-i", filename)
	//cmd := exec.Command("file", filename)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	oStrs := strings.Split(strings.TrimSpace(string(output)), " ")
	parsedData := CheckKeyValue(oStrs, "charset", "=")

	return parsedData[1], nil
}

/*
*******************************************************************************************

	Function    : StringUpper
	Description : 알파벳 String 변수를 대문자로 변경
	Argumet     : 1. (string) 변경 할 문자
	Return      : 1. (string) 변경 된 문자
	ETC         :

******************************************************************************************
*/
func StringUpper(input string) string {
	result := ""
	for _, char := range input {
		// 소문자인 경우 대문자로 변환
		if unicode.IsLower(char) {
			result += string(unicode.ToUpper(char))
		} else {
			result += string(char)
		}
	}

	return result
}

/*
*******************************************************************************************

	Function    : SymbolSplit
	Description : 한쌍의 특수 기호 내부 데이터 추출
	Argumet     : 1. ([]byte) 검색 할 슬라이스
	              2. (byte) 추출 할 구분 값
	Return      : 1. ([]byte) 추출 한 문자열
	ETC         :

******************************************************************************************
*/
func SymbolSplit(data []byte, ch byte) []string {
	var result []string
	var field []byte
	var inQuote bool

	for _, b := range data {
		if b == ' ' && !inQuote {
			if len(field) > 0 {
				result = append(result, string(field))
				field = nil
			}
		} else if b == ch {
			inQuote = !inQuote
		} else {
			field = append(field, b)
		}
	}

	if len(field) > 0 {
		result = append(result, string(field))
	}

	return result
}

func Json2String(v any) string {
	retn, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	} else {
		return string(retn)
	}
}
