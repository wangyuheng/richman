package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/geeklubcn/feishu-bitable-db/db"
	"github.com/wangyuheng/richman/client"
	"github.com/wangyuheng/richman/model"
)

const (
	dreamTable    = "dream"
	dreamKeyword  = "keyword"
	dreamTarget   = "目标"
	dreamCurVal   = "当前值"
	dreamProgress = "进度"
)

const (
	dreamRecordTable   = "dream_record"
	dreamRecordKeyword = "keyword"
	dreamRecordAmount  = "金额"
	dreamRecordDate    = "日期"
	dreamRecordMaker   = "maker"
)

type DreamSvc interface {
	ListDream(appToken string) []*model.Dream
	GetDream(appToken, keyword string) (*model.Dream, bool)
	Save(appToken string, m *model.Dream) (string, error)
	Record(appToken string, m *model.DreamRecord) (*model.Dream, error)
}

func NewDreamSvc(appId, appSecret string, bitable client.Bitable) DreamSvc {
	it, _ := db.NewDB(appId, appSecret)

	return &dreamSvc{db: it, bitable: bitable}
}

type dreamSvc struct {
	db      db.DB
	bitable client.Bitable
	cache   sync.Map
}

func (d *dreamSvc) ListDream(appToken string) []*model.Dream {
	ctx := context.Background()

	rs := d.db.Read(ctx, appToken, dreamTable, []db.SearchCmd{})
	if len(rs) == 0 {
		return nil
	}
	ds := make([]*model.Dream, 0)

	for _, r := range rs {
		res := &model.Dream{
			Keyword: r[dreamKeyword].(string),
		}
		if r[dreamTarget] != nil {
			res.Target, _ = strconv.ParseFloat(r[dreamTarget].(string), 10)
		}

		if r[dreamCurVal] != nil {
			res.CurVal, _ = strconv.ParseFloat(r[dreamCurVal].(string), 10)
		}

		res.Progress = fmt.Sprintf("%.2f%%", res.CurVal*100/res.Target)
		res.Id = r[db.ID].(string)
		ds = append(ds, res)
	}
	return ds
}

func (d *dreamSvc) GetDream(appToken, keyword string) (*model.Dream, bool) {
	ctx := context.Background()

	rs := d.db.Read(ctx, appToken, dreamTable, []db.SearchCmd{
		{dreamKeyword, "=", keyword},
	})
	if len(rs) == 0 {
		return nil, false
	}
	r := rs[0]
	res := &model.Dream{
		Keyword: r[dreamKeyword].(string),
	}

	if r[dreamTarget] != nil {
		res.Target, _ = strconv.ParseFloat(r[dreamTarget].(string), 10)
	}

	if r[dreamCurVal] != nil {
		res.CurVal, _ = strconv.ParseFloat(r[dreamCurVal].(string), 10)
	}

	res.Progress = fmt.Sprintf("%.2f%%", res.CurVal*100/res.Target)
	res.Id = r[db.ID].(string)
	return res, true
}

func (d *dreamSvc) Record(appToken string, m *model.DreamRecord) (*model.Dream, error) {
	ctx := context.Background()
	var err error

	if m.Date == 0 {
		m.Date = time.Now().UnixNano() / 1e6
	}

	fm := d.bitable.ListFields(ctx, appToken, dreamRecordTable)
	author := dreamRecordMaker
	for _, f := range fm {
		if f.Type == 11 {
			author = f.FieldName
		}
	}

	_, err = d.db.Create(ctx, appToken, dreamRecordTable, map[string]interface{}{
		dreamRecordKeyword: m.Keyword,
		dreamRecordAmount:  m.Amount,
		dreamRecordDate:    m.Date,
		author: []map[string]string{
			{
				"id": m.Maker,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var total float64
	for _, r := range d.db.Read(ctx, appToken, dreamRecordTable, []db.SearchCmd{
		{dreamRecordKeyword, "=", m.Keyword},
	}) {
		amount, _ := strconv.ParseFloat(r[dreamRecordAmount].(string), 10)
		total += amount
	}

	old, exists := d.GetDream(appToken, m.Keyword)
	if !exists {
		return nil, fmt.Errorf("keyword [%s] 不存在，可以先创建dream格式为： dream keyword targetVal initVal。 \\r\\n 比如： dream home 100 1\"", m.Keyword)
	}

	res := &model.Dream{
		Id:       old.Id,
		Keyword:  old.Keyword,
		Target:   old.Target,
		CurVal:   total,
		Progress: fmt.Sprintf("%.2f%%", total*100/old.Target),
	}
	err = d.db.Update(ctx, appToken, dreamTable, res.Id, map[string]interface{}{
		dreamKeyword:  res.Keyword,
		dreamTarget:   res.Target,
		dreamCurVal:   res.CurVal,
		dreamProgress: res.Progress,
	})
	return res, err
}

func (d *dreamSvc) Save(appToken string, m *model.Dream) (string, error) {
	ctx := context.Background()

	d.initTable(ctx, appToken)

	for _, r := range d.db.Read(ctx, appToken, dreamTable, []db.SearchCmd{
		{dreamKeyword, "=", m.Keyword},
	}) {
		_ = d.db.Delete(ctx, appToken, dreamTable, db.GetID(r))
	}

	return d.db.Create(ctx, appToken, dreamTable, map[string]interface{}{
		dreamKeyword:  m.Keyword,
		dreamTarget:   m.Target,
		dreamCurVal:   m.CurVal,
		dreamProgress: m.Progress,
	})
}

func (d *dreamSvc) initTable(ctx context.Context, appToken string) {
	if v, ok := d.cache.Load(fmt.Sprintf("dream-%s", appToken)); ok {
		if vv := v.(bool); vv {
			return
		}
	}
	has := false
	for _, t := range d.db.ListTables(ctx, appToken) {
		if t == dreamTable {
			has = true
			break
		}
	}
	if !has {
		_, _ = d.db.SaveTable(ctx, appToken, db.Table{
			Name: dreamTable,
			Fields: []db.Field{
				{Name: dreamKeyword, Type: db.String},
				{Name: dreamTarget, Type: db.Int},
				{Name: dreamCurVal, Type: db.Int},
				{Name: dreamProgress, Type: db.String},
			},
		})

		_, _ = d.db.SaveTable(ctx, appToken, db.Table{
			Name: dreamRecordTable,
			Fields: []db.Field{
				{Name: dreamRecordKeyword, Type: db.String},
				{Name: dreamRecordAmount, Type: db.Int},
				{Name: dreamRecordDate, Type: db.Date},
				{Name: dreamRecordMaker, Type: db.People},
			},
		})
	}
	d.cache.Store(fmt.Sprintf("dream-%s", appToken), true)
}
