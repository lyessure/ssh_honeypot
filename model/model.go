package model

type LoginAttempt struct {
	ID          int    `xorm:"'id' pk autoincr"`
	IP          string `xorm:"'ip'"`
	Username    string `xorm:"'username'"`
	Password    string `xorm:"'password'"`
	ASN         string `xorm:"'asn'"`
	AttemptTime int64  `xorm:"'attempt_time'"`
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

type PasswordCount struct {
	Password string `xorm:"'password'"`
	Count    int    `xorm:"'count'"`
}

func (PasswordCount) TableName() string {
	return "passwordcount"
}

type ASNCount struct {
	ASN   string `xorm:"'asn'"`
	Count int    `xorm:"'count'"`
}

func (ASNCount) TableName() string {
	return "asncount"
}

type IPCount struct {
	IP    string `xorm:"'ip'"`
	Count int    `xorm:"'count'"`
}

func (IPCount) TableName() string {
	return "ipcount"
}
