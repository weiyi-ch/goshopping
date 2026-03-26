package logic

import (
	"context"
	"fmt"
	"mall/service/product/rpc/internal/svc"
	"mall/service/product/rpc/types/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type DetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DetailLogic) Detail(in *product.DetailRequest) (*product.DetailResponse, error) {

	cacheKey := fmt.Sprintf("product:id:%d", in.Id)

	if val, ok := l.svcCtx.LocalCache.Get(cacheKey); ok {
		return val.(*product.DetailResponse), nil
	}
	res, err := l.svcCtx.ProductModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		return nil, err
	}

	resp := &product.DetailResponse{
		Id:     int64(res.Id),
		Name:   res.Name,
		Desc:   res.Desc,
		Stock:  int64(res.Stock),
		Amount: int64(res.Amount),
		Status: int64(res.Status),
	}

	l.svcCtx.LocalCache.Set(cacheKey, resp)

	return resp, nil
}
