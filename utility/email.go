package utility

import (
	"api-login/consts"
	"api-login/mail"
	"api-login/redis"
	"bytes"
	"fmt"
	"text/template"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

func SendMail(redisKey, redisKeyLock, addr, title, msg string, expTime int64) (code string, timeInterval int64, errCode int64, err error) {
	code = GetRandomNumber(6)
	timenow := int64(0)
	ex, _ := redis.Exists(fmt.Sprintf(redisKeyLock, addr))
	if ex {
		returnTime, _ := redis.Get(fmt.Sprintf(redisKeyLock, addr))
		errCode = consts.VALID_CODE_EXIST
		timeInterval = StringToInt64(returnTime)
		return "", timeInterval, errCode, nil
	}

	err = redis.Set(fmt.Sprintf(redisKey, addr), code, time.Duration(expTime)*time.Minute)
	if err != nil {
		logs.Error(err)
	}

	timenow = time.Now().Add(30 * time.Second).Unix()
	if err := redis.Set(fmt.Sprintf(redisKeyLock, addr), timenow, time.Duration(30)*time.Second); err != nil {
		logs.Error("[SendMail]Set Redis Key RegisterEmailValidCodeLock Error:", err, addr)
		errCode = consts.VALID_CODE_COOL_DOWN

		return "", 0, errCode, nil
	}

	message := generateValidCodeEmailTemplate(addr, title, msg, code)
	cli := mail.Cli()
	err = cli.Send(cli.Address(), []string{addr}, []byte(message))
	if err != nil {
		logs.Error("[SendMail] Send email error", err)
		DelEmailValidCodeLock(addr)
		return "", 0, 0, err
	}
	return
}

func generateValidCodeEmailTemplate(addr, title, msg, code string) (res string) {
	type EmailData struct {
		Title string
		Body  string
	}
	// Create an instance of EmailData with the data for the template
	data := EmailData{
		Title: title,
		Body: `<p><span style="font-size:20px"><strong>` + title + `</strong></span></p>
		<p><span style="color:#3498db"><span style="font-size:28px">Validation Code` + code + `</span></span></p>
		<p>This Validation Code is availavle for 15 min.</p>
		<p>` + msg + `</p>
		<p>&nbsp;</p>
		<p>Thank you</p>
		`,
	}

	// Create a buffer to store the rendered HTML content
	var tplBuffer bytes.Buffer

	// Define your HTML template
	htmlTemplate := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>{{.Title}}</title>
	</head>
	<body>
		{{.Body}}
	</body>
	</html>
	`

	// Parse the HTML template
	tmpl, err := template.New("emailTemplate").Parse(htmlTemplate)
	if err != nil {
		panic(err)
	}

	// Execute the template and write the output to the buffer
	err = tmpl.Execute(&tplBuffer, data)
	if err != nil {
		panic(err)
	}

	res = "From: no-reply \n" +
		"To: " + addr + "\n" +
		"Subject: " + title + "\n" +
		"Content-Type: text/html; charset=UTF-8\n\n" +
		tplBuffer.String()
	return
}

func DelEmailValidCodeLock(addr string) {
	_, _ = redis.Del(fmt.Sprintf(consts.RegisterEmailValidCode, addr))
}
