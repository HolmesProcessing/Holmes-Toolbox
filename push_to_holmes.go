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
	"path/filepath"
	"time"

	"gopkg.in/mgo.v2/bson"
	"sync"

	"encoding/json"
	"fmt"
	"github.com/rakyll/magicmime"
	"golang.org/x/crypto/ssh/terminal"
	"net/url"
	"strconv"
	"strings"
)

type critsSample struct {
	Id  bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	MD5 string        `json:"md5"`
}

type Task struct {
	PrimaryURI   string              `json:"primaryURI"`
	SecondaryURI string              `json:"secondaryURI"`
	Filename     string              `json:"filename"`
	Tasks        map[string][]string `json:"tasks"`
	Tags         []string            `json:"tags"`
	Attempts     int                 `json:"attempts"`
	Source       string              `json:"source"`
	Download     bool                `json:"download"`
	Comment      string              `json:"comment"`
}

var (
	client          *http.Client
	critsFileServer string
	directory       string
	comment         string
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
	tagsStr    string
	tags       []string
	gatewayURI string
	username   string
	password   string
	tasking    bool

	debug   *log.Logger
	info    *log.Logger
	warning *log.Logger
)

func worker() {
	for true {
		sample := <-c
		debug.Printf("Working on %s\n", sample)
		copySample(sample)
		wg.Done()
	}
}

func main() {
	log.Println("Preparing...")

	// cmd line flags
	flag.StringVar(&fPath, "file", "", "List of samples (MD5, SHAX, CRITs ID) to upload. Files are first searched locally. If they are not found and a CRITs file server is specified, they are taken from there. (optional)")
	flag.StringVar(&comment, "comment", "", "Comment of submitter")
	flag.StringVar(&source, "src", "", "Source information for the files")
	flag.BoolVar(&insecure, "insecure", false, "If set, disables certificate checking")
	flag.BoolVar(&tasking, "tasking", false, "Specify whether to do sample upload or tasking")
	flag.StringVar(&username, "user", "", "Your username for authenticating to the master-gateway.")
	flag.StringVar(&password, "pw", "", "Your password for authenticating to the master-gateway. If this value is not set, you will be prompted for it.")
	flag.StringVar(&gatewayURI, "gateway", "", "The URI of the master-gateway.")
	flag.StringVar(&tagsStr, "tags", "", "The tags for these tasks.")

	// object specific
	flag.StringVar(&critsFileServer, "cfs", "", "Full URL to your CRITs file server, as a fallback (optional)")
	flag.StringVar(&mimetypePattern, "mime", "", "Only upload files with the specified mime-type (as substring)")
	flag.StringVar(&directory, "dir", "", "Directory of samples to upload")
	flag.IntVar(&numWorkers, "workers", 1, "Number of parallel workers")
	flag.BoolVar(&recursive, "rec", false, "If set, the directory specified with \"-dir\" will be iterated recursively")

	// tasking specific
	flag.StringVar(&tasks, "tasks", "", "The tasks to execute.")

	flag.Parse()

	// setup logging
	warning = log.New(os.Stderr, "\033[31m[WARNING]\033[0m ", log.Ldate|log.Ltime|log.Lshortfile)
	info = log.New(os.Stdout, "\033[92m[INFO]\033[0m ", log.Ldate|log.Ltime)
	debug = log.New(os.Stdout, "\033[34m[DEBUG]\033[0m ", log.Ldate|log.Ltime|log.Lshortfile)

	err := json.Unmarshal([]byte(tagsStr), &tags)
	if err != nil {
		warning.Fatal("Error while parsing list of tags!", err)
	}

	// if no password is given via arg ask for it here
	if password == "" {
		println("Please input your password for the master-gateway: ")
		pw, err := terminal.ReadPassword(0)
		if err != nil {
			warning.Fatal("Error reading password from terminal:", err)
		}
		password = string(pw)
	}

	// setup global http client
	tr := &http.Transport{}
	if insecure {
		// Disable SSL verification
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client = &http.Client{Transport: tr}

	// decide to add new tasks OR upload objects
	if tasking {
		main_tasking()
	} else {
		main_object()
	}

	info.Println("==================")
	info.Println("Finished execution")
}

func main_tasking() {
	info.Println("Doing tasking...")

	allTasks := make([]Task, 0)
	task := &Task{PrimaryURI: "", SecondaryURI: "", Filename: "", Tasks: nil, Tags: tags, Attempts: 0, Source: "", Comment: comment, Download: true}

	file, err := os.Open(fPath)
	if err != nil {
		warning.Fatal("Couln't open file containing sample list:", err.Error())
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	err = json.Unmarshal([]byte(tasks), &task.Tasks)
	if err != nil {
		warning.Fatal("Error while parsing list of tasks:", err)
	}

	// line by line
	for scanner.Scan() {
		t := scanner.Text()
		fmt.Sscanf(t, "%s %s %s", &task.PrimaryURI, &task.Filename, &task.Source)
		allTasks = append(allTasks, *task)
	}
	file.Close()

	jsoned, err := json.Marshal(allTasks)
	if err != nil {
		warning.Fatal("Failed to marshal allTasks:", err)
	}

	debug.Printf("All tasks packed: %+v\n", string(jsoned))

	data := &url.Values{}
	data.Set("task", string(jsoned))
	data.Add("username", username)
	data.Add("password", password)

	req, err := http.NewRequest("POST", gatewayURI+"/task/", bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)
	if err != nil {
		warning.Fatal("Error sending allTasks: ", err)
	}

	tskerrors, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		warning.Fatal("Error reading allTasks response: ", err)
	}

	if string(tskerrors) == "" {
		info.Println("The server returned an empty string (success)")
	} else {
		warning.Println("The server returned the following errors:")
		warning.Println(string(tskerrors))
	}
}

func main_object() {
	info.Println("Uploading objects...")

	c = make(chan string)
	for i := 0; i < numWorkers; i++ {
		debug.Printf("Starting worker #%d\n", i)
		go worker()
	}

	if fPath != "" {
		file, err := os.Open(fPath)
		if err != nil {
			warning.Println("Couln't open file containing sample list!", err.Error())
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
			warning.Println("path error:", err)
			return
		}
		topLevel = true
		err = filepath.Walk(fullPath, walkFn)
		if err != nil {
			warning.Println("walk error:", err)
			return
		}
	}

	wg.Wait()
}

func walkFn(path string, fi os.FileInfo, err error) error {
	if fi.IsDir() {
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
		warning.Println("mimetype error (skipping "+path+"):", err)
		return nil
	}
	if strings.Contains(mimetype, mimetypePattern) {
		info.Println("Adding " + path + " (" + mimetype + ")")
		wg.Add(1)
		c <- path
		return nil
	} else {
		info.Println("Skipping " + path + " (" + mimetype + ")")
		return nil
	}
}

func copySample(name string) {
	// set all necessary parameters
	parameters := url.Values{}
	//"user_id": user id of uploader; is filled in by Gateway based on the specified username
	parameters.Add("source", source)            // (TODO) Gateway should match existing sources (command line argument)
	parameters.Add("name", filepath.Base(name)) // filename
	parameters.Add("date", time.Now().Format(time.RFC3339))
	parameters.Add("comment", comment) // comment from submitter (command line argument)
	parameters["tags"] = tags
	parameters.Add("username", username)
	parameters.Add("password", password)

	request, err := buildRequest(gatewayURI+"/samples/", parameters, name)
	if err != nil {
		warning.Fatal("buildRequest failed:", err.Error())
	}

	resp, err := client.Do(request)
	if err != nil {
		warning.Fatal("sending sample request failed:", err.Error())
	}

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		warning.Fatal("reading sample request response failed:", err.Error())
	}
	SafeResponseClose(resp)
	//resp.Body.Close()

	info.Println("-----------------------------------------------------")
	info.Println("Uploaded: ", name)
	info.Println("Resp.Code:", resp.StatusCode)
	info.Println("Resp.Body:", body)
	info.Println("-----------------------------------------------------")
}

func buildRequest(uri string, params url.Values, hash string) (*http.Request, error) {
	debug.Println("Building request...")

	var r io.Reader

	// check if local file
	r, err := os.Open(hash)
	defer r.(*os.File).Close()

	if err != nil {
		debug.Println("Found non local file", hash)

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

	for key, valMul := range params {
		for _, val := range valMul {
			err = writer.WriteField(key, val)
			if err != nil {
				return nil, err
			}
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
