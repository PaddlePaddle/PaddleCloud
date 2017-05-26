package pfsmod

type ChunkMeta struct {
	Offset   int64 `json:"offset"`
	Checksum []byte
	Len      int64 `chunksize`
}

type ChunkMetaCommandResult struct {
	StatusCode int32
	StatusText string
	Metas      []ChunkMeta
}

type ChunkMataCommand struct {
	FilePath  string
	ChunkSize int64
}

func (p *ChunkMetaCommand) ToUrl() string {
}

func (p *ChunkMetaCommand) ToJson() []byte {
}

func (p *ChunkMetaCommand) Run() interface{} {
}

func NewChunkMetaFromJson([]byte) *ChunkMeta {
}
