package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	config   *Config
	info     *log.Logger
	pdfparse string
	metadata Metadata = Metadata{
		Name:        "{$name}",
		Version:     "{$version}",
		Description: "./README.md",
		Copyright:   "Copyright 2017 Holmes Group LLC",
		License:     "./LICENSE",
	}
)

//Result structs
type Result struct {

}

// Config structs
type Setting struct {
	HTTPBinding string `json:"HTTPBinding"`
}

type {$name_toUpper} struct {

}

type Config struct {
	Settings			Setting  `json:"settings"`
	{$name_capital}    {$name_toUpper} `json:"{$name}"`
}

type Metadata struct {
	Name        string
	Version     string
	Description string
	Copyright   string
	License     string
}

func main() {

	var (
		err        error
		configPath string
	)
	info = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

	flag.StringVar(&configPath, "config", "", "Path to the configuration file")
	flag.Parse()

	config, err = load_config(configPath)
	if err != nil {
		log.Fatalln("Couldn't decode config file without errors!", err.Error())
	}

	router := httprouter.New()
	router.GET("/analyze/", handler_analyze)
	router.GET("/", handler_info)
	info.Printf("Binding to %s\n", config.Settings.HTTPBinding)
	log.Fatal(http.ListenAndServe(config.Settings.HTTPBinding, router))
}

func handler_info(f_response http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(f_response, `<p>%s - %s</p>
		<hr>
		<p>%s</p>
		<hr>
		<p>%s</p>
		`,
		metadata.Name,
		metadata.Version,
		metadata.Description,
		metadata.License)
}

func load_config(configPath string) (*Config, error) {
	config := &Config{}

	// if no path is supplied look in the current dir
	if configPath == "" {
		configPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		configPath += "/service.conf"
	}

	cfile, _ := os.Open(configPath)
	if err := json.NewDecoder(cfile).Decode(&config); err != nil {
		return config, err
	}

	if metadata.Description != "" {
		if data, err := ioutil.ReadFile(string(metadata.Description)); err == nil {
			metadata.Description = strings.Replace(string(data), "\n", "<br>", -1)
		}
	}

	if metadata.License != "" {
		if data, err := ioutil.ReadFile(string(metadata.License)); err == nil {
			metadata.License = strings.Replace(string(data), "\n", "<br>", -1)
		}
	}

	return config, nil
}

func handler_analyze(f_response http.ResponseWriter, request *http.Request, params httprouter.Params) {
	obj := request.URL.Query().Get("obj")
	if obj == "" {
		http.Error(f_response, "Missing argument 'obj'", 400)
		return
	}
	sample_path := "/tmp/" + obj
	if _, err := os.Stat(sample_path); os.IsNotExist(err) {
		http.NotFound(f_response, request)
		info.Printf("Error accessing sample (file: %s):", sample_path)
		info.Println(err)
		return
	}

	result := &Result{}
/************ ADD your Service logic *******************
*
*
*
*
*
*
*
*
*
*******************************************************/

		f_response.Header().Set("Content-Type", "text/json; charset=utf-8")
	json2http := json.NewEncoder(f_response)

	if err := json2http.Encode(result); err != nil {
		http.Error(f_response, "Generating JSON failed", 500)
		return
	}
}
