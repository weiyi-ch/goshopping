// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"context"
	"fmt"
	"mall/service/product/api/internal/config"
	"mall/service/product/rpc/productclient"
	"os"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	ProductRpc productclient.Product
	LocalCache *collection.Cache
}

func NewServiceContext(c config.Config) *ServiceContext {
	localCache, _ := collection.NewCache(2 * time.Minute)
	svc := &ServiceContext{
		Config:     c,
		ProductRpc: productclient.NewProduct(zrpc.MustNewClient(c.ProductRpc)),
		LocalCache: localCache,
	}
	svc.initRpcLocalCacheWatcher()
	return svc
}
func (s *ServiceContext) initRpcLocalCacheWatcher() {
	// 获取 hostname 作为唯一消费者组ID
	hostname, err := os.Hostname()
	if err != nil {
		logx.Errorf("failed to get hostname: %v, fallback to 'unknown'", err)
		hostname = "unknown"
	}
	uniqueGroupID := fmt.Sprintf("rpc-cache-sync-%s", hostname)

	// 创建 Kafka 消费者
	q, err := kq.NewQueue(kq.KqConf{
		Brokers: s.Config.Kafka.Addrs,
		Topic:   s.Config.Kafka.Topic,
		Group:   uniqueGroupID,
	}, kq.WithHandle(func(ctx context.Context, key, value string) error {
		if value == "" {
			return nil
		}

		// 使用 recover 捕获潜在 panic
		defer func() {
			if r := recover(); r != nil {
				logx.WithContext(ctx).Errorf("panic recovered in cache purge: %v", r)
			}
		}()

		cacheKey := fmt.Sprintf("product:id:%s", value)

		// 删除缓存操作需线程安全，确保 LocalCache 实现支持并发
		if s.LocalCache != nil {
			s.LocalCache.Del(cacheKey)
		}

		// 异步记录日志，减少高并发压力
		go logx.WithContext(ctx).Infof("RPC L1 Cache Purged: %s", cacheKey)

		return nil
	}))
	if err != nil {
		logx.Errorf("failed to create kafka queue: %v", err)
		return
	}

	// 启动消费者（非阻塞）
	go q.Start()
}
