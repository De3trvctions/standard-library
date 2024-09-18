package validation

import (
	"encoding/json"
	"fmt"
	"net"
	neturl "net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/core/validation"
)

func Init() {
	logs.Info("-----新增Beego自定义验证开始-----")
	if err := validation.AddCustomFunc("IsAlipay", IsAlipay); err != nil {
		logs.Error("IsAlipay is not registered")
	}
	if err := validation.AddCustomFunc("IsBindPhone", IsBindPhone); err != nil {
		logs.Error("IsBindPhone is not registered")
	}
	if err := validation.AddCustomFunc("IsJson", IsJson); err != nil {
		logs.Error("IsJson is not registered")
	}
	if err := validation.AddCustomFunc("IsDescription", IsDescription); err != nil {
		logs.Error("IsDescription is not registered")
	}
	if err := validation.AddCustomFunc("IsDescriptionNoSpace", IsDescriptionNoSpace); err != nil {
		logs.Error("IsDescriptionNoSpace is not registered")
	}
	if err := validation.AddCustomFunc("IP", IP); err != nil {
		logs.Error("IP is not registered")
	}
	if err := validation.AddCustomFunc("Phone", Phone); err != nil {
		logs.Error("Phone is not registered")
	}
	if err := validation.AddCustomFunc("IsUsername", IsUsername); err != nil {
		logs.Error("IsUsername is not registered")
	}
	if err := validation.AddCustomFunc("IsLoginAccounts", IsLoginAccounts); err != nil {
		logs.Error("IsLoginAccounts is not registered")
	}
	if err := validation.AddCustomFunc("IsUsernameNetCash", IsUsernameNetCash); err != nil {
		logs.Error("IsUsernameNetCash is not registered")
	}
	if err := validation.AddCustomFunc("IsPassword", IsPassword); err != nil {
		logs.Error("IsPassword is not registered")
	}
	if err := validation.AddCustomFunc("IsUrl", IsUrl); err != nil {
		logs.Error("IsUrl is not registered")
	}
	if err := validation.AddCustomFunc("IsNumberComma", IsNumberComma); err != nil {
		logs.Error("IsNumberComma is not registered")
	}
	if err := validation.AddCustomFunc("Is24HourTime", Is24HourTime); err != nil {
		logs.Error("Is24HourTime is not registered")
	}
	if err := validation.AddCustomFunc("IsAlphaDashComma", IsAlphaDashComma); err != nil {
		logs.Error("IsAlphaDashComma is not registered")
	}
	if err := validation.AddCustomFunc("IsAlphaComma", IsAlphaComma); err != nil {
		logs.Error("IsAlphaComma is not registered")
	}
	if err := validation.AddCustomFunc("Min1", Min1); err != nil {
		logs.Error("Min1 is not registered")
	}
	if err := validation.AddCustomFunc("Min0", Min0); err != nil {
		logs.Error("Min0 is not registered")
	}
	if err := validation.AddCustomFunc("IsVipLevel", IsVipLevel); err != nil {
		logs.Error("IsVipLevel is not registered")
	}
	if err := validation.AddCustomFunc("IsEditVipLevel", IsEditVipLevel); err != nil {
		logs.Error("IsEditVipLevel is not registered")
	}
	if err := validation.AddCustomFunc("IsRealName", IsRealName); err != nil {
		logs.Error("IsRealName is not registered")
	}
	if err := validation.AddCustomFunc("IsDescriptionNoChineseComma", IsDescriptionNoChineseComma); err != nil {
		logs.Error("IsDescriptionNoChineseComma is not registered")
	}
	if err := validation.AddCustomFunc("IsVersionName", IsVersionName); err != nil {
		logs.Error("IsVersionName is not registered")
	}
	if err := validation.AddCustomFunc("IsSafetyCode", IsSafetyCode); err != nil {
		logs.Error("IsSafetyCode is not registered")
	}
	if err := validation.AddCustomFunc("IsNewCreditNetPassword", IsNewCreditNetPassword); err != nil {
		logs.Error("IsNewCreditNetPassword is not registered")
	}
	if err := validation.AddCustomFunc("IsFirstLoginUsername", IsFirstLoginUsername); err != nil {
		logs.Error("IsFirstLoginUsername is not registered")
	}
	if err := validation.AddCustomFunc("IsCreditNetGameLoginAccount", IsCreditNetGameLoginAccount); err != nil {
		logs.Error("IsCreditNetGameLoginAccount is not registered")
	}
	logs.Info("-----新增Beego自定义验证结束-----")
}

// ValidateRequest returns error based on the valid() StructTag setting in the parameter req
// the req parameter must be a struct or a struct pointer
func ValidateRequest(req interface{}) error {

	valid := validation.Validation{}
	canSkipFuncInit(&valid)

	if b, err := valid.RecursiveValid(req); err != nil {
		logs.Error("[ValidateRequest error]:", err)
		return err
	} else if !b {
		for _, e := range valid.Errors {
			logs.Error("[ValidateRequest error]:", e)
			return e
		}
	}

	return nil
}

func canSkipFuncInit(v *validation.Validation) {
	//valid.RequiredFirst = true可以激活beego内置CanSkipFuncs
	//从而跳过field为空的valid
	v.RequiredFirst = true

	// Add this Schema to CanSkipFunc
	// This is to make sure that when pass in empty data to the Schema, error wont be throw and validation will be skipped
	v.CanSkipAlso("Range")
	v.CanSkipAlso("Min")
}

func Min1(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "Min1"
	check := reflect.TypeOf(obj).String()
	ok := true
	var num int
	switch check {
	case "int":
		num, ok = obj.(int)
	case "int64":
		var newNum int64
		newNum, ok = obj.(int64)
		num = int(newNum)
	case "int32":
		var newNum int32
		newNum, ok = obj.(int32)
		num = int(newNum)
	case "uint":
		var newNum uint
		newNum, ok = obj.(uint)
		num = int(newNum)
	}
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "incorrect format", obj.(string), fieldName))
	}
	if num < 1 {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "value is less than 1", obj.(string), fieldName))
	}
}

func Min0(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "Min1"
	check := reflect.TypeOf(obj).String()
	ok := true
	var num int
	switch check {
	case "int":
		num, ok = obj.(int)
	case "int64":
		var newNum int64
		newNum, ok = obj.(int64)
		num = int(newNum)
	case "int32":
		var newNum int32
		newNum, ok = obj.(int32)
		num = int(newNum)
	case "uint":
		var newNum uint
		newNum, ok = obj.(uint)
		num = int(newNum)
	}
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "incorrect format", obj.(string), fieldName))
	}
	if num < 0 {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "value is less than 1", obj.(string), fieldName))
	}
}

func IsJson(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsJson"
	var x interface{}
	data, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", data, fieldName))
	}
	if len(data) > 0 {
		err := json.Unmarshal([]byte(data), &x)
		if err != nil {
			logs.Error("[Validator IsJson error]:", err)
			_ = v.SetError(fieldName, buildErrorMsg(funcName, "JSON format incorrect", data, fieldName))
		}
	}

	// This is to print the type of the data, if it is a object/struct, then it will return map,
	// is list of string, then will be returning string
	// tv := reflect.TypeOf(v[0])
	// fmt.Println(tv.Kind())
	// fmt.Printf("\n%+v\n", v)
}

func IsAlipay(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsAlipay"
	acc, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", acc, fieldName))
		return
	}
	if len(acc) == 0 {
		return
	}

	v.Email(obj, key)
	if !v.HasErrors() {
		return
	}
	v.Clear()

	phonePattern := "^[0-9]{1,4}[-_]?[0-9]+$"
	reg := regexp.MustCompile(phonePattern)
	if reg.MatchString(acc) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "either email or phone number", acc, fieldName))
}

func IsBindPhone(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsBindPhone"
	if len(obj.(string)) == 0 {
		return
	}
	n, ok := toInt(obj)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", obj.(string), fieldName))
		return
	}

	if n == -1 || n == 1 || n == 2 {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "has to be either 1 or 2", obj.(string), fieldName))
}

func toInt(obj interface{}) (int, bool) {
	val, ok := obj.(string)
	if !ok {
		return 0, false
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}
	return n, true
}

/**
 * IsDescription()
 * This function will check on the field which will match with
 * 	- Chinese Character
 * 	- English Character
 * 	- Numeric Character
 * 	- Special Symbol as such [(space), (.), (_), (-)]
 */
func IsDescription(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsDescription"
	desc, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", desc, fieldName))
	}
	if len(desc) == 0 {
		return
	}
	// descriptionPattern := `^([0-9a-zA-Z\p{Han}-\[\]_.·()【】])([ 0-9a-zA-Z\p{Han}-\[\]_.·()【】]*)$`
	descriptionPattern := `^([[:alpha:]\p{Han}\pNl\pP])([[:alpha:][:blank:]\p{Han}\pNl\pP])*$`
	regex := regexp.MustCompile(descriptionPattern)
	if regex.MatchString(desc) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "description pattern does not match", desc, fieldName))
}

func IsDescriptionNoChineseComma(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsDescriptionNoChineseComma"
	desc, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", desc, fieldName))
	}
	if len(desc) == 0 {
		return
	}
	descriptionPattern := `^([[:alnum:]\.]*)(,([[:alnum:]\.]*))*$`
	regex := regexp.MustCompile(descriptionPattern)
	if regex.MatchString(desc) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "description pattern does not match", desc, fieldName))
}

func IsDescriptionNoSpace(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsDescriptionNoSpace"
	desc, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", desc, fieldName))
	}
	if len(desc) == 0 {
		return
	}
	// descriptionPattern := `^[0-9a-zA-Z\p{Han}-_.·()]*$`
	descriptionPattern := `^[[:alpha:]\p{Han}\pNl\pP]*$`
	regex := regexp.MustCompile(descriptionPattern)
	if regex.MatchString(desc) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "description with no space pattern does not match", desc, fieldName))
}

// IP Overwrites beego IP validation function so that both ipv4 and ipv6 can be validated
func IP(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IP"
	val, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", val, fieldName))
		return
	}
	if ip := net.ParseIP(val); ip != nil {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid Ip address", val, fieldName))
}

func Phone(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "Phone"
	val, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", val, fieldName))
		return
	}
	if len(val) <= 0 {
		return
	}
	phonePattern := "^[0-9]{1,4}[-_]?[0-9]+$"
	reg := regexp.MustCompile(phonePattern)
	if reg.MatchString(val) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid phone number", val, fieldName))
}

func IsUsername(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsUsername"
	username, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", username, fieldName))
	}
	// If the field is empty, then return true as there is nothing to check
	if len(username) == 0 {
		return
	}
	descriptionPattern := `^[a-zA-Z][a-zA-Z0-9_]{5,15}$`

	regex := regexp.MustCompile(descriptionPattern)
	if regex.MatchString(obj.(string)) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "username pattern does not match", username, fieldName))
}
func IsFirstLoginUsername(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsFirstLoginUsername"
	username, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", username, fieldName))
	}
	if len(username) == 0 {
		return
	}

	pattern4 := `^([[:alnum:]]){6,12}$`
	pattern5 := `^[\S]+$` // Check no space should exist

	isFalse := false
	regex4 := regexp.MustCompile(pattern4)
	if !regex4.MatchString(obj.(string)) {
		isFalse = true
	}
	regex5 := regexp.MustCompile(pattern5)
	if !regex5.MatchString(obj.(string)) {
		isFalse = true
	}

	if !isFalse {
		return
	}

	_ = v.SetError(fieldName, buildErrorMsg(funcName, "creditnet username pattern does not match", username, fieldName))
}

func IsSafetyCode(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsSafetyCode"
	safetycode, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", safetycode, fieldName))
	}
	if len(safetycode) == 0 {
		return
	}

	pattern1 := `^(.*[a-z].*)$`         // Check contain lowercase
	pattern2 := `^(.*[A-Z].*)$`         // Check contain uppercase
	pattern3 := `^(.*\d.*)$`            // Check contain numeric
	pattern4 := `^([[:alnum:]]){8,15}$` // Check length should be 8 - 15
	pattern5 := `^[\S]+$`               // Check no space should exist

	isFalse := false
	regex := regexp.MustCompile(pattern1)
	if !regex.MatchString(obj.(string)) {
		isFalse = true
	}
	regex2 := regexp.MustCompile(pattern2)
	if !regex2.MatchString(obj.(string)) {
		isFalse = true
	}
	regex3 := regexp.MustCompile(pattern3)
	if !regex3.MatchString(obj.(string)) {
		isFalse = true
	}
	regex4 := regexp.MustCompile(pattern4)
	if !regex4.MatchString(obj.(string)) {
		isFalse = true
	}
	regex5 := regexp.MustCompile(pattern5)
	if !regex5.MatchString(obj.(string)) {
		isFalse = true
	}

	if !isFalse {
		return
	}

	_ = v.SetError(fieldName, buildErrorMsg(funcName, "safetycode pattern does not match", safetycode, fieldName))
}

func IsLoginAccounts(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsLoginAccounts"
	acc, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", acc, fieldName))
	}
	// If the field is empty, then return true as there is nothing to check
	if len(acc) == 0 {
		return
	}
	descriptionPattern := `^[a-zA-Z0-9][a-zA-Z0-9_]{4,15}(,[a-zA-Z0-9][a-zA-Z0-9_]{4,15})*$`
	// 暂时更换，有一批错误的游戏账号被导入 无法支持批量
	//descriptionPattern := `^[[:alnum:]]{4,16}$`

	regex := regexp.MustCompile(descriptionPattern)
	// If the field is empty, then return true as there is nothing to check
	if regex.MatchString(obj.(string)) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "login accounts pattern does not match", acc, fieldName))
}

func IsUsernameNetCash(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsUsernameNetCash"
	username, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", username, fieldName))
	}
	if len(username) == 0 {
		return
	}
	descriptionPattern := `^([[:alpha:]])([[:alnum:][:blank:]_]){7,11}$`

	regex := regexp.MustCompile(descriptionPattern)
	// If the field is empty, then return true as there is nothing to check
	if len(obj.(string)) == 0 {
		return
	}
	if regex.MatchString(obj.(string)) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "username pattern does not match", username, fieldName))
}
func IsCreditNetGameLoginAccount(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsCreditNetGameLoginAccount"
	loginAccount, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", loginAccount, fieldName))
	}

	// If the field is empty, then return true as there is nothing to check
	if len(loginAccount) == 0 {
		return
	}
	descriptionPattern := `^([[:alpha:]])([[:alnum:]]){5,14}$`

	regex := regexp.MustCompile(descriptionPattern)
	if regex.MatchString(loginAccount) {
		return
	}

	_ = v.SetError(fieldName, buildErrorMsg(funcName, "LoginAccount pattern does not match", loginAccount, fieldName))
}

func IsPassword(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsPassword"
	descriptionPattern := `^[a-zA-Z0-9_]{6,20}$`

	regex := regexp.MustCompile(descriptionPattern)
	// If the field is empty, then return true as there is nothing to check
	if len(obj.(string)) == 0 {
		return
	}
	if regex.MatchString(obj.(string)) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "password pattern does not match", obj.(string), fieldName))
}
func IsNewCreditNetPassword(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsNewCreditNetPassword"
	safetycode, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", safetycode, fieldName))
	}
	if len(safetycode) == 0 {
		return
	}

	pattern1 := `^(.*[a-z].*){1}$`      // Check contain lowercase
	pattern2 := `^(.*[A-Z].*){1}$`      // Check contain uppercase
	pattern3 := `^(.*\d.*)$`            // Check contain numeric
	pattern4 := `^([[:alnum:]]){6,12}$` // Check length should be 8 - 16
	pattern5 := `^[\S]+$`               // Check no space should exist

	isFalse := false
	regex1 := regexp.MustCompile(pattern1)
	regex2 := regexp.MustCompile(pattern2)

	if !regex1.MatchString(obj.(string)) && !regex2.MatchString(obj.(string)) {
		isFalse = true
	}

	regex3 := regexp.MustCompile(pattern3)
	if !regex3.MatchString(obj.(string)) {
		isFalse = true
	}
	regex4 := regexp.MustCompile(pattern4)
	if !regex4.MatchString(obj.(string)) {
		isFalse = true
	}
	regex5 := regexp.MustCompile(pattern5)
	if !regex5.MatchString(obj.(string)) {
		isFalse = true
	}

	if !isFalse {
		return
	}

	_ = v.SetError(fieldName, buildErrorMsg(funcName, "password pattern does not match", safetycode, fieldName))
}

func IsUrl(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsUrl"
	url, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", url, fieldName))
		return
	}
	if len(obj.(string)) == 0 {
		return
	}
	temp := url
	url = strings.ReplaceAll(url, "http://", "")
	url = strings.ReplaceAll(url, "https://", "")
	url = strings.ReplaceAll(url, "www.", "")
	url = strings.ReplaceAll(url, "w2w.", "")

	// force add a prefix stating for better validation
	if url != temp {
		url = "https://www." + url
	} else {
		url = "https://" + url
	}

	parsedUrl, err := neturl.Parse(url)
	if err != nil {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "URL Error"+err.Error(), url, fieldName))
		return
	}
	path := parsedUrl.Path
	if url != path {
		url = strings.Trim(url, parsedUrl.Query().Encode())
		url = strings.Trim(url, "?")
		url = strings.Replace(url, path, "", 1)
	}

	descriptionPattern := `^(((http|https?)://)?([\w]{0,3}.)([a-z0-9]?(?:[-a-z0-9]*).){1,2}([a-z]{2,5}){1,1})$`

	regex := regexp.MustCompile(descriptionPattern)
	if regex.MatchString(url) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "url pattern does not match", url, fieldName))
}

func IsNumberComma(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsNumberComma"
	numberList, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", numberList, fieldName))
		return
	}
	if len(numberList) == 0 {
		return
	}

	numberPattern := `[0-9]+(,[0-9]+)*$`

	regex := regexp.MustCompile(numberPattern)
	if regex.MatchString(numberList) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "NumericComma pattern does not match", numberList, fieldName))
}

func IsAlphaDashComma(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsAlphaDashComma"
	DescList, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", DescList, fieldName))
		return
	}
	if len(DescList) == 0 {
		return
	}

	descPattern := `^[0-9a-zA-Z-_]+(,[0-9a-zA-Z-_]+)*$`

	regex := regexp.MustCompile(descPattern)
	if regex.MatchString(DescList) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "AlphaDashComma pattern does not match", DescList, fieldName))
}

func IsAlphaComma(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsAlphaComma"
	DescList, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", DescList, fieldName))
		return
	}
	if len(DescList) == 0 {
		return
	}

	descPattern := `^[0-9a-zA-Z]+(,[0-9a-zA-Z]+)*$`

	regex := regexp.MustCompile(descPattern)
	if regex.MatchString(DescList) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "AlphaComma pattern does not match", DescList, fieldName))
}

func Is24HourTime(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "Is24HourTime"
	time, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", time, fieldName))
		return
	}
	if len(time) == 0 {
		return
	}

	// HH:MM:SS 24-hour format with leading 0
	timePattern := `^(?:[01]\d|2[0-3]):(?:[0-5]\d):(?:[0-5]\d)$`

	regex := regexp.MustCompile(timePattern)
	if regex.MatchString(time) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "time pattern does not match", time, fieldName))
}

/**
 * Please not that this IsvipLevel is only for checking vip level range from -1 - 10
 * Please do being awared before using this func
 */
func IsVipLevel(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsVipLevel"
	if reflect.TypeOf(obj).String() != "int" && reflect.TypeOf(obj).String() != "int64" {
		vipLevelList, ok := obj.(string)
		if !ok {
			_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid string converted value", vipLevelList, fieldName))
			return
		}
		if len(vipLevelList) == 0 {
			return
		}
		list := strings.Split(vipLevelList, ",")
		listMap := make(map[string]bool)
		/**
		 * str == list of the number in string type
		 * num == list of the number in int type
		 */
		for _, str := range list {
			num, ok := toInt(str)
			if !ok {
				_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value", str, fieldName))
				return
			}
			//default -1 represents all vip level, no other values should coexist in the list
			if len(list) > 1 && num == -1 {
				_ = v.SetError(fieldName, buildErrorMsg(funcName, "-1 coexist with other values", str, fieldName))
				return
			}
			//check if @str is a duplicate
			if !listMap[str] {
				listMap[str] = true
			} else {
				_ = v.SetError(fieldName, buildErrorMsg(funcName, "a repeating value", str, fieldName))
				return
			}
			//vip level should be in the range of [0,10] and accepts -1 representing all levels
			if !isInVipRange(num) && num != -1 {
				_ = v.SetError(fieldName, buildErrorMsg(funcName, "not in the valid vip level range", str, fieldName))
				return
			}
		}
	} else {
		getint, _ := toInt(obj)
		if obj == -1 {
			return
		} else if isInVipRange(getint) {
			return
		}
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "not in the valid vip level range. Something goes wrong2", "", fieldName))
	}
}

func IsEditVipLevel(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsEditVipLevel"

	vip, ok := toInt(obj)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid string converted value", obj.(string), fieldName))
		return
	}
	if isInVipRange(vip) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "not in the valid vip level range", obj.(string), fieldName))
}

// isInVipRange checks if the argument num is in the range of [0,10]
func isInVipRange(num int) bool {
	return num >= 0 && num <= 10
}

/**
 * !This is just for a sample purpose
 * Fucntion is not used anymore, Beego validation package does do the validation Size for chinese character
 * Was browsing on wrong function and taken as the MaxSize function does not cvalidation on chinese character
 */
// /**
//  * This is as custom function to check the Maximum size for a string
//  * Reason to use this function instead of original MaxSize() function is because of
//  * 	1. Beego Validation package MaxSize() function does not support word count checking for chinese character (len = 3 for beego validation)
//  */
// func Max5(v *validation.Validation, obj interface{}, key string) {
// 	const size = 5
// 	data, ok := obj.(string)
// 	if !ok {
// 		_ = v.SetError("Max5", "invalid max5 value")
// 		return
// 	}
// 	if len(data) == 0 {
// 		return
// 	}

// 	if err := customMinMax(2, size, data); err == true {
// 		return
// 	}

// 	_ = v.SetError("Max5", "Value size does not match Max5 condition: "+strings.Split(key, ".")[0])
// }

/**
 * @param minOrMax
 * 	- If minOrMax value == 1, checking minimum size of string
 * 	- If minOrMax value == 2, checking maximum size of string
 * @return, true or false value
 */
func customMinMax(minOrMax int, size int, data string) bool {
	result := true
	switch minOrMax {
	case 1:
		{
			if len([]rune(data)) < size {
				logs.Error("[tsValidV2][CustomMinMax] Min condition does not match")
				result = false
			}

		}
	case 2:
		{
			if len([]rune(data)) > size {
				logs.Error("[tsValidV2][CustomMinMax] Max condition does not match")
				result = false
			}

		}
	}
	return result
}

// IsRealName constrains names to contain only alphabets, chinese characters and middle dots
func IsRealName(v *validation.Validation, obj interface{}, key string) {
	fieldName, funcName := strings.Split(key, ".")[0], "IsRealName"
	name, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value:", name, fieldName))
		return
	}
	if len(name) == 0 {
		return
	}
	namePattern := `^$|^[a-zA-Z\p{Han}][a-zA-Z\p{Han}·・*]*$`
	regex := regexp.MustCompile(namePattern)
	if regex.MatchString(name) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "name pattern does not match:", name, fieldName))
}

// IsVersionName is used on channel app area, check suit the versionCode Pattern
func IsVersionName(v *validation.Validation, obj interface{}, key string) {
	// 这个玩意是有小数点的玩意， 例如： 			3.0.0.1
	fieldName, funcName := strings.Split(key, ".")[0], "IsVersionCode"
	name, ok := obj.(string)
	if !ok {
		_ = v.SetError(fieldName, buildErrorMsg(funcName, "invalid value:", name, fieldName))
		return
	}
	if len(name) == 0 {
		return
	}
	namePattern := `^(((\d+)(\.+)){2}(\d+){1}|0)(,(((\d+)(\.+)){2}(\d+){1}|0))*$`
	regex := regexp.MustCompile(namePattern)
	if regex.MatchString(name) {
		return
	}
	_ = v.SetError(fieldName, buildErrorMsg(funcName, "name pattern does not match:", name, fieldName))
}

func buildErrorMsg(funcName, msg, val, field string) string {
	return fmt.Sprintf("%s: %s. Input: %s. Field: %s", funcName, msg, val, field)
}
