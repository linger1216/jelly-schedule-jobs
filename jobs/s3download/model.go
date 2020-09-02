package s3download

type Request struct {
	Keys              []string `json:"keys,omitempty" yaml:"keys" `
	DownloadDirectory string   `json:"downloadDirectory,omitempty" yaml:"downloadDirectory" `
	// MB
	ReserveSpace       uint64 `json:"reserveSpace,omitempty" yaml:"reserveSpace"`
	SpaceCheckInterval int    `json:"spaceCheckInterval,omitempty" yaml:"spaceCheckInterval" `

	// tar.gz, gz, zip
	DeCompress bool `json:"decompress,omitempty" yaml:"decompress"`
}
