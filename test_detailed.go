package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

func main() {
	fmt.Println("=== 详细连接测试 ===")

	// 连接本地服务器
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8388", 10*time.Second)
	if err != nil {
		fmt.Printf("连接服务器失败: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("✓ 成功连接到本地服务器")

	// 生成salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		fmt.Printf("生成salt失败: %v\n", err)
		return
	}
	fmt.Printf("生成salt: %x\n", salt[:8])

	// 发送salt
	if _, err := conn.Write(salt); err != nil {
		fmt.Printf("发送salt失败: %v\n", err)
		return
	}
	fmt.Println("✓ 已发送salt")

	// 创建加密器
	key := pbkdf2.Key([]byte("13687401432Fan!"), salt, 10000, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("创建AES加密器失败: %v\n", err)
		return
	}
	encryptor, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("创建GCM加密器失败: %v\n", err)
		return
	}
	fmt.Println("✓ 加密器创建成功")

	// 构造目标地址数据
	target := "httpbin.org:80"
	targetData := constructTargetData(target)
	fmt.Printf("目标地址: %s, 数据长度: %d\n", target, len(targetData))

	// 加密目标地址数据
	nonce := make([]byte, encryptor.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		fmt.Printf("生成nonce失败: %v\n", err)
		return
	}
	encryptedTarget := encryptor.Seal(nonce, nonce, targetData, nil)

	// 发送加密的目标地址数据
	length := len(encryptedTarget)
	lengthBuf := []byte{byte(length >> 8), byte(length)}
	if _, err := conn.Write(lengthBuf); err != nil {
		fmt.Printf("发送长度失败: %v\n", err)
		return
	}
	if _, err := conn.Write(encryptedTarget); err != nil {
		fmt.Printf("发送加密目标地址失败: %v\n", err)
		return
	}
	fmt.Println("✓ 已发送加密的目标地址数据")

	// 发送测试HTTP请求
	testData := []byte("GET /ip HTTP/1.1\r\nHost: httpbin.org\r\nConnection: close\r\n\r\n")
	fmt.Printf("发送测试数据: %s\n", string(testData))

	// 加密测试数据
	nonce = make([]byte, encryptor.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		fmt.Printf("生成nonce失败: %v\n", err)
		return
	}
	encryptedData := encryptor.Seal(nonce, nonce, testData, nil)

	// 发送长度和数据
	length = len(encryptedData)
	lengthBuf = []byte{byte(length >> 8), byte(length)}
	if _, err := conn.Write(lengthBuf); err != nil {
		fmt.Printf("发送数据长度失败: %v\n", err)
		return
	}
	if _, err := conn.Write(encryptedData); err != nil {
		fmt.Printf("发送加密数据失败: %v\n", err)
		return
	}
	fmt.Println("✓ 已发送加密的测试数据")

	// 等待响应
	fmt.Println("等待响应...")
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// 读取响应
	responseBuf := make([]byte, 4096)
	n, err := conn.Read(responseBuf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("⚠️  超时 - 服务器没有响应")
		} else {
			fmt.Printf("❌ 读取响应失败: %v\n", err)
		}
		return
	}

	fmt.Printf("✓ 收到响应: %d 字节\n", n)
	fmt.Printf("响应数据: %s\n", string(responseBuf[:n]))

	// 尝试解密响应
	if n > 0 {
		// 尝试解密响应
		decryptedResponse, err := encryptor.Open(nil, nil, responseBuf[:n], nil)
		if err != nil {
			fmt.Printf("解密响应失败: %v\n", err)
		} else {
			fmt.Printf("✓ 解密响应: %s\n", string(decryptedResponse))
		}
	}

	fmt.Println("测试完成！")
}

func constructTargetData(target string) []byte {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "80"
	}

	portNum := 80
	if port != "" {
		if p, err := net.LookupPort("tcp", port); err == nil {
			portNum = p
		}
	}

	var addrData []byte

	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			addrData = append(addrData, 0x01)
			addrData = append(addrData, ip.To4()...)
		} else {
			addrData = append(addrData, 0x04)
			addrData = append(addrData, ip...)
		}
	} else {
		addrData = append(addrData, 0x03)
		addrData = append(addrData, byte(len(host)))
		addrData = append(addrData, []byte(host)...)
	}

	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(portNum))
	addrData = append(addrData, portBytes...)

	return addrData
}
