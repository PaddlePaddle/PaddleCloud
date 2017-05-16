package pfscommon

type FilePath struct {
	Path       string
	Local      bool
	DataCenter string
	User       string
	UserPath   string
}

type Cmd struct {
	Method  string
	Options []string
	Src     []FilePath
	Dest    FilePath
}

func CheckLsSynopsis(cmdStr string) (Cmd, error) {
	cmd := Cmd{}
	return cmd, nil
}

func CheckRmSynopsis(cmdStr string) (Cmd, error) {
	cmd := Cmd{}
	return cmd, nil
}

func CheckCpSynopsis(cmdStr string) (Cmd, error) {
	cmd := Cmd{}
	return cmd, nil
}

func CheckCmdAuthor(cmd Cmd, user string) error {
	return nil
}
