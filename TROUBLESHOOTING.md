# 故障排除指南

## 连接问题诊断

### 常见错误类型

#### 1. Unexpected EOF 错误
```
"Failed to handle proxy connection from 106.226.79.72:31190: failed to read target: failed to read encrypted data (expected 14631 bytes): unexpected EOF"
```

**原因分析：**
- 客户端连接在服务器读取数据时意外关闭
- 网络不稳定导致连接中断
- 客户端超时设置过短
- 协议不匹配或数据格式错误

**解决方案：**
1. 检查客户端配置是否正确
2. 增加客户端超时时间
3. 检查网络连接稳定性
4. 验证加密方法和密码设置

#### 2. 超时错误
```
"timeout reading encrypted data (read 1024/14631 bytes): i/o timeout"
```

**原因分析：**
- 网络延迟过高
- 服务器负载过重
- 客户端发送数据速度过慢

**解决方案：**
1. 增加连接超时时间
2. 检查服务器性能
3. 优化网络配置

#### 3. 长度错误
```
"invalid encrypted data length: 70000"
```

**原因分析：**
- 数据长度超出限制（最大65535字节）
- 协议版本不匹配
- 数据损坏

**解决方案：**
1. 检查客户端协议版本
2. 验证数据完整性
3. 更新客户端软件

### 配置优化建议

#### 1. 连接监控设置
```yaml
shadowsocks:
  connection_monitoring:
    enabled: true
    log_failed_connections: true
    log_successful_connections: false
    max_connection_retries: 3
    connection_timeout_seconds: 60
```

#### 2. 超时设置
```yaml
shadowsocks:
  timeout: 300  # 增加超时时间到5分钟
```

#### 3. 日志级别
```yaml
log:
  level: "debug"  # 临时设置为debug以获取更多信息
```

### 监控和诊断

#### 1. 启用详细日志
设置日志级别为 `debug` 可以获取更详细的连接信息。

#### 2. 连接统计
系统会自动记录连接统计信息，包括：
- 成功连接数
- 失败连接数
- 平均响应时间
- 错误类型分布

#### 3. 网络诊断
使用内置的诊断工具测试连接：
```bash
# 测试服务器连接
curl -X POST http://localhost:8080/api/v1/diagnose/connection \
  -H "Content-Type: application/json" \
  -d '{"host":"example.com","port":443}'
```

### 预防措施

1. **定期更新客户端软件**
2. **监控服务器资源使用情况**
3. **设置合理的超时时间**
4. **启用连接监控**
5. **定期检查日志文件**

### 紧急处理

如果遇到大量连接错误：

1. **立即检查服务器状态**
   ```bash
   systemctl status vps-vpn
   ```

2. **查看实时日志**
   ```bash
   tail -f /var/log/vps-vpn.log
   ```

3. **重启服务**
   ```bash
   systemctl restart vps-vpn
   ```

4. **检查网络连接**
   ```bash
   netstat -tulpn | grep 8388
   ```

### 联系支持

如果问题持续存在，请提供以下信息：
- 错误日志完整内容
- 客户端配置信息
- 服务器系统信息
- 网络环境描述 