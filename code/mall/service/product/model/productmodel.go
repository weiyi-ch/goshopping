package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 编译器检查 *customProductModel 是否完全实现了 ProductModel 接口里定义的所有方法。
// 实则defaultProductModel是否实现了productModel的方法
var _ ProductModel = (*customProductModel)(nil)

type (
	// ProductModel is an interface to be customized, add more methods here,
	// and implement the added methods in customProductModel.
	//ProductModel继承productModel的所有方法
	ProductModel interface {
		productModel
	}
	//customProductModel 会自动“继承” defaultProductModel 的所有方法
	customProductModel struct {
		*defaultProductModel
	}
)

// NewProductModel returns a model for the database table.
func NewProductModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ProductModel {
	return &customProductModel{
		defaultProductModel: newProductModel(conn, c, opts...),
	}
}
