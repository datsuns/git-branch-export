package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var verboseMode = false

var OptionFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  "verbose, V",
		Usage: "verbose mode",
	},
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

func entry(c *cli.Context) error {
	if c.NArg() < 2 {
		return cli.NewExitError("please specify branch and export path", 86)
	}
	branch := c.Args()[0]
	dest_root := c.Args()[1]
	fmt.Printf(">> branch[%s] export to [%s]\n", branch, dest_root)

	verboseMode = c.GlobalBool("verbose")
	git_branch_export(branch, dest_root)

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "git-branch-export"
	app.Usage = "branch's diff exporter to specified path"
	app.Version = "1.0.0"
	app.Commands = nil
	app.Action = entry
	app.Flags = OptionFlags

	app.Run(os.Args)
}
