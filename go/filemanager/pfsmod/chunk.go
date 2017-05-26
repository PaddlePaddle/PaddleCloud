package pfsmod

type Chunk struct {
	Meta ChunkMeta
	Data []byte
}

type ChunkCommand struct {
	Path string
	Data Chunk
}

func (p *ChunkCommand) ToUrl() string {
}

func (p *ChunkCommand) ToJson() []byte {
}

func (p *ChunkCommand) Run() interface{} {
}

func NewChunkCommandFromUrl() *ChunkCommand {
}
