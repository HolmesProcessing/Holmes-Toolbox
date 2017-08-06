package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Config
type Config struct {
	Type        string
	Language    string
	Name        string
	Version     string
	Description string
}

var config *Config

func main() {
	var (
		err        error
		configPath string
	)

	flag.StringVar(&configPath, "config", "", "Path to the configuration file")
	flag.Parse()
	config, err = load_config(configPath)
	if err != nil {
		panic(err.Error())
	}

	servicename := config.Name
	lang := config.Language
	// create a new directory for the service.
	filenames := []string{"service.conf", "Dockerfile", "README.md", "serviceREST.scala"}
	createDir(servicename)
	for i := 0; i < 4; i++ {
		dest := createFile(servicename, filenames[i], lang)
		parseAndReplace(servicename, dest)
	}
	if lang == "go" {
		dest := createFile(servicename, "service.go", lang)
		parseAndReplace(servicename, dest)
	} else {
		dest := createFile(servicename, "service.py", lang)
		parseAndReplace(servicename, dest)
	}
}

func load_config(configPath string) (*Config, error) {
	config := &Config{}

	if configPath == "" {
		configPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		configPath += "/parse.conf"
	}

	cfile, _ := os.Open(configPath)
	if err := json.NewDecoder(cfile).Decode(&config); err != nil {
		return config, err
	}

	return config, nil
}

func Check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func createDir(service_name string) {
	name := service_name
	os.Mkdir(name, 0700)
}

func createFile(service_name, filename, lang string) string {
	lang = strings.ToLower(lang)
	src := lang + "/" + filename + ".tpl"
	dest := service_name + "/" + filename
	sFile, err := os.Open(src)
	Check(err)

	eFile, err := os.Create(dest)
	Check(err)

	_, err = io.Copy(eFile, sFile)
	if err != nil {
		log.Fatal(err)
	}
	err = eFile.Sync()
	if err != nil {
		log.Fatal(err)
	}

	return dest
}

func readFile(dest string) (input []byte, err error) {
	input, err = ioutil.ReadFile(dest) // or anyfile
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	return input, err
}

func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func parseAndReplace(servicename, dest string) {
	// replace a word other name
	version := config.Version
	description := config.Description

	upper := strings.ToUpper(servicename)
	ucfirst := UcFirst(servicename)

	input, err := readFile(dest)
	output := bytes.Replace(input, []byte("{$name}"), []byte(servicename), -1)
	if err = ioutil.WriteFile(dest, output, 0666); err != nil {
		Check(err)
		os.Exit(1)
	}

	input, err = readFile(dest)
	output = bytes.Replace(input, []byte("{$name_toUpper}"), []byte(upper), -1)
	if err = ioutil.WriteFile(dest, output, 0666); err != nil {
		Check(err)
		os.Exit(1)
	}

	input, err = readFile(dest)
	output = bytes.Replace(input, []byte("{$name_capital}"), []byte(ucfirst), -1)
	if err = ioutil.WriteFile(dest, output, 0666); err != nil {
		Check(err)
		os.Exit(1)
	}

	input, err = readFile(dest)
	output = bytes.Replace(input, []byte("{$version}"), []byte(version), -1)
	if err = ioutil.WriteFile(dest, output, 0666); err != nil {
		Check(err)
		os.Exit(1)
	}

	input, err = readFile(dest)
	output = bytes.Replace(input, []byte("{$description}"), []byte(description), -1)
	if err = ioutil.WriteFile(dest, output, 0666); err != nil {
		Check(err)
		os.Exit(1)
	}
}
