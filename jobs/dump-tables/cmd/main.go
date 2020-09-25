package main

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	dump_tables "github.com/linger1216/jelly-schedule-jobs/jobs/dump-tables"
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

	reqs := make([]*dump_tables.Request, 0)
	for k, arr := range request.Values {
		if strings.HasPrefix(k, dump_tables.KeyPrefix) {
			for _, v := range arr {
				r := &dump_tables.Request{}
				if err := jsoniter.ConfigFastest.UnmarshalFromString(v, r); err == nil {
					reqs = append(reqs, r)
				} else {
					return "", err
				}
			}
		}
	}

	ret := make([]string, 0)
	for i := range reqs {
		for j := range reqs[i].Uri {
			tables, err := _exec(reqs[i].Uri[j])
			if err != nil {
				return "", err
			}
			ret = append(ret, tables...)
		}
	}

	resp := core.NewJobRequestByMeta(request)
	resp.Values[dump_tables.HandlePrefix] = append(resp.Values[dump_tables.HandlePrefix], ret...)
	return core.MarshalJobRequest(resp)
}

func _exec(uri string) ([]string, error) {
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

func _readPostgresTables(uri string) ([]string, error) {
	const ShowTablesQuery = `select relname as table_name,cast(obj_description(relfilenode,'pg_class') as varchar) as comment from pg_class c 
where  relkind = 'r' and relname not like 'pg_%' and relname not like 'sql_%' and relchecks=0 order by relname;`

	db := dump_tables.NewPostgres(&dump_tables.PostgresConfig{
		Uri: uri,
	})

	defer db.Close()

	rows, err := db.Queryx(ShowTablesQuery)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	ret := make([]string, 0)
	for rows.Next() {
		line := make(map[string]interface{})
		err = rows.MapScan(line)
		if err != nil {
			return nil, err
		}

		if v, ok := line["table_name"]; ok {
			t := _toString(v)
			if len(t) > 0 {
				ret = append(ret, t)
			}
		}
	}
	return ret, nil
}

func _toString(v interface{}) string {
	if ret, ok := v.(string); ok {
		return ret
	}

	if ret, ok := v.([]byte); ok {
		return string(ret)
	}
	return ""
}

func main() {
	//test()
	core.StartClientJob(NewDumpTablesJob())
}

func test() {
	req := &dump_tables.Request{
		Uri: []string{
			"postgres://lid.guan:@localhost:15432/geocoding?sslmode=disable",
		},
	}
	buf, _ := jsoniter.ConfigFastest.Marshal(req)
	jobRequest := core.NewJobRequest()
	jobRequest.Values[dump_tables.KeyPrefix] = append(jobRequest.Values[dump_tables.KeyPrefix], string(buf))
	para, _ := core.MarshalJobRequest(jobRequest)
	xxx, _ := NewDumpTablesJob().Exec(nil, para)
	fmt.Printf("%v\n", xxx)
}
