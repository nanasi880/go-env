package goenv

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nanasi880/go-env/internal/config"
	"github.com/nanasi880/go-env/internal/remote"
)

func Main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	if os.Args[1] == "init" {
		os.Exit(execInit())
	}
	if os.Args[1] == "install" {
		os.Exit(execInstall())
	}
	if os.Args[1] == "upgrade" {
		os.Exit(execInstall())
	}
	if os.Args[1] == "list-remote" {
		os.Exit(execListRemote())
	}

	os.Exit(1)
}

func execInit() int {
	var (
		location        = config.LoadLocation()
		binDirectory    = filepath.Join(location, config.BinaryDirectoryName)
		binaryPath      = filepath.Join(binDirectory, config.BinaryName)
		goRootDirectory = filepath.Join(location, config.GoRootDirectoryName)
	)

	fmt.Printf("%s にツールをセットアップします ", location)
	if yesno() == false {
		fmt.Println("中止")
		return 1
	}

	mkdir(binDirectory, 0755)
	mkdir(goRootDirectory, 0755)

	// install myself
	executablePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// copy executable file
	func() {
		srcFile := mustOpen(executablePath)
		defer mustClose(srcFile)

		dstFile := mustOpenFile(binaryPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
		defer mustClose(dstFile)

		mustCopy(dstFile, srcFile)
		fmt.Println("install binary: " + binaryPath)
	}()

	// install symlinks
	mustSymlink(binaryPath, filepath.Join(binDirectory, "go"))
	fmt.Println("create symlink: " + filepath.Join(binDirectory, "go") + " => " + binaryPath)

	mustSymlink(binaryPath, filepath.Join(binDirectory, "gofmt"))
	fmt.Println("create symlink: " + filepath.Join(binDirectory, "gofmt") + " => " + binaryPath)

	return 0
}

func execInstall() int {
	panic("not implemented")
}

func execUpgrade() int {
	sources := mustCrawl()
	_ = sources
	return 0
}

func execListRemote() int {
	sources := mustCrawl()
	for _, source := range sources {
		fmt.Println(source.Version.String())
	}
	return 0
}

func mustCrawl() []remote.Source {
	sources, err := remote.Crawl()
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
	return sources
}

func mustOpen(name string) *os.File {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	return f
}

func mustOpenFile(name string, flag int, perm os.FileMode) *os.File {
	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		panic(err)
	}
	return f
}

func mustCopy(dst io.Writer, src io.Reader) int64 {
	n, err := io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
	return n
}

func mustClose(f io.Closer) {
	err := f.Close()
	if err != nil {
		panic(err)
	}
}

func mustSymlink(old, new string) {
	_ = os.Remove(new)
	err := os.Symlink(old, new)
	if err != nil {
		panic(err)
	}
}

func mkdir(name string, perm os.FileMode) {
	stat, err := os.Stat(name)
	if os.IsNotExist(err) {
		err = os.MkdirAll(name, perm)
		if err != nil {
			panic(err)
		}
		return
	}

	if stat.IsDir() {
		// already exists
		return
	}

	panic("file already exists")
}

func yesno() bool {
	for {
		fmt.Printf("yes/no: ")

		var s string
		_, err := fmt.Scanln(&s)
		if err != nil {
			panic(err)
		}
		switch s {
		case "yes":
			return true
		case "no":
			return false
		}
	}
}

func mustDownloadSource(source remote.Source) remote.Archive {
	temp, err := os.MkdirTemp("", "goenv-")
	if err != nil {
		panic(err)
	}

	archive, err := source.Download(temp)
	if err != nil {
		panic(err)
	}
	return archive
}
