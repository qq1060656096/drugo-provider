
## conf/db.yaml
```yaml
db:
  public: # 默认组
    test_common:
      name: "test_common"
      dsn: "root:123456@tcp(172.16.123.1:3306)/test_common?charset=utf8mb4&parseTime=true"
      driver_type: "mysql"
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 3600

  business: # 业务组
    test_data_1:
      name: "test_common"
      dsn: "root:123456@tcp(172.16.123.1:3306)/test_common?charset=utf8mb4&parseTime=true"
      driver_type: "mysql"
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 3600

    test_data_2:
      name: "test_common"
      dsn: "root:123456@tcp(172.16.123.1:3306)/test_common?charset=utf8mb4&parseTime=true"
      driver_type: "mysql"
      max_idle_conns: 10
      max_open_conns: 100
      conn_max_lifetime: 3600
```

```go
app := drugo.MustNewApp(
		drugo.WithContext(ctx),
		drugo.WithRoot(root),
		drugo.WithService(dbsvc.New()),
	)
```