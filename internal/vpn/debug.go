package vpn

import (
	"encoding/hex"
	"io"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// DebugConnection 调试连接信息
type DebugConnection struct {
	conn   net.Conn
	logger *logrus.Logger
}

// NewDebugConnection 创建调试连接
func NewDebugConnection(conn net.Conn, logger *logrus.Logger) *DebugConnection {
	return &DebugConnection{
		conn:   conn,
		logger: logger,
	}
}

// DebugReadLength 调试读取长度字段
func (d *DebugConnection) DebugReadLength(lengthBytes int) ([]byte, error) {
	lengthBuf := make([]byte, lengthBytes)

	// 设置读取超时
	d.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	n, err := io.ReadFull(d.conn, lengthBuf)
	if err != nil {
		d.logger.Errorf("Debug: Failed to read length field: %v (read %d bytes)", err, n)
		return nil, err
	}

	d.logger.Infof("Debug: Read length field: %s (hex)", hex.EncodeToString(lengthBuf))
	return lengthBuf, nil
}

// DebugReadData 调试读取数据
func (d *DebugConnection) DebugReadData(length int) ([]byte, error) {
	data := make([]byte, length)

	// 设置读取超时
	d.conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	n, err := io.ReadFull(d.conn, data)
	if err != nil {
		d.logger.Errorf("Debug: Failed to read data: %v (read %d/%d bytes)", err, n, length)
		return nil, err
	}

	d.logger.Infof("Debug: Read data: %d bytes", n)
	return data, nil
}

// DebugConnectionInfo 调试连接信息
func DebugConnectionInfo(conn net.Conn, logger *logrus.Logger) {
	logger.Infof("Debug: New connection from %s", conn.RemoteAddr().String())
	logger.Infof("Debug: Local address: %s", conn.LocalAddr().String())

	// 设置连接属性
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		logger.Infof("Debug: TCP keep-alive enabled")
	}
}

// DebugProtocolVersion 调试协议版本
func DebugProtocolVersion(data []byte, logger *logrus.Logger) {
	if len(data) >= 4 {
		logger.Infof("Debug: Protocol data: %s (hex)", hex.EncodeToString(data[:4]))
	}
}
