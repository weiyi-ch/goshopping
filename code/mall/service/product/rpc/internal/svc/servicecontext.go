package svc

import (
	"context"
	"fmt"
	"mall/service/product/model"
	"mall/service/product/rpc/internal/config"
	"os"
	"time"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	ProductModel model.ProductModel
	LocalCache   *collection.Cache
	KqPusher     *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	//在生成model.product时增加-c参数因此说明需要缓存来记录
	//单例化svc 其中首先包含config的数据，再包含model.ProductModel模型，把逻辑服务与数据库缓存等操作解耦。
	//在生成model时便构建连接池和redis缓存，实现单例化和连接池管理连接避免连接过多服务崩溃。
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	localCache, _ := collection.NewCache(2 * time.Minute)
	svc := &ServiceContext{
		Config:       c,
		ProductModel: model.NewProductModel(conn, c.CacheRedis),
		LocalCache:   localCache,
		KqPusher:     kq.NewPusher(c.Kafka.Addrs, c.Kafka.Topic),
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
			if _, ok := s.LocalCache.Get(cacheKey); ok {
				fmt.Println("失败：Key 依然存在，可能删错对象了")
			} else {
				fmt.Println("成功：本地缓存已清空")
			}
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
	go func() {
		logx.Infof("🚀 Kafka 消费者正在后台启动 [Topic: %s]", s.Config.Kafka.Topic)
		q.Start()
		// 如果走到这一行，说明消费者因为错误退出了
		logx.Errorf("❌ 警报：Kafka 消费者已异常退出！")
	}()
}
