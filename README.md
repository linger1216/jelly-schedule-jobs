# jelly-schedule-jobs

- S3DownLoadJob
- SplitJob





func Test_S3(t *testing.T) {
	s3config := &s3.S3Config{}
	_ = core.LoadUserConfig("s3", s3config)
	j := main.NewS3DownloadJob(*s3config)
	req := &main.S3DownloadRequest{
		Keys: []string{"20200801/35b277c.txt.tar.gz",
			"20200801/368372c.txt.tar.gz",
			"20200801/344b624.txt.tar.gz",
			"20200801/35f0fc.txt.tar.gz"},
		DownloadDirectory:  "/data1/download",
		ReserveSpace:       1024,
		SpaceCheckInterval: 600,
		DeCompress:         false,
	}
	buf, _ := jsoniter.ConfigFastest.Marshal(req)
	_, err := j.Exec(context.Background(), string(buf))
	if err != nil {
		panic(err)
	}
}