package pfsmod

const (
	StatusFileNotFound                     = 520
	StatusDirectoryNotAFile                = 521
	StatusCopyFromLocalToLocal             = 523
	StatusDestShouldBeDirectory            = 524
	StatusOnlySupportUploadOrDownloadFiles = 526
)

var statusText = map[int]string{
	StatusFileNotFound:                     "no such file or directory",
	StatusDirectoryNotAFile:                "should be a file not a directory",
	StatusCopyFromLocalToLocal:             "don't support copy local to local",
	StatusDestShouldBeDirectory:            "dest should be a directory",
	StatusOnlySupportUploadOrDownloadFiles: "only support upload or download files, not directories",
}

func StatusText(code int) string {
	return statusText[code]
}
