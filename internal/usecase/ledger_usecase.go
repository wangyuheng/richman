package usecase

import (
	"context"
	"fmt"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/domain"
	"time"
)

type LedgerUseCase interface {
	Generate(creator domain.User) (*domain.Ledger, error)
	QueryByUID(UID string) (*domain.Ledger, bool)
}

type ledgerUseCase struct {
	ledgerRepository     domain.LedgerRepository
	cli                  *lark.Client
	templateAppToken     string
	targetFolderAppToken string
}

func NewLedgerUseCase(cfg *config.Config, ledgerRepository domain.LedgerRepository, cli *lark.Client) LedgerUseCase {
	return &ledgerUseCase{
		ledgerRepository:     ledgerRepository,
		cli:                  cli,
		templateAppToken:     cfg.LarkDBConfig.TemplateAppToken,
		targetFolderAppToken: cfg.LarkDBConfig.TargetFolderAppToken,
	}
}

func (l *ledgerUseCase) Generate(creator domain.User) (*domain.Ledger, error) {
	ctx := context.Background()
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
	ledger := &domain.Ledger{
		AppToken:    *copyFile.Token,
		TableToken:  *tableID,
		Name:        *copyFile.Name,
		URL:         *copyFile.Url,
		CreatorID:   creator.UID,
		CreatorName: creator.Name,
	}
	_ = l.ledgerRepository.Save(ledger)
	return ledger, nil
}

func (l *ledgerUseCase) QueryByUID(UID string) (*domain.Ledger, bool) {
	return l.ledgerRepository.QueryByUID(UID)

}

func (l *ledgerUseCase) copyFromTemplate(ctx context.Context) (*larkdrive.File, error) {
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

func (l *ledgerUseCase) setPermissionPublic(ctx context.Context, appToken string) error {
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

func (l *ledgerUseCase) getBillTableID(ctx context.Context, appToken string) (*string, error) {
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
		if *t.Name == "个人账单记录" {
			return t.TableId, nil
		}
	}
	return nil, nil
}
