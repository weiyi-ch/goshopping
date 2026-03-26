// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"mall/service/product/api/internal/config"
	"mall/service/product/rpc/productclient"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	ProductRpc productclient.Product
	LocalCache *collection.Cache
}

func NewServiceContext(c config.Config) *ServiceContext {
	localCache, _ := collection.NewCache(2 * time.Minute)
	return &ServiceContext{
		Config:     c,
		ProductRpc: productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		LocalCache: localCache,
	}
}
