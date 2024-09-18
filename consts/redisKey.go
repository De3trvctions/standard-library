package consts

const (
	FailLoginCount           = "FailLoginCount:%s"
	FailLoginAccountLock     = "FailLoginAccountLock:%s"
	FailLoginAccountLockTime = "FailLoginAccountLockTime:%s"

	AccountLoginByToken    = "AccountLoginByToken:%v"
	AccountLoginByUsername = "AccountLoginByUsername:%v"

	RegisterEmailValidCode           = "RegisterEmailValidCode:%s"
	RegisterEmailValidCodeLock       = "RegisterEmailValidCodeLock:%v"
	ForgetPasswordEmailValidCode     = "ForgetPasswordEmailValidCode:%s"
	ForgetPasswordEmailValidCodeLock = "ForgetPasswordEmailValidCodeLock:%v"
)

// Daily Report Key
const (
	LoginLogTaskLock       = "Report-Login-Log-TaskLock"
	LoginLogRepairTaskLock = "Report-Login-Log-Repair-TaskLock"
	LoginLogPrefix         = "Report-Login-Log-Prefix"
)
