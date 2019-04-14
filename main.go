package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

var (
	port     string
	endpoint string
	url      string
	pooltime int
)

type tunnel struct {
	URL  string
	Quit bool
}

func logToFile(msg string) {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 066)
	if err != nil {
		log.Fatalf("Error writing log file: %s", err.Error())
	} else {
		log.SetOutput(f)
		log.Println(msg)
	}
	defer f.Close()

	log.SetOutput(os.Stdout)
}

// Keep send get request to see if tunnel is alive
func (s *tunnel) KeepAlive() {
	go func() {
		log.Println("Pooling at URL: " + url + "/" + endpoint)
		for {
			r, err := http.Get(s.URL + "/" + endpoint)
			if err != nil {
				logToFile(err.Error())
				log.Println(err.Error())
				os.Exit(3)
			} else {
				defer r.Body.Close()
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println(err.Error())
				} else {
					if len(content) == 0 {
						msg := "Tunnel [" + url + "] is down"
						logToFile(msg)
						log.Println(msg)
					} else {
						log.Println(string(content))
					}
				}
			}
			time.Sleep(time.Minute * time.Duration(pooltime))
		}
	}()
}

// Set default configurations
func setDefaultConfig() {
	pooltime = 10
	endpoint = "__ping"
	url = ""
}

// Check for config file
func checkIniConfig() {
	if _, err := os.Stat("config.ini"); err != nil {
		f, err := os.Create("config.ini")
		if err != nil {
			f.Close()
			panic(err)
		} else {
			setDefaultConfig()
			f.Write([]byte("[tunnel]\nendpoint=" + endpoint + "\nurl=\npooltime=" + strconv.Itoa(pooltime)))
			f.Close()
		}

	} else {
		cfg, err := ini.Load("config.ini")
		if err != nil {
			log.Println("Can't load config file. Error: ", err.Error())
			setDefaultConfig()
		} else {
			endpoint = cfg.Section("tunnel").Key("endpoint").String()
			url = cfg.Section("tunnel").Key("url").String()
			v, err := cfg.Section("tunnel").Key("pooltime").Int()
			if err != nil {
				pooltime = 10
			} else {
				pooltime = v
			}
		}
	}

	if url == "" {
		reader := bufio.NewReader(os.Stdin)
		print("Insert URL to ping: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		println(text)
		text = strings.Replace(text, "\"", "", -1)
		text = strings.Trim(text, "\"\"\"")
		text = strings.Replace(text, "\n", "", -1)
		url = text

		cfg, err := ini.Load("config.ini")
		if err == nil {
			cfg.Section("tunnel").Key("url").SetValue(url)
			cfg.SaveTo("config.ini")
		}
	}
}

func main() {
	done := make(chan bool)
	checkIniConfig()

	s := tunnel{URL: url}
	s.KeepAlive()

	<-done
}
