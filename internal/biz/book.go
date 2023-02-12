package biz

import (
	"context"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/sirupsen/logrus"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/model"
	"github.com/wangyuheng/richman/internal/repo"
)

type Book interface {
	Generate(ctx context.Context, creator model.User) (*model.Book, error)
	Bind(ctx context.Context, url string, creator model.User) error
}

type book struct {
	templateToken string
	folderToken   string
	books         repo.Books
	cli           *lark.Client
}

func NewBook(cfg *config.Config, books repo.Books, cli *lark.Client) Book {
	return &book{
		templateToken: cfg.TemplateAppToken,
		folderToken:   cfg.TargetFolderAppToken,
		books:         books,
		cli:           cli,
	}
}

func (b book) Bind(ctx context.Context, url string, creator model.User) error {
	return nil
}

func (b book) Generate(ctx context.Context, creator model.User) (*model.Book, error) {
	logger := logrus.WithContext(ctx).WithField("creator", creator)
	larkCtx := context.Background()
	// 根据模板复制到指定文件夹
	resp, err := b.cli.Drive.File.Copy(larkCtx, larkdrive.NewCopyFileReqBuilder().
		FileToken(b.templateToken).
		Body(larkdrive.NewCopyFileReqBodyBuilder().
			Type("bitable").
			FolderToken(b.folderToken).
			Name("飞书记账").
			Build()).
		Build())
	if err != nil || !resp.Success() {
		logger.WithError(err).Errorf("copy file fail! resp:%+v", resp)
		return nil, err
	}
	// 允许外部访问
	permissionResp, err := b.cli.Drive.PermissionPublic.Patch(larkCtx, larkdrive.NewPatchPermissionPublicReqBuilder().
		Token(*resp.Data.File.Token).
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
		logger.WithError(err).Errorf("patch file permission fail! resp:%+v", permissionResp)
		return nil, err
	}
	// 持久化存储
	it := &model.Book{
		AppToken:    *resp.Data.File.Token,
		Name:        *resp.Data.File.Name,
		URL:         *resp.Data.File.Url,
		CreatorID:   creator.UID,
		CreatorName: creator.Name,
	}

	if _, err = b.books.Save(it); err != nil {
		logger.WithError(err).Error("save book err!")
		return nil, err
	}

	return it, nil
}
