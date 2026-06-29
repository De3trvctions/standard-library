package utility

import (
	"fmt"
	"standard-library/consts"
	"standard-library/jwt"
	"standard-library/models/dto"
	"standard-library/nacos"
	"standard-library/redis"
	"strconv"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

func GetRedisLoginStatus(username string) (ableLogin bool, remaindingTime int) {
	ableLogin = true
	ex1, _ := redis.Exists(fmt.Sprintf(consts.FailLoginAccountLock, username))

	if ex1 {
		ableLogin = false
		timeCache, _ := redis.Get(fmt.Sprintf(consts.FailLoginAccountLockTime, username))
		redisTime, _ := strconv.ParseInt(timeCache, 10, 64)
		timeLeft := time.Unix(redisTime, 0)
		remaindingTime = int(time.Until(timeLeft).Seconds())
	}
	return
}

func SetRedisLoginFail(username string) (err error) {
	failCount := 0
	ex, _ := redis.Exists(fmt.Sprintf(consts.FailLoginCount, username))
	if ex {
		count, _ := redis.Get(fmt.Sprintf(consts.FailLoginCount, username))
		failCount, _ = strconv.Atoi(count)
		_, _ = redis.Del(fmt.Sprintf(consts.FailLoginCount, username))
	}

	if failCount >= 5 {
		_ = redis.Set(
			fmt.Sprintf(consts.FailLoginCount, username),
			failCount,
			time.Duration(15)*time.Minute,
		)
		_ = redis.Set(
			fmt.Sprintf(consts.FailLoginAccountLock, username),
			1,
			time.Duration(15)*time.Minute,
		)
		_ = redis.Set(
			fmt.Sprintf(consts.FailLoginAccountLockTime, username),
			time.Now().Add(time.Duration(15)*time.Minute).Unix(),
			time.Duration(15)*time.Minute,
		)
	} else {
		_ = redis.Set(fmt.Sprintf(consts.FailLoginCount, username), failCount)
	}
	return
}

func DelRedisLoginFail(username string) {
	_, _ = redis.Del(fmt.Sprintf(consts.FailLoginCount, username))
	_, _ = redis.Del(fmt.Sprintf(consts.FailLoginAccountLock, username))
	_, _ = redis.Del(fmt.Sprintf(consts.FailLoginAccountLockTime, username))
}

func GetToken(req dto.ReqLogin, id int64) (accessToken, refreshToken string) {
	DelToken(req.Username)
	accessToken = jwt.Gen(map[string]any{
		"Username":  req.Username,
		"AccountId": id,
		"TokenType": consts.ACCESS_TOKEN,
	}, nacos.TokenSalt, time.Duration(nacos.TokenExpMinute)*time.Minute)
	refreshToken = jwt.Gen(map[string]any{
		"Username":  req.Username,
		"AccountId": id,
		"TokenType": consts.REFRESH_TOKEN,
	}, nacos.TokenSalt, time.Duration(nacos.TokenRefreshExpMinute)*time.Minute)
	SetToken(accessToken, req.Username, consts.ACCESS_TOKEN)
	SetToken(refreshToken, req.Username, consts.REFRESH_TOKEN)
	return
}

func SetToken(token, username, tokenType string) {
	var duration time.Duration
	var key1, key2 string
	switch tokenType {
	case consts.ACCESS_TOKEN:
		key1 = fmt.Sprintf(consts.AccountLoginByToken, token)
		key2 = fmt.Sprintf(consts.AccountLoginByUsername, username)
		duration = time.Duration(nacos.TokenExpMinute) * time.Minute
	case consts.REFRESH_TOKEN:
		key1 = fmt.Sprintf(consts.RefreshAccountLoginByToken, token)
		key2 = fmt.Sprintf(consts.RefreshAccountLoginByUsername, username)
		duration = time.Duration(nacos.TokenRefreshExpMinute) * time.Minute
	default:
		logs.Error("[SetToken] Unknown token type:", tokenType)
		return
	}

	// Set Redis token
	_ = redis.Set(key1, username, duration)
	_ = redis.Set(key2, token, duration)
}

func DelToken(username string) {
	ex, _ := redis.Exists(fmt.Sprintf(consts.AccountLoginByUsername, username))
	if ex {
		token, _ := redis.Get(fmt.Sprintf(consts.AccountLoginByUsername, username))
		_, err1 := redis.Del(fmt.Sprintf(consts.AccountLoginByToken, token))
		_, err2 := redis.Del(fmt.Sprintf(consts.AccountLoginByUsername, username))
		if err1 != nil {
			logs.Error("[DelToken] Error 1: ", err1)
		}
		if err2 != nil {
			logs.Error("[DelToken] Error 2: ", err2)
		}
	}

	ex, _ = redis.Exists(fmt.Sprintf(consts.RefreshAccountLoginByUsername, username))
	if ex {
		token, _ := redis.Get(fmt.Sprintf(consts.RefreshAccountLoginByUsername, username))
		_, err1 := redis.Del(fmt.Sprintf(consts.RefreshAccountLoginByToken, token))
		_, err2 := redis.Del(fmt.Sprintf(consts.RefreshAccountLoginByUsername, username))
		if err1 != nil {
			logs.Error("[DelToken] Error 3: ", err1)
		}
		if err2 != nil {
			logs.Error("[DelToken] Error 4: ", err2)
		}
	}
}
