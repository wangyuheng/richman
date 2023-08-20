package domain

type AuditLogService interface {
	Send(log AuditLog)
	StartConsume()
}
