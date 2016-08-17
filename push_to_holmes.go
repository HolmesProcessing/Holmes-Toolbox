package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"
	"path/filepath"

	"gopkg.in/mgo.v2/bson"
	"sync"

	"github.com/rakyll/magicmime"
	"strings"
	"golang.org/x/crypto/ssh/terminal"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type critsSample struct {
	Id  bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	MD5 string        `json:"md5"`
}

type Task struct {
	PrimaryURI     string              `json:"primaryURI"`
	SecondaryURI   string              `json:"secondaryURI"`
	Filename       string              `json:"filename"`
	Tasks          map[string][]string `json:"tasks"`
	Tags           []string            `json:"tags"`
	Attempts       int                 `json:"attempts"`
	Source         string              `json:"source"`
}

var (
	client          *http.Client
	critsFileServer string
	directory       string
	comment         string
	userid          string
	source          string
	mimetypePattern string
	topLevel        bool
	recursive       bool
	insecure        bool
	numWorkers      int
	wg              sync.WaitGroup
	c               chan string

	fPath      string
	tasks      string
	tags       string
	gatewayURI string
	username   string
	password   string
	tasking    bool
)

func init() {
	// http client
	tr := &http.Transport{}
	/*
		// Disable SSL verification
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	*/

	client = &http.Client{Transport: tr}
}

func worker() {
	for true{
		sample :=<- c
		//log.Printf("Working on %s\n", sample)
		copySample(sample)
		wg.Done()
	}
}

func main() {
	log.Println("Running...")

	// cmd line flags
	flag.StringVar(&fPath, "file", "", "List of samples (MD5, SHAX, CRITs ID) to upload. Files are first searched locally. If they are not found and a CRITs file server is specified, they are taken from there. (optional)")
	flag.StringVar(&comment, "comment", "", "Comment of submitter")
	flag.StringVar(&source, "src", "", "Source information for the files")
	flag.BoolVar(&insecure, "insecure", false, "If set, disables certificate checking")
	flag.BoolVar(&tasking, "tasking", false, "Specify whether to do sample upload or tasking")
	flag.StringVar(&username, "user", "", "Your username for authenticating to the master-gateway.")
	flag.StringVar(&password, "pw", "", "Your password for authenticating to the master-gateway. If this value is not set, you will be prompted for it.")
	flag.StringVar(&gatewayURI, "gateway", "", "The URI of the master-gateway.")
	flag.StringVar(&tags, "tags", "", "The tags for these tasks.")

	// object specific
	flag.StringVar(&critsFileServer, "cfs", "", "Full URL to your CRITs file server, as a fallback (optional)")
	flag.StringVar(&mimetypePattern, "mime", "", "Only upload files with the specified mime-type (as substring)")
	flag.StringVar(&directory, "dir", "", "Directory of samples to upload")
	flag.StringVar(&userid, "uid", "-1", "User ID of submitter")
	flag.IntVar(&numWorkers, "workers", 1, "Number of parallel workers")
	flag.BoolVar(&recursive, "rec", false, "If set, the directory specified with \"-dir\" will be iterated recursively")

	// tasking specific
	flag.StringVar(&tasks, "tasks", "", "The tasks to execute.")

	flag.Parse()

	if password == "" {
		println("Please input your password for the master-gateway: ")
		pw, err := terminal.ReadPassword(0)
		if err != nil {
			log.Fatal(err)
		}
		password = string(pw)
	}

	if tasking {
		main_tasking()
	} else {
		main_object()
	}
}

func main_tasking() {
	log.Println("Tasking...")
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
	if err != nil {
		log.Fatal("Error: ", err)
	}
	tskerrors, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	log.Println("The server returned:")
	log.Println(tskerrors)
}

func main_object() {
	log.Println("Object Uploading...")
	c = make(chan string)
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	if fPath != "" {
		file, err := os.Open(fPath)
		if err != nil {
			log.Println("Couln't open file containing sample list!", err.Error())
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		// line by line
		for scanner.Scan() {
			wg.Add(1)
			c <- scanner.Text()
			//go copySample(scanner.Text())
		}
	}
	if directory != "" {
		magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
		defer magicmime.Close()

		fullPath, err := filepath.Abs(directory)

		if err != nil {
			log.Println("path error:", err)
			return
		}
		topLevel = true
		err = filepath.Walk(fullPath, walkFn)
		if err != nil {
			log.Println("walk error:", err)
			return
		}
	}
	wg.Wait()
}

func walkFn(path string, fi os.FileInfo, err error) (error) {
	if fi.IsDir(){
		if recursive {
			return nil
		} else {
			if topLevel {
				topLevel = false
				return nil
			} else {
				return filepath.SkipDir
			}
		}
	}
	mimetype, err := magicmime.TypeByFile(path)
	if err != nil {
		log.Println("mimetype error (skipping " + path + "):", err)
		return nil
	}
	if strings.Contains(mimetype, mimetypePattern) {
		log.Println("Adding " + path + " (" + mimetype + ")")
		wg.Add(1)
		c <- path
		return nil
	} else {
		log.Println("Skipping " + path + " (" + mimetype + ")")
		return nil
	}
}

func copySample(name string) {
	// set all necessary parameters
	parameters := map[string]string{
		"user_id":  userid, // user id of uploader (command line argument)
		"source":   source, // (TODO) Gateway should match existing sources (command line argument)
		"name":     filepath.Base(name), // filename
		"date":     time.Now().Format(time.RFC3339),
		"comment":  comment, // comment from submitter (command line argument)
		"tags":     tags,
		"username": username,
		"password": password,
	}

	request, err := buildRequest(gatewayURI+"/samples/", parameters, name)

	if err != nil {
		log.Fatal("ERROR: " + err.Error())
	}

	tr := &http.Transport{}
	if insecure{
		// Disable SSL verification
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client = &http.Client{Transport: tr}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal("ERROR: " + err.Error())
		return
	}

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		log.Fatal("ERROR: " + err.Error())
	}
	resp.Body.Close()

	log.Println("Uploaded", name)
	log.Println(resp.StatusCode)
	log.Println(body)
	log.Println("-------------------------------------------")
}

func buildRequest(uri string, params map[string]string, hash string) (*http.Request, error) {
	var r io.Reader

	// check if local file
	r, err := os.Open(hash)
	if err != nil {
		// not a local file
		// try to get file from crits file server

		cId := &critsSample{}
		if err := bson.Unmarshal([]byte(hash), cId); err != nil {
			return nil, err
		}
		rawId := cId.Id.Hex()

		resp, err := client.Get(critsFileServer + "/" + rawId)
		if err != nil {
			return nil, err
		}
		defer SafeResponseClose(resp)

		// return if file does not exist
		if resp.StatusCode != 200 {
			return nil, errors.New("Couldn't download file")
		}

		r = resp.Body
		// For files coming from CRITs: TODO: find real name somehow
	}

	// build Holmes-Storage PUT request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("sample", hash)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, r)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		err = writer.WriteField(key, val)
		if err != nil {
			return nil, err
		}
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("PUT", uri, body)
	//request, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", writer.FormDataContentType())

	return request, nil
}

func SafeResponseClose(r *http.Response) {
	if r == nil {
		return
	}

	io.Copy(ioutil.Discard, r.Body)
	r.Body.Close()
}
