package main

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/linger1216/jelly-schedule-jobs/jobs/scp"
	"github.com/linger1216/jelly-schedule/core"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type ScpJob struct {
}

func NewScpJob() *ScpJob {
	return &ScpJob{}
}

func (s *ScpJob) Name() string {
	return "ScpJob"
}

func (s *ScpJob) Exec(ctx context.Context, req string) (string, error) {
	if len(req) == 0 {
		return "", nil
	}

	request, err := core.UnMarshalJobRequest(req)
	if err != nil {
		return "", err
	}

	reqs := make([]*scp.Request, 0)
	for k, arr := range request.Values {
		if strings.HasPrefix(k, scp.KeyPrefix) {
			for _, v := range arr {
				r := &scp.Request{}
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
		scpFiles, err := _exec(reqs[i])
		if err != nil {
			return "", err
		}
		resp.Values[scp.HandlePrefix] = append(resp.Values[scp.HandlePrefix], scpFiles...)
	}

	return core.MarshalJobRequest(resp)
}

// 哪个目录下, 有哪些文件进行了scp
func _exec(request *scp.Request) ([]string, error) {
	if strings.HasSuffix(request.DstDirectory, string(os.PathSeparator)) {
		request.DstDirectory = request.DstDirectory[:len(request.DstDirectory)-1]
	}

	cmds := make([]string, 0)
	handleFiles := make([]string, 0)

	// scp -P 22 -r tmp root@114.67.106.133:/root
	port := 22
	if request.DstPort > 0 {
		port = request.DstPort
	}

	// 处理目录的情况
	for i := range request.SrcDirectories {
		srcDirectory := request.SrcDirectories[i]
		parentSrcDirectory := path.Dir(srcDirectory)
		cmds = append(cmds, fmt.Sprintf("scp -P %d -r %s %s@%s:%s", port, srcDirectory,
			request.DstUser, request.DstHost, request.DstDirectory))
		srcDirectoryFiles := getFileList(srcDirectory)
		for _, srcDirectoryFile := range srcDirectoryFiles {
			handleFiles = append(handleFiles, strings.Replace(srcDirectoryFile, parentSrcDirectory, request.DstDirectory, 1))
		}
	}

	// 处理文件的情况
	for i := range request.SrcFiles {
		srcFile := request.SrcFiles[i]
		cmds = append(cmds, fmt.Sprintf("scp -P %d %s %s@%s:%s", port, srcFile,
			request.DstUser, request.DstHost, request.DstDirectory))
		srcFileName := path.Base(srcFile)
		handleFiles = append(handleFiles, request.DstDirectory+string(os.PathSeparator)+srcFileName)
	}

	for _, cmd := range cmds {
		command := exec.Command("/bin/sh", "-c", cmd)
		_, err := command.Output()
		if err != nil {
			return nil, err
		}
	}

	return handleFiles, nil
}

func getFileList(path string) []string {
	arr := make([]string, 0)
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		arr = append(arr, path)
		return nil
	})

	if err != nil {
		panic(err)
	}
	return arr
}

func main() {
	core.StartClientJob(NewScpJob())
	//test()
}

func test() {

	req := &scp.Request{
		SrcFiles: []string{
			"/Users/lid.guan/Downloads/go_module_proc/cron-task/jelly-schedule-jobs/go.mod",
			"/Users/lid.guan/Downloads/go_module_proc/cron-task/jelly-schedule-jobs/README.md",
			"/Users/lid.guan/Desktop/neighborhood.tar.gz",
		},
		SrcDirectories: []string{
			"/Users/lid.guan/Downloads/go_module_proc/cron-task/jelly-schedule-jobs",
		},
		DstUser:      "root",
		DstHost:      "114.67.106.133",
		DstPort:      22,
		DstDirectory: "/root/scp",
	}

	//req := &scp.Request{
	//	SrcFiles: []string{"/Users/lid.guan/Downloads/go_module_proc/cron-task/jelly-schedule-jobs/go.mod",
	//		"/Users/lid.guan/Downloads/go_module_proc/cron-task/jelly-schedule-jobs/README.md"},
	//	DstUser:      "root",
	//	DstHost:      "114.67.106.133",
	//	DstPort:      22,
	//	DstDirectory: "/root/scp",
	//}

	buf, _ := jsoniter.ConfigFastest.Marshal(req)

	jobRequest := core.NewJobRequest()
	jobRequest.Values[scp.KeyPrefix] = append(jobRequest.Values[scp.KeyPrefix], string(buf))

	para, _ := core.MarshalJobRequest(jobRequest)
	xxx, _ := NewScpJob().Exec(nil, para)
	fmt.Printf("%v\n", xxx)
}
