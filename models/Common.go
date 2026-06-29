package models

type CommStruct struct {
	Id         int64  `orm:"description(PK)"`
	CreateTime uint64 `orm:"description(Created Time)"`
	UpdateTime uint64 `orm:"description(Updated Time)"`
	Deleted    bool   `orm:"description(Is this row of data soft deleted, 0=false, 1=ture)"`
}
