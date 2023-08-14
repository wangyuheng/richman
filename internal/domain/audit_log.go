package domain

type AuditLog struct {
	Req, Resp, Key, Operator string
}
