package scp

const (
	KeyPrefix    = "SCP"
	HandlePrefix = "SCP-HANDLE"
)

type Request struct {
	SrcFiles       []string `json:"srcFiles,omitempty" yaml:"srcFiles" `
	SrcDirectories []string `json:"srcDirectories,omitempty" yaml:"srcDirectories" `
	DstUser        string   `json:"dstUser,omitempty" yaml:"dstUser" `
	DstHost        string   `json:"dstHost,omitempty" yaml:"dstHost" `
	DstPort        int      `json:"dstPort,omitempty" yaml:"dstPort" `
	DstDirectory   string   `json:"dstDirectory,omitempty" yaml:"dstDirectory" `
}
