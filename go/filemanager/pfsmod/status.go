package pfsmod

//TODO:
// need a custome error type?

const (
	StatusFileNotFound          = 520
	StatusDirectoryNotAFile     = 521
	StatusCopyFromLocalToLocal  = 523
	StatusDestShouldBeDirectory = 524
	StatusOnlySupportFiles      = 526
	StatusBadFileSize           = 527
	StatusDirectoryAlreadyExist = 528
	StatusBadChunkSize          = 529
	StatusShouldBePfsPath       = 530
	StatusNotEnoughArgs         = 531
	StatusInvalidArgs           = 532
	StatusUnAuthorized          = 533
	StatusJsonErr               = 534
	StatusCannotDelDirectory    = 535
	StatusAlreadyExist          = 536
)

var statusText = map[int]string{
	StatusFileNotFound:          "no such file or directory",
	StatusDirectoryNotAFile:     "should be a file not a directory",
	StatusCopyFromLocalToLocal:  "don't support copy local to local",
	StatusDestShouldBeDirectory: "dest should be a directory",
	StatusOnlySupportFiles:      "only support upload or download files, not directories",
	StatusBadFileSize:           "bad file size",
	StatusDirectoryAlreadyExist: "directory already exist",
	StatusBadChunkSize:          "chunksize error",
	StatusShouldBePfsPath:       "should be pfs path",
	StatusNotEnoughArgs:         "not enough arguments",
	StatusInvalidArgs:           "invalid arguments",
	StatusUnAuthorized:          "what you request is unauthorized",
	StatusJsonErr:               "parse json error",
	StatusCannotDelDirectory:    "can't del directory",
	StatusAlreadyExist:          "already exist",
}

func StatusText(code int) string {
	return statusText[code]
}
