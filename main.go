package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var verboseMode bool

type Options struct {
	verbose bool
}

func parse_option() *Options {
	ret := &Options{}
	flag.BoolVar(&ret.verbose, "verbose, V", false, "verbose mode")
	flag.Parse()
	return ret
}

func execute(tool string, params []string, debug bool) string {
	if debug {
		fmt.Printf(" >> %v %v\n", tool, params)
	}
	log, err := exec.Command(tool, params...).CombinedOutput()
	if err != nil {
		fmt.Println(string(log))
		panic(err)
	}
	return string(log)
}

func executable(bin string, verbose bool) bool {
	if path, err := exec.LookPath(bin); err != nil {
		fmt.Printf("%v\n", err)
		return false
	} else {
		if verbose {
			fmt.Printf("[%s] is located at [%s]\n", bin, path)
		}
		return true
	}
}

func move_to_repo_root() {
	log := execute("git", []string{"rev-parse", "--show-cdup"}, verboseMode)
	relative := strings.Trim(log, "\n")
	if len(relative) == 0 {
		return
	}
	if verboseMode {
		fmt.Printf(" > repo root is [%s]\n", relative)
	}
	os.Chdir(relative)
	return
}

func get_diff_list(target string) []string {
	ret := []string{}
	log := execute("git", []string{"diff", "--name-only", target}, verboseMode)
	for _, s := range strings.Split(log, "\n") {
		fmt.Println(s)
		if len(s) > 0 {
			ret = append(ret, s)
		}
	}
	return ret
}

func file_exists(path string) (bool, os.FileInfo) {
	s, e := os.Stat(path)
	return e == nil, s
}

func make_export_directory(dest, path string) {
	fullpath := filepath.Join(dest, path)
	if verboseMode {
		fmt.Printf(" > mkdir [%s]\n", fullpath)
	}
	err := os.MkdirAll(fullpath, 0777)
	if err != nil {
		panic(err)
	}
}

func export_entry(dest, path string) {
	make_export_directory(dest, filepath.Dir(path))
	fullpath := filepath.Join(dest, path)
	err := os.Link(path, fullpath)
	if err != nil {
		panic(err)
	}
	fmt.Printf(">> export [%s]\n", path)
}

func git_export_entry(dest, path string) {
	exists, stat := file_exists(path)
	if !exists {
		fmt.Printf("!! skipped [%s]\n", path)
		return // `diff` may contain removed entry
	}
	if stat.IsDir() {
		make_export_directory(dest, path)
	} else {
		export_entry(dest, path)
	}
}

func git_branch_export(branch, dest string) {
	move_to_repo_root()
	entries := get_diff_list(branch)
	for _, entry := range entries {
		git_export_entry(dest, entry)
	}
}

func entry(opt *Options) error {
	if len(os.Args) < 2 {
		return errors.New("please specify branch and export path")
	}
	branch := os.Args[1]
	dest_root := os.Args[2]
	fmt.Printf(">> branch[%s] export to [%s]\n", branch, dest_root)

	verboseMode = opt.verbose
	git_branch_export(branch, dest_root)

	return nil
}

func main() {
	opt := parse_option()
	entry(opt)
}
