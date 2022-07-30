package client

import (
	"context"

	"github.com/larksuite/oapi-sdk-go/core"
	"github.com/larksuite/oapi-sdk-go/core/config"
	"github.com/larksuite/oapi-sdk-go/core/tools"
	larkBitable "github.com/larksuite/oapi-sdk-go/service/bitable/v1"
	"github.com/sirupsen/logrus"
)

type Bitable interface {
	ListTables(ctx context.Context, appToken string) map[string]*larkBitable.AppTable
	ListFields(ctx context.Context, appToken, tableName string) map[string]*larkBitable.AppTableField
}

type bitable struct {
	s *larkBitable.Service
}

func NewBitable(cfg *config.Config) Bitable {
	return &bitable{s: larkBitable.NewService(cfg)}
}

func (b bitable) ListTables(ctx context.Context, appToken string) map[string]*larkBitable.AppTable {
	c := core.WrapContext(ctx)
	res := make(map[string]*larkBitable.AppTable, 0)

	reqCall := b.s.AppTables.List(c)
	reqCall.SetAppToken(appToken)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(c).WithError(err).Errorf("ListTables fail! appToken:%s,response:%s", appToken, tools.Prettify(message))
		return res
	}
	logrus.WithContext(c).Debugf("response:%s", tools.Prettify(message))
	for _, it := range message.Items {
		res[it.Name] = it
	}
	return res
}

func (b bitable) ListFields(ctx context.Context, appToken, tableName string) map[string]*larkBitable.AppTableField {
	c := core.WrapContext(ctx)
	res := make(map[string]*larkBitable.AppTableField, 0)

	tableId := b.ListTables(ctx, appToken)[tableName].TableId

	reqCall := b.s.AppTableFields.List(c)
	reqCall.SetAppToken(appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(c).WithError(err).Errorf("ListFields fail! appToken:%s,tableId:%s,response:%s", appToken, tableId, tools.Prettify(message))
		return res
	}
	logrus.WithContext(c).Debugf("response:%s", tools.Prettify(message))
	for _, it := range message.Items {
		res[it.FieldName] = it
	}
	return res
}
