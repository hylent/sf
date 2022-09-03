package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/hylent/sf/logger"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type EsClient struct {
	Url          string `yaml:"url"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	TimeoutMilli int64  `yaml:"timeout_milli"`

	client *elasticsearch.Client
}

func (x *EsClient) Init(ctx context.Context) error {
	c := elasticsearch.Config{
		Addresses: []string{x.Url},
		Username:  x.Username,
		Password:  x.Password,
	}
	client, clientErr := elasticsearch.NewClient(c)
	if clientErr != nil {
		return fmt.Errorf("es_client_fail: err=%v", clientErr)
	}

	ctx2, cancelFunc := context.WithTimeout(ctx, time.Millisecond*time.Duration(x.TimeoutMilli))
	defer cancelFunc()
	infoResp, infoErr := client.Info(
		client.Info.WithContext(ctx2),
	)
	if infoErr != nil || infoResp == nil || infoResp.StatusCode != http.StatusOK {
		return fmt.Errorf("es_info_fail: infoResp=%+v err=%v", infoResp, infoErr)
	}
	info := &struct {
		ClusterName string `json:"cluster_name"`
		Version     struct {
			Number string `json:"number"`
		} `json:"version"`
	}{}
	if readErr := x.readRespBody(infoResp.Body, info); readErr != nil {
		return fmt.Errorf("es_info_read_fail: err=%v", readErr)
	}

	log.Info("es_connected", logger.M{
		"url":  x.Url,
		"info": info,
	})

	x.client = client
	return nil
}

type EsHitItem struct {
	Index  string                 `json:"_index"`
	Id     string                 `json:"_id"`
	Score  float32                `json:"_score"`
	Source map[string]interface{} `json:"_source"`
}

type EsHitList struct {
	Total    int         `json:"total"`
	MaxScore float32     `json:"max_score"`
	Hits     []EsHitItem `json:"hits"`
}

func (x *EsClient) Search(ctx context.Context, index string, input interface{}, timeoutMilli ...int64) (*EsHitList, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(input); err != nil {
		return nil, fmt.Errorf("search_input_json_fail: err=%v", err)
	}

	timeout := time.Millisecond * time.Duration(x.TimeoutMilli)
	if len(timeoutMilli) > 0 {
		timeout = time.Millisecond * time.Duration(timeoutMilli[0])
	}

	searchResp, searchErr := x.client.Search(
		x.client.Search.WithContext(ctx),
		x.client.Search.WithIndex(index),
		x.client.Search.WithBody(&buf),
		x.client.Search.WithTrackTotalHits(true),
		x.client.Search.WithRestTotalHitsAsInt(true),
		x.client.Search.WithTrackScores(true),
		x.client.Search.WithTimeout(timeout),
	)
	if searchErr != nil || searchResp == nil {
		return nil, fmt.Errorf("search_api_fail: err=%v", searchErr)
	}

	resp := &struct {
		Error    interface{} `json:"error"`
		Took     int         `json:"took"`
		TimedOut bool        `json:"timed_out"`
		Hits     EsHitList   `json:"hits"`
	}{}
	if readErr := x.readRespBody(searchResp.Body, resp); readErr != nil {
		return nil, fmt.Errorf("search_read_fail: err=%v", readErr)
	}
	if resp.Error != nil {
		ba, _ := json.Marshal(resp.Error)
		return nil, fmt.Errorf("search_resp_error: error=%s", string(ba))
	}

	return &resp.Hits, nil
}

func (x *EsClient) readRespBody(body io.ReadCloser, target interface{}) error {
	defer body.Close()
	ba, baErr := ioutil.ReadAll(body)
	if baErr != nil {
		return fmt.Errorf("resp_body_read_fail: err=%s", baErr.Error())
	}
	if err := json.Unmarshal(ba, target); err != nil {
		return fmt.Errorf("resp_body_json_fail: err=%s", err.Error())
	}
	return nil
}

const (
	_ int32 = iota
	EsPredicateOpAnd
	EsPredicateOpOr
	EsPredicateOpNot
	EsPredicateOpTerm
	EsPredicateOpTerms
	EsPredicateOpGt
	EsPredicateOpGte
	EsPredicateOpLt
	EsPredicateOpLte
	EsPredicateOpBetween
	EsPredicateOpMatch
	EsPredicateOpMatchOrEqual
	EsPredicateOpMultiMatch
	EsPredicateOpMultiMatchOrEqual
)

type EsPredicate struct {
	Op    int32         `json:"op"`
	Key   string        `json:"key"`
	Value interface{}   `json:"value"`
	Inner []EsPredicate `json:"inner"`
	Boost float32       `json:"boost"`
}

func (x *EsPredicate) ToQuery() (interface{}, error) {
	meth, content, err := x.build()
	if err != nil {
		return nil, fmt.Errorf("query_build_fail: err=%v", err)
	}
	q := map[string]interface{}{
		meth: content,
	}
	return q, nil
}

func (x *EsPredicate) build() (string, map[string]interface{}, error) {
	switch x.Op {
	case EsPredicateOpAnd, EsPredicateOpOr:
		if len(x.Inner) < 1 {
			return "", nil, fmt.Errorf("empty_predicate: op=%d", x.Op)
		}
		var qs []interface{}
		for offset, inner := range x.Inner {
			meth, qInner, err := inner.build()
			if err != nil {
				return "", nil, fmt.Errorf("inner_error: offset=%d err=%v", offset, err)
			}
			qs = append(qs, map[string]interface{}{
				meth: qInner,
			})
		}
		var q2 map[string]interface{}
		switch x.Op {
		case EsPredicateOpAnd:
			q2 = map[string]interface{}{
				"must": qs,
			}
		case EsPredicateOpOr:
			q2 = map[string]interface{}{
				"must": map[string]interface{}{
					"bool": map[string]interface{}{
						"should": qs,
					},
				},
			}
		default:
			return "", nil, fmt.Errorf("imposible_branch")
		}
		if x.Boost != 0 {
			q2["boost"] = x.Boost
		}
		return "bool", q2, nil

	case EsPredicateOpNot:
		if l := len(x.Inner); l != 1 {
			return "", nil, fmt.Errorf("invalid_not_predicate_len: len=%d", l)
		}
		meth, qInner, err := x.Inner[0].build()
		if err != nil {
			return "", nil, fmt.Errorf("inner_error: err=%v", err)
		}
		q2 := map[string]interface{}{
			"must_not": map[string]interface{}{
				meth: qInner,
			},
		}
		if x.Boost != 0 {
			q2["boost"] = x.Boost
		}
		return "bool", q2, nil

	case EsPredicateOpTerm, EsPredicateOpTerms:
		value := map[string]interface{}{
			"value": x.Value,
		}
		if x.Boost != 0 {
			value["boost"] = x.Boost
		}
		q := map[string]interface{}{
			x.Key: value,
		}
		var meth string
		switch x.Op {
		case EsPredicateOpTerm:
			meth = "term"
		case EsPredicateOpTerms:
			meth = "terms"
		default:
			return "", nil, fmt.Errorf("imposible_branch")
		}
		return meth, q, nil

	case EsPredicateOpGt, EsPredicateOpGte, EsPredicateOpLt, EsPredicateOpLte:
		var opStr string
		switch x.Op {
		case EsPredicateOpGt:
			opStr = "gt"
		case EsPredicateOpGte:
			opStr = "gte"
		case EsPredicateOpLt:
			opStr = "lt"
		case EsPredicateOpLte:
			opStr = "lte"
		default:
			return "", nil, fmt.Errorf("imposible_branch")
		}
		value := map[string]interface{}{
			opStr: x.Value,
		}
		if x.Boost != 0 {
			value["boost"] = x.Boost
		}
		q := map[string]interface{}{
			x.Key: value,
		}
		return "range", q, nil

	case EsPredicateOpBetween:
		if x.Value == nil {
			return "", nil, fmt.Errorf("between_value_nil")
		}
		valueAsList, valueAsListOk := x.Value.([]interface{})
		if !valueAsListOk {
			return "", nil, fmt.Errorf("between_value_not_list")
		}
		if len(valueAsList) != 2 {
			return "", nil, fmt.Errorf("invalid_between_value_len")
		}
		value := map[string]interface{}{
			"gte": valueAsList[0],
			"lte": valueAsList[1],
		}
		if x.Boost != 0 {
			value["boost"] = x.Boost
		}
		q := map[string]interface{}{
			x.Key: value,
		}
		return "range", q, nil

	case EsPredicateOpMatch:
		q := map[string]interface{}{
			x.Key: map[string]interface{}{
				"query":    x.Value,
				"operator": "and",
			},
		}
		return "match", q, nil

	case EsPredicateOpMatchOrEqual:
		var boost float32 = 100
		if x.Boost != 0 {
			boost = x.Boost
		}
		rewritePred := EsPredicate{
			Op: EsPredicateOpOr,
			Inner: []EsPredicate{
				{
					Op:    EsPredicateOpTerm,
					Key:   fmt.Sprintf("%s.keyword", x.Key),
					Value: x.Value,
					Boost: boost,
				},
				{
					Op:    EsPredicateOpMatch,
					Key:   x.Key,
					Value: x.Value,
				},
			},
		}
		return rewritePred.build()

	case EsPredicateOpMultiMatch:
		q := map[string]interface{}{
			"query":    x.Value,
			"fields":   strings.Split(x.Key, ","),
			"operator": "and",
		}
		return "multi_match", q, nil

	case EsPredicateOpMultiMatchOrEqual:
		rewritePred := EsPredicate{
			Op: EsPredicateOpOr,
			Inner: []EsPredicate{
				{
					Op:    EsPredicateOpMultiMatch,
					Key:   x.Key,
					Value: x.Value,
				},
			},
		}
		var boost float32 = 100
		if x.Boost != 0 {
			boost = x.Boost
		}
		for _, k := range strings.Split(x.Key, ",") {
			rewritePred.Inner = append(rewritePred.Inner, EsPredicate{
				Op:    EsPredicateOpTerm,
				Key:   fmt.Sprintf("%s.keyword", k),
				Value: x.Value,
				Boost: boost,
			})
		}
		return rewritePred.build()
	}

	return "", nil, fmt.Errorf("invalid_op: op=%d", x.Op)
}
