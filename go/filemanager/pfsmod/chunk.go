package pfsmod

type ChunkCmd struct {
	Path      string
	Offset    int64
	ChunkSize int64
	Data      []byte
}

func (p *ChunkCmd) ToUrl() string {
	parameters := url.Values{}
	parameters.Add("path", p.path)

	str := fmt.Sprint(offset)
	parameters.Add("offset", p.Meta.offset)

	str = fmt.Sprint(p.Meta.ChunkSize)
	parameters.Add("chunksize", str)

	return parameters.Encode()
}

func (p *ChunkCmd) ToJson() []byte {
	return nil
}

func (p *ChunkCmd) Run() interface{} {
	return nil
}

// path example:
// 	  path=/pfs/datacenter1/1.txt&offset=4096&chunksize=4096
func NewChunkCmdFromUrl(path string) (*ChunkCmd, int32) {
	cmd := ChunkCmd{}

	m, err := url.ParseQuery(path)
	if err != nil ||
		len(m["path"]) == 0 ||
		len(m["offset"]) == 0 ||
		len(m["chunksize"]) == 0 {
		return nil, http.StatusBadRequest
	}

	//var err error
	cmd.Path = m["path"][0]
	cmd.Offset, err = strconv.ParseInt(m["offset"][0], 10, 64)
	if err != nil {
		return nil, http.StatusBadRequest
	}

	chunkSize, err := strconv.ParseInt(m["chunksize"][0], 10, 64)
	if err != nil {
		return nil, errors.New("bad arguments offset")
	}
	cmd.ChunkSize = chunkSize

	return &cmd, nil
}

func (p *ChunkCmd) WriteChunkData(w io.Writer) error {
	f, err := os.Open(p.path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(p.Offset, 0)
	if err != nil {
		return err
	}

	writer := multipart.NewWriter(w)
	defer writer.Close()

	writer.SetBoundary(DefaultMultiPartBoundary)

	fName := p.ToUrl()
	part, err := writer.CreateFormFile("chunk", fName)
	if err != nil {
		return err
	}

	_, err = io.CopyN(part, f, p.ChunkSize)
	if err != nil {
		return err
	}
	return nil
}
