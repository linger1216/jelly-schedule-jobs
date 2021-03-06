package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/linger1216/jelly-schedule-jobs/jobs/s3download/s3"
	"github.com/linger1216/jelly-schedule/core"
	"io"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

import _ "net/http/pprof"

type S3DownloadJob struct {
	s3svc *s3.S3Svc
}

func NewS3DownloadJob(conf s3.S3Config) *S3DownloadJob {
	return &S3DownloadJob{s3svc: s3.NewS3Svc(conf)}
}

func (e *S3DownloadJob) Name() string {
	return "S3Download"
}

func (e *S3DownloadJob) Exec(ctx context.Context, req string) (resp string, err error) {

	fmt.Printf("req: %s\n", req)

	reqs, err := core.UnMarshalJobRequests(req, ";")
	if err != nil {
		return "", err
	}

	if len(reqs) == 0 {
		return "", fmt.Errorf("empty request")
	}

	if len(reqs) > 1 {
		fmt.Printf("too many job request\n")
	}

	request := reqs[0]
	DeCompress := request.GetBoolFromMeta(fmt.Sprintf("%s-%s", e.Name(), "DeCompress"))
	DownloadDirectory := request.GetStringFromMeta(fmt.Sprintf("%s-%s", e.Name(), "DownloadDirectory"))
	if len(DownloadDirectory) == 0 {
		return "", fmt.Errorf("DownloadDirectory is empty")
	}
	if !strings.HasSuffix(DownloadDirectory, string(os.PathSeparator)) {
		DownloadDirectory += string(os.PathSeparator)
	}

	ReserveSpace := uint64(request.GetInt64FromMeta(fmt.Sprintf("%s-%s", e.Name(), "ReserveSpace")))
	if ReserveSpace == 0 {
		ReserveSpace = 1024
	}

	SpaceCheckInterval := request.GetInt64FromMeta(fmt.Sprintf("%s-%s", e.Name(), "SpaceCheckInterval"))
	if SpaceCheckInterval == 0 {
		SpaceCheckInterval = 600
	}

	if len(request.Values) == 0 {
		return "", nil
	}

	// create path
	err = os.MkdirAll(DownloadDirectory, os.ModePerm)
	if err != nil {
		return "", err
	}

	m := make(map[string][]string)

	insertM := func(k, v string) {
		if _, ok := m[k]; ok {
			m[k] = append(m[k], v)
		} else {
			m[k] = []string{v}
		}
	}

	for group := range request.Values {
		for i := 0; i < len(request.Values[group]); i++ {
			key := request.Values[group][i]

			// check 剩余容量
			usage := DiskUsage(DownloadDirectory)
			if usage.Free <= ReserveSpace*1024*1024 {
				fmt.Printf("space not enough %dMB<%dMB\n", usage.Free/1024/1024, ReserveSpace)
				timer := time.NewTimer(time.Second * 600)
				select {
				case <-timer.C:
				}
				i--
				continue
			}

			// 下载
			downloadFilename := DownloadDirectory + path.Base(key)
			tmpDownloadFilename := downloadFilename + ".tmp"

			fmt.Printf("download file:%s...\n", key)
			//if !PathExists(tmpDownloadFilename) {
			_, err := e.s3svc.DownloadObject(key, tmpDownloadFilename)
			if err != nil {
				return "", err
			}
			//}

			compressFormat := exactCompressFormat(path.Base(key))
			if !DeCompress || len(compressFormat) == 0 {
				// 说明不需要解压
				// 重命名
				if err := os.Rename(tmpDownloadFilename, downloadFilename); err != nil {
					return "", err
				}

				// 保存新values
				insertM(group, downloadFilename)
			} else {
				fmt.Printf("decompress %s\n", downloadFilename)
				// 解压
				prefix := strings.ReplaceAll(path.Dir(key), string(os.PathSeparator), "_") + "_"
				switch compressFormat {
				case ".zip":
					arr, err := deCompressZipFile(tmpDownloadFilename, DownloadDirectory, prefix)
					if err != nil {
						return "", err
					}
					for j := range arr {
						insertM(group, arr[j])
					}
				case ".gz":
				case ".targz":
					arr, err := deCompressTarGzFile(tmpDownloadFilename, DownloadDirectory, prefix)
					if err != nil {
						return "", err
					}
					for j := range arr {
						insertM(group, arr[j])
					}
				}

				// 删除临时文件
				err = os.Remove(tmpDownloadFilename)
				if err != nil {
					return "", err
				}
			}
		}
	}

	request.Values = m

	// end
	//endFile := "end"
	//file, err := os.Create(request.DownloadDirectory + endFile)
	//if err != nil {
	//	return "", err
	//}
	//file.Close()

	return core.MarshalJobRequests(";", request)
}

func deCompressTarGzFile(gzFileName, directory, prefix string) ([]string, error) {

	arr := make([]string, 0)

	srcFile, err := os.Open(gzFileName)
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	gr, err := gzip.NewReader(srcFile)
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	decompressSingle := func(filename string) error {
		filenameTmp := filename + ".tmp"
		file, err := os.OpenFile(filenameTmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(file, tr); err != nil {
			return err
		}
		if err := os.Rename(filenameTmp, filename); err != nil {
			return err
		}

		arr = append(arr, filename)
		return nil
	}

	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		err = decompressSingle(directory + prefix + hdr.Name)
		if err != nil {
			return nil, err
		}
	}

	return arr, nil
}

func deCompressZipFile(zipFile, directory, prefix string) ([]string, error) {

	arr := make([]string, 0)

	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	deCompressSingle := func(directory, filename string, file *zip.File) error {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			return err
		}

		tmp := filename + ".tmp"
		w, err := os.Create(tmp)
		if err != nil {
			return err
		}
		defer w.Close()

		_, err = io.Copy(w, rc)
		if err != nil {
			return err
		}
		w.Close()
		rc.Close()

		if err := os.Rename(tmp, filename); err != nil {
			return err
		}

		arr = append(arr, filename)
		return nil
	}

	for i := range reader.File {
		err = deCompressSingle(directory, reader.File[i].Name, reader.File[i])
		if err != nil {
			return nil, err
		}
	}
	return arr, nil
}

func main() {
	s3config := &s3.S3Config{}
	_ = core.LoadUserConfig("s3", s3config)
	core.StartClientJob(NewS3DownloadJob(*s3config))
}

func exactCompressFormat(filename string) string {
	ext := path.Ext(filename)
	switch strings.ToLower(ext) {
	case ".gz":
		remain := path.Ext(filename[:len(filename)-3])
		if remain == ".tar" {
			return ".targz"
		}
		return ext
	case ".zip":
		return ext
	}
	return ""
}

type DiskStatus struct {
	All  uint64 `json:"all"`
	Used uint64 `json:"used"`
	Free uint64 `json:"free"`
}

func DiskUsage(path string) (disk DiskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	disk.All = fs.Blocks * uint64(fs.Bsize)
	disk.Free = fs.Bfree * uint64(fs.Bsize)
	disk.Used = disk.All - disk.Free
	return
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	return false
}
