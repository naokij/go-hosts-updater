package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const split = "##GO HOSTS##"
const goHostsURL = "https://github.com/Lerist/Go-Hosts/raw/master/hosts"

var etcHostsPath = hostsPath()
var temp = tmpPath()

func getHosts() (f *os.File, err error) {
	resp, err := http.Get(goHostsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	f, err = ioutil.TempFile(temp, "go-hosts-")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return nil, err
	}
	return
}

func main() {
	goHostsFile, err := getHosts()
	if err != nil {
		log.Println("Error while downloading go hosts file:", err)
		return
	}
	defer os.Remove(goHostsFile.Name())
	defer goHostsFile.Close()

	goHostsFile.Seek(0, 0)
	goHostsScanner := bufio.NewScanner(goHostsFile)

	etcHostsFile, err := os.OpenFile(etcHostsPath, os.O_RDWR, 0644)
	if err != nil {
		log.Println("Error while opening ", etcHostsPath, " - ", err)
		return
	}
	defer etcHostsFile.Close()
	tmpHostsFile, err := ioutil.TempFile(temp, "hosts-")
	if err != nil {
		log.Println("Error while creating ", tmpHostsFile.Name(), " - ", err)
		return
	}
	defer os.Remove(tmpHostsFile.Name())
	defer tmpHostsFile.Close()
	tmpHostsWriter := bufio.NewWriter(tmpHostsFile)

	etcHostsScanner := bufio.NewScanner(etcHostsFile)
	for etcHostsScanner.Scan() {
		if etcHostsScanner.Text() == split {
			break
		}
		line := etcHostsScanner.Text() + "\n"
		if _, err := tmpHostsWriter.WriteString(line); err != nil {
			log.Println("Error while writing ", tmpHostsFile.Name(), " - ", err)
			return
		}
	}
	if etcHostsScanner.Err() != nil {
		log.Println("Error while opening ", etcHostsPath, " - ", err)
		return
	}
	t := time.Now()
	splitLine := fmt.Sprintf("%s\n# Merged: %s\n", split, t.Format(time.RFC3339))
	if _, err := tmpHostsWriter.WriteString(splitLine); err != nil {
		log.Println("Error while writing ", tmpHostsFile.Name(), " - ", err)
		return
	}

	for goHostsScanner.Scan() {
		line := goHostsScanner.Text()
		if strings.HasPrefix(line, "127.0.0.1") || strings.HasPrefix(line, "::1") {
			continue
		}

		if _, err := fmt.Fprintln(tmpHostsWriter, line); err != nil {
			log.Println("Error while writing ", tmpHostsFile.Name(), " - ", err)
			return
		}
	}
	if goHostsScanner.Err() != nil {
		log.Println("Error while writing ", goHostsFile.Name(), " - ", err)
	}
	fmt.Fprintln(tmpHostsWriter, "#END")
	if tmpHostsWriter.Flush() != nil {
		log.Println("Error while writing ", tmpHostsFile.Name(), " - ", err)
		return
	}

	tmpHostsFile.Seek(0, 0)
	etcHostsFile.Seek(0, 0)
	tmpHostsReader := bufio.NewReader(tmpHostsFile)
	etcHostsWriter := bufio.NewWriter(etcHostsFile)
	if _, err := io.Copy(etcHostsWriter, tmpHostsReader); err != nil {
		log.Println("Error while copy ", tmpHostsFile.Name(), " to ", etcHostsFile.Name(), " - ", err)
		return
	}
	if etcHostsWriter.Flush() != nil {
		log.Println("Error while copy ", tmpHostsFile.Name(), " to ", etcHostsFile.Name(), " - ", err)
		return
	}
}
