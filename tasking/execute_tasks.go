package main

import (
	"net/http"
	"net/url"
	"flag"
	"os"
	"log"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"bytes"
	"strconv"
	"crypto/tls"
	"golang.org/x/crypto/ssh/terminal"
)

type Task struct {
	PrimaryURI     string              `json:"primaryURI"`
	SecondaryURI   string              `json:"secondaryURI"`
	Filename       string              `json:"filename"`
	Tasks          map[string][]string `json:"tasks"`
	Tags           []string            `json:"tags"`
	Attempts       int                 `json:"attempts"`
	Source         string              `json:"source"`
}

func main() {
	var (
		fPath      string
		tasks      string
		tags       string
		gatewayURI string
		username   string
		password   string
		insecure   bool
	)

	flag.StringVar(&fPath, "file", "", "List of samples (SHA256) to process.")
	flag.StringVar(&tasks, "tasks", "", "The tasks to execute.")
	flag.StringVar(&tags, "tags", "", "The tags for these tasks.")
	flag.StringVar(&gatewayURI, "gateway", "", "The URI of the master-gateway.")
	flag.StringVar(&username, "user", "", "Your username for authenticating to the master-gateway.")
	flag.StringVar(&password, "pw", "", "Your password for authenticating to the master-gateway. If this value is not set, you will be prompted for it.")
	flag.BoolVar(&insecure, "insecure", false, "Disable SSL certificate checking.")
	flag.Parse()

	if password == "" {
		println("Please input your password for the master-gateway: ")
		pw, err := terminal.ReadPassword(0)
		if err != nil {
			log.Fatal(err)
		}
		password = string(pw)
	}


	allTasks := make([]Task,0)

	file, err := os.Open(fPath)
	if err != nil {
		log.Println("Couln't open file containing sample list!", err.Error())
		return
	}
	defer file.Close()


	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	task := &Task{PrimaryURI:"", SecondaryURI:"", Filename:"", Tasks:nil, Tags:nil, Attempts:0, Source:""}
	err = json.Unmarshal([]byte(tasks), &task.Tasks)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal([]byte(tags), &task.Tags)
	if err != nil {
		log.Fatal(err)
	}

	// line by line
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Sscanf(t, "%s %s %s", &task.PrimaryURI, &task.Filename, &task.Source)
		allTasks = append(allTasks, *task)
	}
	jsoned, err := json.Marshal(allTasks)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v\n", string(jsoned))
	data := &url.Values{}
	data.Set("task", string(jsoned))
	data.Add("username", username)
	data.Add("password", password)
	
	req, err := http.NewRequest("POST", gatewayURI, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	tr := &http.Transport{}
	if insecure{
		// Disable SSL verification
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	tskerrors, _ := ioutil.ReadAll(resp.Body)
	log.Println("The server returned:")
	log.Println(tskerrors)
}
