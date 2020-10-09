package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	dump_tables "github.com/linger1216/jelly-schedule-jobs/jobs/dump-tables"
	"github.com/linger1216/jelly-schedule/core"
)

func main() {
	//test()
	core.StartClientJob(dump_tables.NewDumpTablesJob())
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
	xxx, _ := dump_tables.NewDumpTablesJob().Exec(nil, para)
	fmt.Printf("%v\n", xxx)
}
