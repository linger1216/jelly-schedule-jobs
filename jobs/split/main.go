package main

import (
	"context"
	"fmt"
)

/*
$date 代表日期2020-12-12
$timestamp 代表unix时间戳 1598956864
$timestamp-ms 代表unix时间戳 1598956864000
$time 代表时间精确到秒2020-12-12_08:00:00
$time-second 代表时间2020-12-12_08:00:00
$time-minute 代表时间2020-12-12_08:00
$time-hour 代表时间2020-12-12_08
$round 占位符
*/

type StringType struct {
	Keys   []string
	Format string
	Round  []string
}

type NumberType struct {
	Min    int
	Max    int
	Format string
}

type SplitRequest struct {
	Count int
}

type SplitJob struct {
}

func NewSplitJob() *SplitJob {
	return &SplitJob{}
}

func (e *SplitJob) Name() string {
	return "Split"
}

func (e *SplitJob) Exec(ctx context.Context, req string) (resp string, err error) {

	fmt.Printf("%s\n", req)

	return req, nil
}

func main() {

}
