package main // import "go.nanasi880.dev/goenv/go"

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"

	"go.nanasi880.dev/semver"
)

var (
	// Go command pattern
	// e.g.
	//   go1.11.5
	//   go1.12
	goCmdPattern = regexp.MustCompile(`^go[1-9]+\.?[0-9]*\.?[0-9]*$`)
)

var (
	// install path of Go command
	// expected directory tree
	//
	// /usr/local/go/bin          <- expected by installLocation value
	//               ├── go       <- this command
	//               ├── go1.11.5 <- symlink of actual Go command
	//               └── go1.12   <- symlink of actual Go command
	installLocation string

	// user instructed Go command version
	userInstructedGoCmd string
)

func main() {

	if runtime.GOOS != "darwin" {
		fmt.Println("darwin only")
		os.Exit(1)
	}

	loadEnv()

	if userInstructedGoCmd != "" {
		execGo(userInstructedGoCmd)
	} else {
		execGo(findLatestGo())
	}
}

// getenv is get environment value. if empty return def
func getenv(key string, def string) string {

	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadEnv is load environment values.
func loadEnv() {
	installLocation = getenv("GOENV_LOCATION", "/usr/local/go/bin")
	userInstructedGoCmd = os.Getenv("GOCMD")
}

// execGo is execution Go command and os.Args[1:] will transfer.
func execGo(cmdName string) {

	cmd := exec.Command(cmdName, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	exitCode := 0
	if err := cmd.Wait(); err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			exitCode = e.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

// mustExecCmd is execution cmdName command.
// if command returned zero, return stdout and stderr string.
// if command returned non-zero value, call log.Fatal.
func mustExecCmd(cmdName string, args ...string) string {

	cmd := exec.Command(cmdName, args...)

	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Start(); err != nil {
		log.Fatal(cmdName, " ", args, " ", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(cmdName, " ", args, " ", err)
	}

	return stdout.String()
}

// findLatestGo is find latest version Go command.
func findLatestGo() string {

	cmdList := strings.Split(mustExecCmd("ls", "-1", installLocation), "\n")

	var goCmdList []string
	for _, c := range cmdList {
		if goCmdPattern.MatchString(c) {
			goCmdList = append(goCmdList, c)
		}
	}

	if len(goCmdList) == 0 {
		fmt.Println("go: available Go not found")
		os.Exit(1)
	}

	versions := make(semver.Versions, len(goCmdList))
	for _, c := range goCmdList {
		v, err := semver.ParseWithPrefix(c, "go")
		if err != nil {
			log.Fatal(err)
		}

		versions = append(versions, v)
	}
	sort.Sort(sort.Reverse(versions))

	return filepath.Join(installLocation, versions[0].Raw)
}
