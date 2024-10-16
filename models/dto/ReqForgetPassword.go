package dto

type ReqForgetPassword struct {
	Username string `valid:"Required;IsUsername"`
	Email    string `valid:"Required;Email"`
}

type ReqForgetPasswordSetNew struct {
	Username  string `valid:"Required;IsUsername"`
	Email     string `valid:"Required;Email"`
	Password  string `valid:"Required;IsPassword"`
	ValidCode string `valid:"Required"`
}
