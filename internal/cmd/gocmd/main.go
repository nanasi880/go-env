package gocmd

import (
	"bytes"
	"os/exec"
)

func Main() {

}

func getGitRootPath() string {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	return string(removeNewLine(out))
}

// removeNewLine は行末から改行コードを削除します。
func removeNewLine(b []byte) []byte {
	return bytes.TrimRight(b, "\n")
}
