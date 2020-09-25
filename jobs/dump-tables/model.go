package dump_tables

const (
	KeyPrefix    = "DUMP"
	HandlePrefix = "DUMP-HANDLE"
)

type Request struct {
	Uri []string `json:"uri,omitempty" yaml:"uri"`
}

type Response struct {
	Uri    string   `json:"uri,omitempty" yaml:"uri"`
	Tables []string `json:"tables,omitempty" yaml:"tables"`
}
