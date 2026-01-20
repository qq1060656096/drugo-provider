
### conf/gin.yaml
```yaml
gin:
  mode: release           # debug, release, test
  host: "0.0.0.0"
  shutdown_timeout: 30s   # 优雅关闭超时
  read_timeout: 15s       # 请求读取超时
  write_timeout: 15s      # 响应写入超时
  idle_timeout: 60s       # Keep-Alive 空闲超时
  # HTTP 配置
  http:
    enabled: true
    port: 18001

  # HTTPS 配置
  https:
    enabled: false
    port: 18443
    cert_file: "./cert/server.crt"
    key_file: "./cert/server.key"
    force_ssl: false

```

```go
app := drugo.MustNewApp(
		drugo.WithContext(ctx),
		drugo.WithRoot(root),
		drugo.WithService(ginsrv.New()),
	)
```