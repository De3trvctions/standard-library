package dto

type ReqRegister struct {
	Username  string `valid:"Required;IsUsername"`
	Password  string `valid:"Required;IsPassword"`
	Email     string `valid:"Required;Email"`
	ValidCode string `valid:"Required"`
}
