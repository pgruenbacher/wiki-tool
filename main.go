package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pgruenbacher/log"
)

const (
	start string = "WIKI_MENUS_START"
	end   string = "END"
	title string = "[menu.wiki]"
)

type fileList []fileI

type fileI struct {
	path string
	info os.FileInfo
}

type menuEntry struct {
	title      string
	name       string
	parent     string
	identifier string
}

func main() {
	searchDirPtr := flag.String("directory", "content/wiki", "the directory for the wiki")
	searchWordPtr := flag.String("replace", "DIRECTORY", "the string to replace")
	configFilePtr := flag.String("config", "src/config.toml", "path to config file")
	flag.Parse()
	log.Info("searching: %s", *searchDirPtr)

	fileList := make([]fileI, 0)

	err := filepath.Walk(*searchDirPtr, func(path string, f os.FileInfo, err error) error {
		fileInfo := fileI{
			path: path,
			info: f,
		}
		fileList = append(fileList, fileInfo)
		return nil
	})

	if err != nil {
		fmt.Println(err)
	}

	menuList := make([]menuEntry, 0)
	for _, file := range fileList {
		array := strings.Split(file.path, "/")
		if len(array) > 1 {
			parent := array[len(array)-1]

			// skip if parent directory is original wiki
			if parent == *searchDirPtr {
				continue
			}
			// for a parent menu
			var grandparent string
			wikiIndex := indexOf(array, "wiki")
			if (len(array) - wikiIndex) > 2 {
				grandparent = array[len(array)-2]
			} else {
				grandparent = ""
			}
			if !within(menuList, parent) && file.info.IsDir() {
				entry := menuEntry{
					title:      title,
					name:       parent,
					parent:     grandparent,
					identifier: parent,
				}
				menuList = append(menuList, entry)
			}

			if !file.info.IsDir() {
				searchAndReplace(file.path, grandparent, *searchWordPtr)
			}

		}

	}
	writeConfig(menuList, *configFilePtr)
	// for _, e := range menuList {
	// 	fmt.Println(e)
	// }
	return
}

func writeConfig(menuList []menuEntry, configFile string) {
	output := make([]string, 0)
	for _, menu := range menuList {
		output = append(output, "[[menu.wiki]]\n")
		output = append(output, fmt.Sprintf("name=\"%s\"\n", menu.name))
		output = append(output, fmt.Sprintf("parent=\"%s\"\n", menu.parent))
		output = append(output, fmt.Sprintf("identifier=\"%s\"\n", menu.name))
	}

	lines, err := readLines(configFile)
	if err != nil {
		fmt.Println(err)
	}

	var st int
	var en int
	// delete all entries within first
	for i, line := range lines {
		if strings.Contains(line, start) {
			st = i + 1
		}
		if strings.Contains(line, end) {
			en = i

		}
	}
	if st > 0 && en > 0 {
		lines = append(lines[:st], lines[en:]...)
	}
	// insert
	lines = append(lines[:st], append(output, lines[st:]...)...)
	// write
	log.Info("writing to config.toml")
	err = writeLines(configFile, lines)
	if err != nil {
		fmt.Println(err)
	}
}

func searchAndReplace(path, parent, searchWordPtr string) {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	//fmt.Println(string(read))
	if strings.Contains(string(read), searchWordPtr) {
		fmt.Println("replacing:", path, "parent='DIRECTORY' with:", parent)

		newContents := strings.Replace(string(read), searchWordPtr, parent, -1)
		err = ioutil.WriteFile(path, []byte(newContents), 0)
		if err != nil {
			panic(err)
		}
	}

}

func indexOf(array []string, index string) int {
	for i, entry := range array {
		if entry == index {
			return i
		}
	}
	return -1
}

func within(array []menuEntry, parent string) bool {
	for _, entry := range array {
		if entry.name == parent {
			return true
		}
	}
	return false
}

func readLines(file string) (lines []string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		const delim = '\n'
		line, err := r.ReadString(delim)
		if err == nil || len(line) > 0 {
			if err != nil {
				line += string(delim)
			}
			lines = append(lines, line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return lines, nil
}

func writeLines(file string, lines []string) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	for _, line := range lines {
		_, err := w.WriteString(line)
		if err != nil {
			return err
		}
	}
	return nil
}
