package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	// 连接到本地服务端
	conn, err := net.Dial("tcp", "127.0.0.1:8389")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	// 生成salt
	salt := make([]byte, 32)
	rand.Read(salt)

	// 写入salt
	conn.Write(salt)

	// 生成密钥
	password := "13687401432Fan!"
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)

	// 创建加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("Failed to create cipher: %v\n", err)
		return
	}
	encryptor, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("Failed to create GCM: %v\n", err)
		return
	}

	// 构建目标地址 (google.com:80)
	target := "google.com:80"
	host := "google.com"
	port := 80

	// 构建地址数据
	addrData := make([]byte, 4+len(host))
	addrData[0] = 3 // 域名类型
	addrData[1] = byte(len(host))
	copy(addrData[2:2+len(host)], host)
	binary.BigEndian.PutUint16(addrData[2+len(host):4+len(host)], uint16(port))

	// 生成nonce并加密数据
	nonce := make([]byte, encryptor.NonceSize())
	rand.Read(nonce)

	encryptedData := encryptor.Seal(nil, nonce, addrData, nil)

	// 写入nonce
	conn.Write(nonce)

	// 写入长度
	length := len(encryptedData)
	lengthBuf := []byte{byte(length >> 8), byte(length)}
	conn.Write(lengthBuf)

	// 写入加密数据
	conn.Write(encryptedData)

	fmt.Printf("Sent target: %s\n", target)

	// 发送一些测试数据
	testData := []byte("GET / HTTP/1.1\r\nHost: google.com\r\n\r\n")

	// 加密并发送测试数据
	nonce2 := make([]byte, encryptor.NonceSize())
	rand.Read(nonce2)
	encryptedTestData := encryptor.Seal(nil, nonce2, testData, nil)

	conn.Write(nonce2)
	length2 := len(encryptedTestData)
	lengthBuf2 := []byte{byte(length2 >> 8), byte(length2)}
	conn.Write(lengthBuf2)
	conn.Write(encryptedTestData)

	fmt.Println("Sent test data")

	// 读取响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}

	fmt.Printf("Received response (%d bytes): %s\n", n, string(buffer[:n]))
}
