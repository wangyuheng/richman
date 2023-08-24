package database

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/geeklubcn/feishu-bitable-db/db"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/domain"
	"sync"
)

type ledgerRepository struct {
	db           db.DB
	cli          *lark.Client
	dbAppToken   string
	dbTableToken string
	cache        sync.Map
}

func NewLedgerRepository(cfg *config.Config, cli *lark.Client, db db.DB) domain.LedgerRepository {
	l := &ledgerRepository{
		cli:          cli,
		dbAppToken:   cfg.DBAppToken,
		dbTableToken: cfg.DBTableToken,
		db:           db,
	}
	l.WarmUP(context.Background())
	return l
}

func (l *ledgerRepository) WarmUP(ctx context.Context) {
	items := l.db.Read(ctx, l.dbAppToken, l.dbTableToken, []db.SearchCmd{})
	for _, it := range items {
		res := &domain.Ledger{
			AppToken:    db.GetString(it, "app_token"),
			TableToken:  db.GetString(it, "table_token"),
			Name:        db.GetString(it, "name"),
			URL:         db.GetString(it, "url"),
			CreatorID:   db.GetString(it, "creator_id"),
			CreatorName: db.GetString(it, "creator_name"),
		}
		l.cache.Store(fmt.Sprintf("REPO:LEDGER:%s", db.GetString(it, res.CreatorID)), res)
	}
}

func (l *ledgerRepository) QueryByUID(UID string) (*domain.Ledger, bool) {
	ctx := context.Background()

	if v, ok := l.cache.Load(fmt.Sprintf("REPO:LEDGER:%s", UID)); ok {
		if vv, ok := v.(*domain.Ledger); ok {
			return vv, true
		}
	}

	req := larkbitable.NewListAppTableRecordReqBuilder().
		AppToken(l.dbAppToken).
		TableId(l.dbTableToken).
		Filter(fmt.Sprintf("AND(CurrentValue.creator_id=\"%s\")", UID)).
		Build()
	resp, err := l.cli.Bitable.AppTableRecord.List(ctx, req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("create record err! resp:%+v", resp)
		return nil, false
	}
	if !resp.Success() {
		logrus.WithContext(ctx).WithError(err).Errorf("create record fail! resp:%+v", resp)
		return nil, false
	}
	if *resp.Data.Total == 0 {
		return nil, false
	}
	var ledger domain.Ledger
	j, err := json.Marshal(resp.Data.Items[0].Fields)
	if err != nil {
		return nil, false
	}
	err = json.Unmarshal(j, &ledger)
	if err != nil {
		return nil, false
	}
	l.cache.Store(fmt.Sprintf("REPO:LEDGER:%s", UID), &ledger)

	return &ledger, true
}

func (l *ledgerRepository) Save(it *domain.Ledger) error {
	ctx := context.Background()
	fields, err := structToMap(it)
	if err != nil {
		return err
	}
	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(l.dbAppToken).
		TableId(l.dbTableToken).
		AppTableRecord(larkbitable.NewAppTableRecordBuilder().
			Fields(fields).
			Build()).
		Build()
	resp, err := l.cli.Bitable.AppTableRecord.Create(ctx, req)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("create record err! resp:%+v", resp)
		return err
	}
	if !resp.Success() {
		logrus.WithContext(ctx).WithError(err).Errorf("create record fail! resp:%+v", resp)
		return fmt.Errorf(resp.Msg)
	}
	return nil
}

func structToMap(input interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
