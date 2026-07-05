package crud

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

// MemStore 是 Repo 的内存实现，供各业务模块的 service 单元测试使用（无需 Mongo）。
// 它记录被调用时的入参，并允许测试预置返回值，从而校验 service 的业务逻辑。
type MemStore[T any] struct {
	Items []T

	// 可预置的返回值（用于唯一性校验、详情等场景）
	FindOneResult  *T
	FindByIDResult *T
	DeleteCount    int64 // 默认 1

	// 记录最近一次调用的入参，供断言
	Seq           int64
	LastQuery     bson.M
	LastInsert    any
	LastUpdateID  int64
	LastUpdateSet bson.M
	LastDeletedID int64
}

// NewMemStore 创建内存仓储，默认删除返回 1 条。
func NewMemStore[T any]() *MemStore[T] {
	return &MemStore[T]{DeleteCount: 1}
}

func (m *MemStore[T]) List(_ context.Context, query bson.M, _ bson.D, _, _ int) ([]T, int64, error) {
	m.LastQuery = query
	return m.Items, int64(len(m.Items)), nil
}

func (m *MemStore[T]) FindByID(_ context.Context, _ int64) (*T, error) {
	return m.FindByIDResult, nil
}

func (m *MemStore[T]) FindOne(_ context.Context, query bson.M) (*T, error) {
	m.LastQuery = query
	return m.FindOneResult, nil
}

func (m *MemStore[T]) Count(_ context.Context, _ bson.M) (int64, error) {
	return int64(len(m.Items)), nil
}

func (m *MemStore[T]) NextID(_ context.Context) (int64, error) {
	m.Seq++
	return m.Seq, nil
}

func (m *MemStore[T]) Insert(_ context.Context, doc any) error {
	m.LastInsert = doc
	return nil
}

func (m *MemStore[T]) Update(_ context.Context, id int64, set bson.M) error {
	m.LastUpdateID = id
	m.LastUpdateSet = set
	return nil
}

func (m *MemStore[T]) Delete(_ context.Context, id int64) (int64, error) {
	m.LastDeletedID = id
	return m.DeleteCount, nil
}
