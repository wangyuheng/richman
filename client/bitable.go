package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/larksuite/oapi-sdk-go/api/core/response"
	"github.com/larksuite/oapi-sdk-go/core"
	"github.com/larksuite/oapi-sdk-go/core/tools"
	larkBitable "github.com/larksuite/oapi-sdk-go/service/bitable/v1"
	"github.com/sirupsen/logrus"
)

type Bitable interface {
	GetApp(ctx *core.Context) (*larkBitable.App, error)
	ListTables(ctx *core.Context) (map[string]*larkBitable.AppTable, error)
	CreateTable(ctx *core.Context, body *larkBitable.AppTableCreateReqBody) (string, error)
	BatchCreateTable(ctx *core.Context, body *larkBitable.AppTableBatchCreateReqBody) ([]string, error)

	ListViews(ctx *core.Context, tableId string) (map[string]*larkBitable.AppTableView, error)
	SyncViews(ctx *core.Context, tableId string, body []*larkBitable.AppTableView) error

	ListFields(ctx *core.Context, tableId string) (map[string]*larkBitable.AppTableField, error)
	CreateField(ctx *core.Context, tableId string, body *larkBitable.AppTableField) (string, error)
	UpdateField(ctx *core.Context, tableId string, body *larkBitable.AppTableField) error

	ListRecords(ctx *core.Context, tableId string, filter string) (map[string]*larkBitable.AppTableRecord, error)
	BatchCreateRecord(ctx *core.Context, tableId string, body *larkBitable.AppTableRecordBatchCreateReqBody) ([]string, error)
	BatchUpdateRecord(ctx *core.Context, tableId string, body *larkBitable.AppTableRecordBatchUpdateReqBody) error

	SyncRecords(ctx *core.Context, tableId string, recordFields []map[string]interface{}) error
}

type bitable struct {
	appToken string
	service  *larkBitable.Service
}

func (b *bitable) SyncRecords(ctx *core.Context, tableId string, recordFields []map[string]interface{}) error {
	records, err := b.ListRecords(ctx, tableId, "")
	if err != nil {
		return err
	}

	existedRecordIDs := make(map[interface{}]*larkBitable.AppTableRecord, 0)
	for _, r := range records {
		if id, ok := r.Fields["ID"]; ok && id != nil {
			existedRecordIDs[id] = r
		}
	}

	prepareCreateRecords := make([]*larkBitable.AppTableRecord, 0)
	prepareUpdateRecords := make([]*larkBitable.AppTableRecord, 0)

	for _, fields := range recordFields {
		if r, ok := existedRecordIDs[fields["ID"]]; ok {
			prepareUpdateRecords = append(prepareUpdateRecords, &larkBitable.AppTableRecord{
				RecordId: r.RecordId,
				Fields:   fields,
			})
		} else {
			prepareCreateRecords = append(prepareCreateRecords, &larkBitable.AppTableRecord{
				Fields: fields,
			})
		}
	}

	if len(prepareCreateRecords) > 0 {
		_, err = b.BatchCreateRecord(ctx, tableId, &larkBitable.AppTableRecordBatchCreateReqBody{
			Records: prepareCreateRecords,
		})
		if err != nil {
			return err
		}
	}
	if len(prepareUpdateRecords) > 0 {
		err = b.BatchUpdateRecord(ctx, tableId, &larkBitable.AppTableRecordBatchUpdateReqBody{
			Records: prepareUpdateRecords,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *bitable) ListRecords(ctx *core.Context, tableId string, filter string) (map[string]*larkBitable.AppTableRecord, error) {
	reqCall := b.service.AppTableRecords.List(ctx)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	reqCall.SetFilter(filter)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("ListRecords fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	res := make(map[string]*larkBitable.AppTableRecord, len(message.Items))
	for _, it := range message.Items {
		res[it.RecordId] = it
	}

	return res, nil
}

func (b *bitable) BatchUpdateRecord(ctx *core.Context, tableId string, body *larkBitable.AppTableRecordBatchUpdateReqBody) error {
	if body == nil || len(body.Records) == 0 {
		return nil
	}
	reqCall := b.service.AppTableRecords.BatchUpdate(ctx, body)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("BatchUpdateRecord fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return nil
}

func (b *bitable) BatchCreateRecord(ctx *core.Context, tableId string, body *larkBitable.AppTableRecordBatchCreateReqBody) ([]string, error) {
	logrus.WithContext(ctx).Debugf("request: %s", tools.Prettify(body))
	if body == nil || len(body.Records) == 0 {
		return nil, nil
	}
	reqCall := b.service.AppTableRecords.BatchCreate(ctx, body)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("BatchCreateRecord fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	recordIds := make([]string, 0)
	for _, r := range message.Records {
		recordIds = append(recordIds, r.RecordId)
	}
	return recordIds, nil
}

func (b *bitable) CreateField(ctx *core.Context, tableId string, body *larkBitable.AppTableField) (string, error) {
	reqCall := b.service.AppTableFields.Create(ctx, body)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("CreateField fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return "", err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return message.Field.FieldId, nil
}

func (b *bitable) UpdateField(ctx *core.Context, tableId string, body *larkBitable.AppTableField) error {
	reqCall := b.service.AppTableFields.Update(ctx, body)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	reqCall.SetFieldId(body.FieldId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).WithFields(logrus.Fields{
			"appToken": b.appToken,
			"tableId":  tableId,
			"body":     body,
			"response": tools.Prettify(message),
		}).Errorf("UpdateField fail!")
		return err
	}
	logrus.WithContext(ctx).Debugf("UpdateField response:%s", tools.Prettify(message))
	return nil
}

func NewBitable(appId string, appSecret string, appToken string) Bitable {
	// 企业自建应用的配置
	// AppID、AppSecret: "开发者后台" -> "凭证与基础信息" -> 应用凭证（App ID、App Secret）
	// EncryptKey、VerificationToken："开发者后台" -> "事件订阅" -> 事件订阅（Encrypt Key、Verification Token）
	// HelpDeskID、HelpDeskToken, 服务台 token：https://open.feishu.cn/document/ukTMukTMukTM/ugDOyYjL4gjM24CO4IjN
	// 更多介绍请看：Github->README.zh.md->如何构建应用配置（AppSettings）
	logrus.Debugf("init lark cli. appId:%s, appSecret:%s, appToken:%s", appId, appSecret, appToken)
	appSettings := core.NewInternalAppSettings(
		core.SetAppCredentials(appId, appSecret), // 必需
	)
	conf := core.NewConfig(core.DomainFeiShu, appSettings, core.SetLoggerLevel(core.LoggerLevelError))
	// 当前访问的是飞书，使用默认的内存存储（app/tenant access token）、默认日志（Error级别）
	// 更多介绍请看：Github->README.zh.md->如何构建整体配置（Config）
	return &bitable{
		appToken: appToken,
		service:  larkBitable.NewService(conf),
	}
}

func (b *bitable) CreateTable(ctx *core.Context, body *larkBitable.AppTableCreateReqBody) (string, error) {
	reqCall := b.service.AppTables.Create(ctx, body)
	reqCall.SetAppToken(b.appToken)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("CreateTable fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return "", err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return message.TableId, nil
}

func (b *bitable) BatchCreateTable(ctx *core.Context, body *larkBitable.AppTableBatchCreateReqBody) ([]string, error) {
	reqCall := b.service.AppTables.BatchCreate(ctx, body)
	reqCall.SetAppToken(b.appToken)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("BatchCreateTable fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return message.TableIds, nil
}

func (b *bitable) SyncViews(ctx *core.Context, tableId string, body []*larkBitable.AppTableView) error {
	views, err := b.listView(ctx, tableId)
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("SyncViews fail! appToken:%s,tableId:%s", b.appToken, tableId)
		return err
	}
	tmp, _ := b.createView(ctx, tableId, &larkBitable.AppTableView{
		ViewName: uuid.New().String(),
	})

	defer func() {
		_ = b.deleteView(ctx, tableId, tmp.ViewId)
	}()
	if err = b.batchDeleteViews(ctx, tableId, views); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("SyncViews fail! appToken:%s,tableId:%s", b.appToken, tableId)
		return err
	}
	if err := b.batchCreateView(ctx, tableId, body); err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("SyncViews fail! appToken:%s,tableId:%s", b.appToken, tableId)
		return err
	}
	return nil
}

func (b *bitable) deleteView(ctx *core.Context, tableId string, viewId string) error {
	reqCall := b.service.AppTableViews.Delete(ctx)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	reqCall.SetViewId(viewId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("deleteView fail! appToken:%s,tableId:%s,viewId:%s,response:%s", b.appToken, tableId, viewId, tools.Prettify(message))
		return err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return nil
}

func (b *bitable) batchDeleteViews(ctx *core.Context, tableId string, views []*larkBitable.AppTableView) error {
	for _, v := range views {
		if err := b.deleteView(ctx, tableId, v.ViewId); err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("batchDeleteViews fail! appToken:%s,tableId:%s", b.appToken, tableId)
			return err
		}
	}
	return nil
}

func (b *bitable) listView(ctx *core.Context, tableId string) ([]*larkBitable.AppTableView, error) {
	reqCall := b.service.AppTableViews.List(ctx)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("CreateView fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return message.Items, nil
}

func (b *bitable) batchCreateView(ctx *core.Context, tableId string, views []*larkBitable.AppTableView) error {
	for _, v := range views {
		if _, err := b.createView(ctx, tableId, v); err != nil {
			logrus.WithContext(ctx).WithError(err).Errorf("BatchCreateView fail! appToken:%s,tableId:%s", b.appToken, tableId)
			return err
		}
	}
	return nil
}

func (b *bitable) createView(ctx *core.Context, tableId string, body *larkBitable.AppTableView) (*larkBitable.AppTableView, error) {
	reqCall := b.service.AppTableViews.Create(core.WrapContext(context.Background()), body)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("CreateView fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	// nil res is shit
	views, _ := b.listView(ctx, tableId)
	for _, v := range views {
		if v.ViewName == body.ViewName {
			return v, nil
		}
	}
	return message.AppTableView, nil
}

func (b *bitable) GetApp(ctx *core.Context) (*larkBitable.App, error) {
	reqCall := b.service.Apps.Get(ctx)
	reqCall.SetAppToken(b.appToken)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("GetApp fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	return message.App, nil
}

func (b *bitable) ListTables(ctx *core.Context) (map[string]*larkBitable.AppTable, error) {
	reqCall := b.service.AppTables.List(ctx)
	reqCall.SetAppToken(b.appToken)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("ListTables fail! appToken:%s,response:%s", b.appToken, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	res := make(map[string]*larkBitable.AppTable, len(message.Items))
	for _, it := range message.Items {
		res[it.Name] = it
	}
	return res, nil
}

func (b *bitable) ListFields(ctx *core.Context, tableId string) (map[string]*larkBitable.AppTableField, error) {
	reqCall := b.service.AppTableFields.List(ctx)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("ListFields fail! appToken:%s,tableId:%s,response:%s", b.appToken, tableId, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	res := make(map[string]*larkBitable.AppTableField, len(message.Items))
	for _, it := range message.Items {
		res[it.FieldName] = it
	}
	return res, nil
}

func (b *bitable) ListViews(ctx *core.Context, tableId string) (map[string]*larkBitable.AppTableView, error) {
	reqCall := b.service.AppTableViews.List(ctx)
	reqCall.SetAppToken(b.appToken)
	reqCall.SetTableId(tableId)
	message, err := reqCall.Do()
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("ListViews fail! appToken:%s,tableId:%s,response:%s", b.appToken, tableId, tools.Prettify(message))
		return nil, err
	}
	logrus.WithContext(ctx).Debugf("response:%s", tools.Prettify(message))
	res := make(map[string]*larkBitable.AppTableView, len(message.Items))
	for _, it := range message.Items {
		res[it.ViewName] = it
	}
	return res, nil
}

func (b *bitable) printErr(ctx *core.Context, err error) {
	logrus.WithContext(ctx).Infof("reqId:%s, status:%d", ctx.GetRequestID(), ctx.GetHTTPStatusCode())
	if err != nil {
		if e, ok := err.(*response.Error); ok {
			logrus.WithContext(ctx).WithField("code", e.Code)
			logrus.WithContext(ctx).WithField("msg", e.Msg)
			logrus.WithContext(ctx).WithField("appToken", b.appToken)
		}
		logrus.WithContext(ctx).WithField("err", tools.Prettify(err)).Error("request call fail!")
		return
	}
}
