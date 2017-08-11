package pfsmodules

import "testing"

func TestValidPfsPath(t *testing.T) {
	path := []string{"/pfs/xxx/public/dataset"}
	err := ValidatePfsPath(path, "user", "ls")
	if err != nil {
		t.Error("valid meets error")
	}

	err = ValidatePfsPath(path, "user", "mkdir")
	if err == nil {
		t.Error("valid meets error")
	}

	path = []string{"/pfs/xxx/home/user/dataset"}
	err = ValidatePfsPath(path, "user1", "ls")
	if err == nil {
		t.Error("valid meets error")
	}

	path = []string{"/pfs/xxx/home/user/dataset"}
	err = ValidatePfsPath(path, "user", "ls")
	if err != nil {
		t.Errorf("valid meets error:%v", err)
	}
}
