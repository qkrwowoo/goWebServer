package common

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

func SockConnect(addr string, timeout time.Duration) (*net.TCPConn, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	TcpConn, _ := conn.(*net.TCPConn)
	TcpConn.SetLinger(0)
	TcpConn.SetReadBuffer(10 * 1024)
	TcpConn.SetWriteBuffer(10 * 1024)
	TcpConn.SetNoDelay(true)
	return TcpConn, nil
}

func SockSend(TcpConn *net.TCPConn, data []byte, dummy []string) ([]byte, error) {
	var dum []byte
	for _, v := range dummy {
		dum = append(dum, []byte(v)...)
	}
	data = append(dum, data...)

	_, err := TcpConn.Write(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SockReader(TcpConn *net.TCPConn, timeout time.Duration) (int, []byte, error) {
	var err error
	var retn int
	var readlen int     // 처음으로 읽은 길이
	var rlen int = 8192 // 읽을 길이
	var failCount int = 0
	var firstBuff []byte // 처음 읽은 데이터
	var tempBuff []byte  // 이후 읽은 데이터
	var readBuff []byte
	// READ
	firstBuff = make([]byte, rlen) // 읽을 길이 세팅

	now := time.Now()
	TcpConn.SetReadDeadline(now.Add(timeout))
	readlen, err = TcpConn.Read(firstBuff)
	if err != nil {
		if err == io.EOF {
			// close session
			return -1, nil, err
		} else if opErr, ok := err.(*net.OpError); ok {
			if opErr.Timeout() {
				// timeout
				return 0, nil, err
			} else {
				// other error
				return -3, nil, err
			}
		} else {
			// other error
			return -4, nil, err
		}
	} else if readlen <= 0 {
		errStr := fmt.Sprintf("socket read length is [%d]", readlen)
		return -5, nil, errors.New(errStr)
	}

	readBuff = append(readBuff, firstBuff[:readlen]...)
	for {
		tempBuff = make([]byte, rlen)
		TcpConn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		retn, err = TcpConn.Read(tempBuff)
		if err == io.EOF { // 세션 닫힘
			readlen += retn
			readBuff = append(readBuff, tempBuff[:retn]...)
			break
		}
		if err != nil || retn <= 0 {
			time.Sleep(10 * time.Millisecond)
			failCount++
			if failCount >= 10 { // 사실 여기가 성공부분이라고 볼 수 있음
				return readlen, readBuff, nil
			} else {
				continue
			}
		}
		readlen += retn // 읽은 길이 누적
		readBuff = append(readBuff, tempBuff[:retn]...)
	}

	return readlen, readBuff, nil
}
