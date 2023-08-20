package business

import (
	"context"
	"encoding/json"
	"fmt"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/model"
	"time"
)

const (
	BillTableName = "个人账单记录"
)

type LedgerSvr interface {
	Generate(ctx context.Context, creator model.User) (*Ledger, error)
	QueryByUID(ctx context.Context, UID string) (*Ledger, bool)
}

type ledgerSvr struct {
	cli                  *lark.Client
	dbAppToken           string
	dbTableToken         string
	templateAppToken     string
	targetFolderAppToken string
}

func NewLedgerSvr(cli *lark.Client, cfg config.LarkDBConfig) LedgerSvr {
	return &ledgerSvr{
		cli:                  cli,
		dbAppToken:           cfg.DBAppToken,
		dbTableToken:         cfg.DBTableToken,
		templateAppToken:     cfg.TemplateAppToken,
		targetFolderAppToken: cfg.TargetFolderAppToken,
	}
}

func (l *ledgerSvr) QueryByUID(ctx context.Context, UID string) (*Ledger, bool) {
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
	var ledger Ledger
	j, err := json.Marshal(resp.Data.Items[0].Fields)
	if err != nil {
		return nil, false
	}
	err = json.Unmarshal(j, &ledger)
	if err != nil {
		return nil, false
	}
	return &ledger, true

}
func (l *ledgerSvr) Generate(ctx context.Context, creator model.User) (*Ledger, error) {
	// 根据模板复制到指定文件夹
	copyFile, err := l.copyFromTemplate(ctx)
	if err != nil {
		return nil, err
	}
	// copying 时调用其他接口可能导致失败
	time.Sleep(2 * time.Second)
	// 允许外部访问
	err = l.setPermissionPublic(ctx, *copyFile.Token)
	if err != nil {
		return nil, err
	}
	// 获取账单表ID
	tableID, err := l.getBillTableID(ctx, *copyFile.Token)
	if err != nil {
		return nil, err
	}
	// 保存账本
	ledger := &Ledger{
		AppToken:    *copyFile.Token,
		TableToken:  *tableID,
		Name:        *copyFile.Name,
		URL:         *copyFile.Url,
		CreatorID:   creator.UID,
		CreatorName: creator.Name,
	}
	_ = l.save(ctx, ledger)
	return ledger, nil
}

func (l *ledgerSvr) save(ctx context.Context, it *Ledger) error {
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

func (l *ledgerSvr) copyFromTemplate(ctx context.Context) (*larkdrive.File, error) {
	resp, err := l.cli.Drive.File.Copy(ctx, larkdrive.NewCopyFileReqBuilder().
		FileToken(l.templateAppToken).
		Body(larkdrive.NewCopyFileReqBodyBuilder().
			Type("bitable").
			FolderToken(l.targetFolderAppToken).
			Name("飞书记账").
			Build()).
		Build())
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Errorf("copy file error! resp:%+v", resp)
		return nil, err
	}
	if !resp.Success() {
		logrus.WithContext(ctx).WithError(err).Errorf("copy file fail! resp:%+v", resp)
		return nil, fmt.Errorf(resp.Msg)
	}
	return resp.Data.File, nil
}

func (l *ledgerSvr) setPermissionPublic(ctx context.Context, appToken string) error {
	permissionResp, err := l.cli.Drive.PermissionPublic.Patch(ctx, larkdrive.NewPatchPermissionPublicReqBuilder().
		Token(appToken).
		Type("bitable").
		PermissionPublicRequest(larkdrive.NewPermissionPublicRequestBuilder().
			ExternalAccess(true).
			SecurityEntity("anyone_can_view").
			CommentEntity("anyone_can_view").
			ShareEntity("anyone").
			LinkShareEntity("anyone_editable").
			InviteExternal(true).
			Build()).
		Build())
	if err != nil || !permissionResp.Success() {
		logrus.WithContext(ctx).WithError(err).Errorf("patch file permission fail! resp:%+v", permissionResp)
		return err
	}
	return nil
}

func (l *ledgerSvr) getBillTableID(ctx context.Context, appToken string) (*string, error) {
	logger := logrus.WithContext(ctx)
	req := larkbitable.NewListAppTableReqBuilder().
		AppToken(appToken).
		PageSize(20).
		Build()

	table, err := l.cli.Bitable.AppTable.List(ctx, req)
	if err != nil {
		logger.WithError(err).Errorf("list app table error! req: %+v, resp:%+v", req, table)
		return nil, err
	}
	if !table.Success() {
		logger.WithError(err).Errorf("list app table fail! req: %+v, resp:%+v", req, table)
		return nil, fmt.Errorf(table.Msg)
	}
	for _, t := range table.Data.Items {
		if *t.Name == BillTableName {
			return t.TableId, nil
		}
	}
	return nil, nil
}

type Ledger struct {
	AppToken    string `json:"app_token"`
	TableToken  string `json:"table_token"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	CreatorID   string `json:"creator_id"`
	CreatorName string `json:"creator_name"`
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
