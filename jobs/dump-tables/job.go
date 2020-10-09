package dump_tables

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/linger1216/go-utils/convert"
	"github.com/linger1216/go-utils/db/postgres"
	"github.com/linger1216/jelly-schedule/core"
	"net/url"
	"strings"
)

type DumpTablesJob struct {
}

func NewDumpTablesJob() *DumpTablesJob {
	return &DumpTablesJob{}
}

func (s *DumpTablesJob) Name() string {
	return "DumpTablesJob"
}

func (s *DumpTablesJob) Exec(ctx context.Context, req string) (string, error) {
	if len(req) == 0 {
		return "", nil
	}

	request, err := core.UnMarshalJobRequest(req)
	if err != nil {
		return "", err
	}

	reqs := make([]*Request, 0)
	for k, arr := range request.Values {
		if strings.HasPrefix(k, KeyPrefix) {
			for _, v := range arr {
				r := &Request{}
				if err := jsoniter.ConfigFastest.UnmarshalFromString(v, r); err == nil {
					reqs = append(reqs, r)
				} else {
					return "", err
				}
			}
		}
	}
	resp := core.NewJobRequestByMeta(request)
	for i := range reqs {
		for j := range reqs[i].Uri {
			uri := reqs[i].Uri[j]
			info, err := _exec(uri)
			if err != nil {
				return "", err
			}
			buf, err := jsoniter.ConfigFastest.Marshal(info)
			if err != nil {
				return "", err
			}
			resp.Values[fmt.Sprintf("%s-%s", HandlePrefix, uri)] =
				append(resp.Values[HandlePrefix], string(buf))
		}
	}

	return core.MarshalJobRequest(resp)
}

func _exec(uri string) (*Response, error) {
	if len(uri) == 0 {
		return nil, nil
	}

	xx, err := url.Parse(strings.ToLower(uri))
	if err != nil {
		return nil, err
	}

	switch xx.Scheme {
	case "postgres":
		return _readPostgresTables(uri)
	case "mysql":
		fmt.Printf("unsupport %s\n", xx.Scheme)
	default:
		fmt.Printf("unsupport %s\n", xx.Scheme)
		return nil, nil
	}

	return nil, nil
}

func _readPostgresTables(uri string) (*Response, error) {
	const ShowTablesQuery = `select relname as table_name,cast(obj_description(relfilenode,'pg_class') as varchar) as comment from pg_class c 
where  relkind = 'r' and relname not like 'pg_%' and relname not like 'sql_%' and relchecks=0 order by relname;`

	db := postgres.NewPostgres(postgres.NewConfig(uri))
	defer db.Close()

	rows, err := db.Queryx(ShowTablesQuery)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	arr := make([]string, 0)
	for rows.Next() {
		line := make(map[string]interface{})
		err = rows.MapScan(line)
		if err != nil {
			return nil, err
		}

		if v, ok := line["table_name"]; ok {
			t := convert.ToString(v)
			if len(t) > 0 {
				arr = append(arr, t)
			}
		}
	}
	return &Response{
		Uri:    uri,
		Tables: arr,
	}, nil
}
