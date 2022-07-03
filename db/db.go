package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/geeklubcn/richman/client"
	"github.com/larksuite/oapi-sdk-go/core"
	larkBitable "github.com/larksuite/oapi-sdk-go/service/bitable/v1"
)

type SearchCmd struct {
	Key, Operator, Val string
}

type Tables interface {
	List() map[string]*larkBitable.AppTable
	Get(name string) (*larkBitable.AppTable, bool)
	Fields(tableId string) map[string]*larkBitable.AppTableField
}

type tables struct {
	service client.Bitable
	cache   sync.Map
}

var instance Tables

func newDB(service client.Bitable) Tables {
	if instance == nil {
		instance = &tables{service: service}
		go warmup()
	}
	return instance
}

func warmup() {
	for _, t := range instance.List() {
		_ = instance.Fields(t.TableId)
	}
}

func (t *tables) Fields(tableId string) map[string]*larkBitable.AppTableField {
	if v, ok := t.cache.Load(fmt.Sprintf("fields-%s", tableId)); ok {
		if vv, ok := v.(map[string]*larkBitable.AppTableField); ok {
			return vv
		}
	}
	ctx := core.WrapContext(context.Background())
	res, _ := t.service.ListFields(ctx, tableId)
	if len(res) > 0 {
		t.cache.Store(fmt.Sprintf("fields-%s", tableId), res)
	}
	return res
}
func (t *tables) List() map[string]*larkBitable.AppTable {
	if v, ok := t.cache.Load("tables"); ok {
		if vv, ok := v.(map[string]*larkBitable.AppTable); ok {
			return vv
		}
	}
	ctx := core.WrapContext(context.Background())
	res, _ := t.service.ListTables(ctx)
	if len(res) > 0 {
		t.cache.Store("tables", res)
	}
	return res
}
func (t *tables) Get(name string) (*larkBitable.AppTable, bool) {
	for _, it := range t.List() {
		if it.Name == name {
			return it, true
		}
	}
	return nil, false
}
