package logic

import (
	"context"
	"mall/service/product/model"
	"strconv"

	"mall/service/product/rpc/internal/svc"
	"mall/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type RemoveLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveLogic {
	return &RemoveLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveLogic) Remove(in *product.RemoveRequest) (*product.RemoveResponse, error) {
	// 查询产品是否存在
	res, err := l.svcCtx.ProductModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(100, "产品不存在")
		}
		return nil, status.Error(500, err.Error())
	}

	err = l.svcCtx.ProductModel.Delete(l.ctx, res.Id)
	if err != nil {
		return nil, status.Error(500, err.Error())
	}
	// 🔥 4. 【核心新增】发送 Kafka 消息，通知所有 API 节点清理本地 L1 缓存
	// 我们只需要把 ID 发送过去即可，API 节点收到后会执行 LocalCache.Del("api:product:ID")
	productIdStr := strconv.FormatUint(uint64(in.Id), 10)

	// 使用异步 Push，不影响主业务响应速度
	err = l.svcCtx.KqPusher.Push(l.ctx, productIdStr)
	if err != nil {
		// 建议打印错误日志，但不直接返回 error 给前端，防止 Kafka 抖动导致业务不可用
		logx.WithContext(l.ctx).Errorf("Kafka 推送失败，productId: %s, err: %v", productIdStr, err)
	}
	return &product.RemoveResponse{}, nil
}
