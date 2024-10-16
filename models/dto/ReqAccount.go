package dto

type ReqAccountDetail struct {
	AccountId  int64
	Username   string
	Email      string
	CreateTime int64
}

type ReqEditAccount struct {
	AccountId   int64 `valid:"Required"`
	Email       string
	Password    string
	NewPassword string
	ValidCode   string
	Phone       int
	CountryCode int
}
