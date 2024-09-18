package utility

import (
	"context"
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

var ErrorFieldsIllegal = errors.New("DB Illigal field")

type DB struct {
	orm.Ormer
	ctx context.Context
}

func NewDB() *DB {
	return Orm()
}

func Orm(aliasName ...string) *DB {
	name := "default"
	if len(aliasName) != 0 && aliasName[0] != "" {
		name = aliasName[0]
	}
	return &DB{orm.NewOrmUsingDB(name), context.Background()}
}

func (d *DB) Get(m any, cols ...string) (err error) {
	return d.ReadWithCtx(d.getCtx(), m, cols...)
}

// Begin 创建事务
func (d *DB) Begin() (*TxOrm, error) {
	tx, err := d.Ormer.BeginWithCtx(d.getCtx())
	if err != nil {
		return nil, err
	}
	return &TxOrm{TxOrmer: tx}, nil
}

// Count 数量
func (d *DB) Count(i any, fields ...any) (count int64, err error) {
	err = verifyFields(fields)
	if err != nil {
		return
	}
	qs := d.QueryTable(i)
	for i := 0; i < len(fields)/2; i++ {
		qs = qs.Filter(fields[i*2+0].(string), fields[i*2+1])
	}
	count, err = qs.CountWithCtx(d.getCtx())
	return
}

// 校验filter是否符合规定
func verifyFields(fields []any) error {
	if len(fields)%2 != 0 {
		return ErrorFieldsIllegal
	}
	return nil
}

func (d *DB) getCtx() context.Context {
	if d.ctx == nil {
		return context.Background()
	}
	return d.ctx
}

type TxOrm struct {
	orm.TxOrmer
	ctx context.Context
}

func (tx *TxOrm) getCtx() context.Context {
	if tx.ctx == nil {
		return context.Background()
	}
	return tx.ctx
}

func (tx *TxOrm) Get(m any, cols ...string) (err error) {
	return tx.ReadWithCtx(tx.getCtx(), m, cols...)
}

// Count 数量
func (tx *TxOrm) Count(i any, fields ...any) (count int64, err error) {
	err = verifyFields(fields)
	if err != nil {
		return
	}
	qs := tx.QueryTable(i)
	for i := 0; i < len(fields)/2; i++ {
		qs = qs.Filter(fields[i*2+0].(string), fields[i*2+1])
	}
	count, err = qs.CountWithCtx(tx.getCtx())
	return
}
