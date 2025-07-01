package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"unicode"
)

var Key256 string = "603deb1015ca71be2b73aef0857d77811f352c073b6108d72d9810a30914dff4"

// null 패딩 함수
func NullPadding(data []byte, blockSize int) ([]byte, int) {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	return append(data, padText...), padding
}

func Encrypt(encType string, blockType string, paddingType string, key string, iv string, data []byte) ([]byte, error) {
	// ECB 는 단순 암호화. 같은 평문 블록은 같은 암호문 블록으로 변환됨. iv 필요없음.
	// CBC 는 암호문 블록이 이전 블록의 평문 블록과 연관되어 있어서 같은 평문 블록이라도 다른 암호문 블록으로 변환됨. iv 필요함.
	var block cipher.Block
	var blockSize int
	var err error
	var byteKey []byte
	var byteiv []byte
	var paddingPlane []byte
	var cipherData []byte
	var encodeBuff []byte

	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	// 1. (string->hex) key 변환
	byteKey = initKey(key)

	padType := strings.Split(paddingType, ",")
	if encType == "AES" || encType == "aes" {
		blockSize = aes.BlockSize
	} else if padType[0] == "NULL" || padType[0] == "null" {
		blockSize = S_Atoi(padType[1])
	} else {
		blockSize = aes.BlockSize
	}

	// 2. padding 버퍼 생성
	var paddingLen int = 0
	var paddingSize int = 0
	switch paddingType {
	case "PKCS5", "pkcs5":
		paddingPlane = pkcs5padding(data, blockSize)
		paddingSize = len(paddingPlane)
	case "PKCS7", "pkcs7":
		paddingPlane = pkcs7padding(data, blockSize)
		paddingSize = len(paddingPlane)
	case "ISO7816-4", "iso7816-4":
		paddingPlane = ISO7816_4Padding(data, blockSize)
		paddingSize = len(paddingPlane)
	default:
		if len(paddingType) == 1 { // paddingType이 1글자일 경우
			paddingPlane = append(data, bytes.Repeat([]byte{paddingType[0]}, blockSize-(len(data)%blockSize))...)
			paddingSize = len(paddingPlane)
		} else if len(padType) > 1 { // null padding
			var nullLen int
			paddingPlane, nullLen = NullPadding(data, blockSize)
			paddingSize = len(data) + nullLen
		} else if (paddingType == "NULL" || paddingType == "null") && (encType == "AES" || encType == "aes") {
			var nullLen int
			paddingPlane, nullLen = NullPadding(data, aes.BlockSize)
			paddingSize = len(data) + nullLen
		} else { // no padding
			return nil, fmt.Errorf("wrong paddingType [%s]", paddingType)
		}
	}
	if paddingSize%blockSize != 0 {
		paddingLen = blockSize - len(paddingPlane)%blockSize
	}
	cipherData = make([]byte, paddingSize+paddingLen) // 암호화 결과 저장 버퍼

	// 3. 암호화
	switch encType {
	case "AES", "aes": // AES 암호화
		block, err = aes.NewCipher(byteKey)
		if err != nil {
			return nil, err
		}
		bSize := block.BlockSize()
		switch blockType {
		case "ECB", "ecb": // ECB 암호화 (iv 불필요)
			for start := 0; start < len(paddingPlane); start += bSize {
				end := start + bSize
				block.Encrypt(cipherData[start:end], paddingPlane[start:end])
			}
		case "CBC", "cbc": // CBC 암호화 (iv 필요, 입력 값 없을 시 data 기반 생성)
			byteiv, err = initIv(data, iv, bSize)
			if err != nil {
				return nil, err
			}
			cipher.NewCBCEncrypter(block, byteiv).CryptBlocks(cipherData, paddingPlane)
		default: // iv 없이 CBC 암호화 처리
			byteiv = make([]byte, aes.BlockSize) // 빈값으로 초기화
			cipher.NewCBCEncrypter(block, byteiv).CryptBlocks(cipherData, paddingPlane)
		}
	case "SHA256", "sha256": // SHA256 암호화
		hash := sha256.New()
		hash.Write(data)
		cipherData = hash.Sum(nil)
	case "SHA512", "sha512": // SHA512 암호화
		hash := sha512.New()
		hash.Write(data)
		cipherData = hash.Sum(nil)
	default:
		return nil, fmt.Errorf("wrong enc/dec Type [%s]", encType)
	}

	// base64로 인코딩
	encodeBuff = make([]byte, base64.StdEncoding.EncodedLen(len(cipherData)))
	base64.StdEncoding.Encode(encodeBuff, cipherData)
	return encodeBuff, nil
}

// NullUnpadding removes null characters padding from the input
func NullUnpadding(data []byte) []byte {
	return bytes.TrimRight(data, "\x00")
}
func Decrypt(encType string, blockType string, paddingType string, key string, iv string, data []byte) ([]byte, error) {
	var block cipher.Block
	//var blockSize int
	var err error
	var byteKey []byte
	var byteiv []byte
	var cipherData []byte
	var decodeBuff []byte

	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	// 1. (string->hex) key 변환
	byteKey = initKey(key)
	padType := strings.Split(paddingType, ",")

	// base64 디코딩
	decodeBuff = make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(decodeBuff, data)
	if err != nil {
		return nil, err
	}
	decodeBuff = decodeBuff[:n]
	cipherData = make([]byte, len(decodeBuff)) // 복호화 결과 저장 버퍼

	// 2. 복호화
	switch encType {
	case "AES", "aes": // AES 복호화
		block, err = aes.NewCipher(byteKey)
		bSize := block.BlockSize()
		if err != nil {
			return nil, err
		}
		switch blockType {
		case "ECB", "ecb": // ECB 복호화 (iv 불필요)
			for bs, be := 0, block.BlockSize(); bs < len(decodeBuff); bs, be = bs+block.BlockSize(), be+block.BlockSize() {
				block.Decrypt(cipherData[bs:be], decodeBuff[bs:be])
			}
		case "CBC", "cbc": // CBC 복호화 (iv 필요, 입력 값 없을 시 data 기반 생성)
			byteiv, err = initIv(decodeBuff, iv, bSize)
			if err != nil {
				return nil, err
			}
			cipher.NewCBCDecrypter(block, byteiv).CryptBlocks(cipherData, decodeBuff)
		default: // iv 없이 CBC 복호화 처리
			byteiv = make([]byte, bSize) // 빈값으로 초기화
			cipher.NewCBCDecrypter(block, byteiv).CryptBlocks(cipherData, decodeBuff)
		}
	default:
		return nil, fmt.Errorf("wrong enc/dec Type [%s]", encType)
	}

	// 3. padding 제거
	var unpaddedData []byte
	switch paddingType {
	case "PKCS5", "pkcs5":
		unpaddedData = pkcs5unpadding(cipherData)
	case "PKCS7", "pkcs7":
		unpaddedData = pkcs7unpadding(cipherData)
	case "ISO7816-4", "iso7816-4":
		unpaddedData = ISO7816_4Unpadding(cipherData)
	default:
		if len(paddingType) == 1 { // paddingType이 1글자일 경우
			unpaddedData = bytes.TrimRight(cipherData, paddingType)

		} else if len(padType) > 1 { // null padding
			unpaddedData = bytes.TrimRight(cipherData, "\x00")
		} else { // no padding
			unpaddedData = cipherData
		}
	}

	return unpaddedData, nil
}

func initKey(key string) []byte {
	byteKey, err := hex.DecodeString(key)
	if err != nil {
		byteKey = []byte(key)
	}
	return byteKey
}
func initIv(data []byte, iv string, blockSize int) ([]byte, error) {
	var cipherData []byte
	var byteiv []byte
	var err error
	if len(iv) <= 0 {
		cipherData = make([]byte, blockSize+len(data))
		byteiv = cipherData[:blockSize]
		if _, err := io.ReadFull(rand.Reader, byteiv); err != nil {
			return nil, err
		}
	} else {
		byteiv, err = hex.DecodeString(iv)
		if err != nil {
			return nil, err
		}
	}
	return byteiv, nil
}

// pkcs5padding 구현
func pkcs5padding(data []byte, blockSize int) []byte {
	paddingSize := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(paddingSize)}, paddingSize)
	return append(data, padding...)
}

// PKCS5 패딩 제거
func pkcs5unpadding(data []byte) []byte {
	paddingSize := int(data[len(data)-1])
	return data[:len(data)-paddingSize]
}

// PKCS7 패딩
func pkcs7padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// PKCS7 패딩 제거
func pkcs7unpadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

// ISO7816-4 Padding 구현
func ISO7816_4Padding(data []byte, blockSize int) []byte {
	paddingSize := blockSize - len(data)%blockSize
	paddedData := append(data, 0x80)
	padding := bytes.Repeat([]byte{0x00}, paddingSize-1)
	return append(paddedData, padding...)
}

// ISO7816-4 패딩 제거
func ISO7816_4Unpadding(data []byte) []byte {
	data = bytes.TrimRight(data, "\x00") // 0x00 패딩 제거
	if len(data) > 0 && data[len(data)-1] == 0x80 {
		return data[:len(data)-1]
	}
	return data
}

// IsBase64 checks if a string is a valid Base64 encoded string
func IsBase64(s string) bool {
	// Check if the string is empty
	if s == "" {
		return false
	}

	// Check if the string contains only valid Base64 characters
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '+' && r != '/' && r != '=' {
			return false
		}
	}

	// Try to decode the string
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}
