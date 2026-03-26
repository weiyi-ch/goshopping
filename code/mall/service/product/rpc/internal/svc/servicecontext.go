package svc

import (
	"mall/service/product/model"
	"mall/service/product/rpc/internal/config"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	ProductModel model.ProductModel
	LocalCache   *collection.Cache
}

func NewServiceContext(c config.Config) *ServiceContext {
	//在生成model.product时增加-c参数因此说明需要缓存来记录
	//单例化svc 其中首先包含config的数据，再包含model.ProductModel模型，把逻辑服务与数据库缓存等操作解耦。
	//在生成model时便构建连接池和redis缓存，实现单例化和连接池管理连接避免连接过多服务崩溃。
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	localCache, _ := collection.NewCache(2 * time.Minute)
	return &ServiceContext{
		Config:       c,
		ProductModel: model.NewProductModel(conn, c.CacheRedis),
		LocalCache:   localCache,
	}
}
