package database

import (
	"context"
	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/domain"
	"sync"
)

var (
	once sync.Once
)

type auditLogService struct {
	db         db.DB
	buf        chan domain.AuditLog
	dbToken    string
	tableToken string
}

func NewAuditLogService(cfg *config.Config, db db.DB) domain.AuditLogService {
	s := &auditLogService{
		db:         db,
		buf:        make(chan domain.AuditLog, 100),
		dbToken:    cfg.AuditLogDBToken,
		tableToken: cfg.AuditLogTableToken,
	}
	once.Do(func() {
		s.StartConsume()
	})
	return s
}

func (a *auditLogService) Send(log domain.AuditLog) {
	select {
	case a.buf <- log:
		return
	}
}

func (a *auditLogService) StartConsume() {
	go func() {
		defer func() {
			if e := recover(); e != nil {
				logrus.Errorf("consume fail! err: %v", e)
			}
		}()
		for {
			select {
			case data := <-a.buf:
				_, _ = a.db.Create(context.Background(), a.dbToken, a.tableToken, map[string]interface{}{
					"req":      data.Req,
					"resp":     data.Resp,
					"key":      data.Key,
					"operator": data.Operator,
				})
			}
		}
	}()
}
