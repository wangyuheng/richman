package business

import (
	"context"
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/wangyuheng/richman/config"
	"github.com/wangyuheng/richman/internal/client"
	"github.com/wangyuheng/richman/internal/model"
	"testing"
)

func TestGenerate(t *testing.T) {
	cli := client.NewFeishu(config.LarkConfig{
		DbAppId:              "",
		DbAppSecret:          "",
		TemplateAppToken:     "",
		TargetFolderAppToken: "",
	})
	svr := NewLedgerSvr(cli, config.LarkDBConfig{
		DBAppToken:           "",
		DBTableToken:         "",
		TemplateAppToken:     "",
		TargetFolderAppToken: "",
	})
	got, err := svr.Generate(context.Background(), model.User{
		UID:  "a",
		Name: "b",
	})
	println(structToMap(got))
	println(err)
	res, exist := svr.QueryByUID(context.Background(), "a")
	println(fmt.Sprintf("res=%+v", res))
	assert.Equal(t, exist, true)
}
