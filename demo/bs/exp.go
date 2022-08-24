package bs

import (
	"context"
	"fmt"
	"github.com/hylent/sf/clients"
	"github.com/hylent/sf/demo/ds"
)

var (
	ExpInstance = new(Exp)
)

type Exp struct {
}

type ExpGetRequest struct {
	Id int64 `json:"id" form:"id" binding:"required"`
}

type ExpGetResponse struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (x *Exp) Get(ctx context.Context, req *ExpGetRequest, resp *ExpGetResponse) error {
	ret := &struct {
		Id   int64  `db:"id"`
		Name string `db:"name"`
	}{}

	if err := ds.Db.First("tb_exp", ret, "", "id=?", req.Id); err != nil {
		return err
	}

	resp.Id = ret.Id
	resp.Name = ret.Name

	return nil
}

func (x *Exp) Post(ctx context.Context, req *ExpGetRequest, resp *ExpGetResponse) error {
	pred := clients.EsPredicate{
		Op:    clients.EsPredicateOpTerm,
		Key:   "_id",
		Value: fmt.Sprintf("%d", req.Id),
	}

	query, queryErr := pred.ToQuery()
	if queryErr != nil {
		return EInvalidParam
	}

	q := map[string]any{
		"query": query,
	}

	hitList, err := ds.Es.Search(ctx, "tb_exp", q)

	if err != nil {
		return err
	}

	if len(hitList.Hits) < 1 {
		return nil
	}

	resp.Id = hitList.Hits[0].Source["id"].(int64)
	resp.Name = hitList.Hits[0].Source["name"].(string)

	return nil
}
