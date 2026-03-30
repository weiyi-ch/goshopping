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
	uniqueGroupID := fmt.Sprintf("api-cache-sync-%s", hostname)
	logx.Infof("🔧 Kafka 配置检查: Brokers=%v, Topic=%s, GroupID=%s",
		s.Config.Kafka.Addrs, s.Config.Kafka.Topic, uniqueGroupID)

	// 创建 Kafka 消费者
	q, err := kq.NewQueue(kq.KqConf{
		Brokers: s.Config.Kafka.Addrs,
		Topic:   s.Config.Kafka.Topic,
		Group:   uniqueGroupID,
	}, kq.WithHandle(func(ctx context.Context, key, value string) error {
		logx.Infof("🔥 收到Kafka消息: key=%s value=%s", key, value)

		// 使用 recover 捕获潜在 panic
		defer func() {
			if r := recover(); r != nil {
				logx.WithContext(ctx).Errorf("⚠ panic recovered in cache purge: %v", r)
			}
		}()

		if value == "" {
			logx.Infof("⚠ Kafka 消息 value 为空, 跳过处理")
			return nil
		}

		cacheKey := fmt.Sprintf("product:id:%s", value)

		// 删除本地缓存操作
		if s.LocalCache != nil {
			s.LocalCache.Del(cacheKey)
			if _, ok := s.LocalCache.Get(cacheKey); ok {
				logx.Infof("❌ 删除失败：Key %s 依然存在", cacheKey)
			} else {
				logx.Infof("✅ 成功：本地缓存已清空 Key=%s", cacheKey)
			}
		} else {
			logx.Infof("⚠ LocalCache 为 nil, 无法删除缓存")
		}

		// 异步记录操作日志
		go logx.WithContext(ctx).Infof("RPC L1 Cache Purged: %s", cacheKey)
		return nil
	}))
	if err != nil {
		logx.Errorf("❌ failed to create Kafka queue: %v", err)
		return
	}

	// 启动消费者（非阻塞）并打印状态信息
	go func() {
		logx.Infof("🚀 Kafka 消费者正在后台启动 [Topic: %s]", s.Config.Kafka.Topic)

		// 包裹 Start，捕获异常退出
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("❌ Kafka 消费者异常退出 panic: %v", r)
			} else {
				logx.Errorf("❌ Kafka 消费者已异常退出！")
			}
		}()

		// 启动消费者
		q.Start()
	}()

}
