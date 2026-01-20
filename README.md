
```go
import(
"github.com/qq1060656096/drugo-provider/dbsvc"
"github.com/qq1060656096/drugo-provider/ginsrv"
"github.com/qq1060656096/drugo-provider/redissvc"
"github.com/qq1060656096/drugo/drugo"
)
// 获取数据库 service
dbSvc := ginsrv.MustGetService[*drugo.Drugo, *dbsvc.DbService](c, dbsvc.Name)
db := dbSvc.Manager().MustGroup("public").MustGet(c.Request.Context(), "test_common")
companyInfo := make(map[string]interface{})
db.Raw("select * from common_company where company_id= 218908").Scan(&companyInfo)
// 获取 redis 缓存
redisSvc := ginsrv.MustGetService[*drugo.Drugo, *redissvc.RedisService](c, redissvc.Name)
r, err := redisSvc.Group().MustGet(c, "session").Set(c.Request.Context(), "api", "demo", 0).Result()
```