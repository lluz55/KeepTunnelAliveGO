package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/ini.v1"
)

var (
	port     string
	endpoint string
	url      string
	pooltime int
)

type server struct {
	URL  string
	Quit bool
}

func (s *server) KeepAlive() {
	go func() {
		log.Println("Pooling at URL: " + url + "/" + endpoint)
		for {
			r, err := http.Get(s.URL + "/" + endpoint)
			if err != nil {
				log.Println(err.Error())
			} else {
				defer r.Body.Close()
				content, err := ioutil.ReadAll(r.Body)
				if err != nil {
					log.Println(err.Error())
				} else {
					if len(content) == 0 {
						log.Println("Tunnel [" + url + "] is down")
					} else {
						log.Println(string(content))
					}
				}
			}
			time.Sleep(time.Minute * time.Duration(pooltime))
		}
	}()
}

func (s *server) Wait() {
	for !s.Quit {
		time.Sleep(time.Second * 10)
	}
}

func (s *server) StartServer() {
	go func() {
		http.HandleFunc("/"+endpoint, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(strconv.Itoa(int(time.Now().Unix()))))
		})

		log.Println("Serving at port " + port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			panic(err.Error())
		}
	}()
}

func setDefaultConfig() {
	port = "1302"
	pooltime = 10
	endpoint = "__ping"
}

func checkIniConfig() {
	if _, err := os.Stat("config.ini"); err != nil {
		f, err := os.Create("config.ini")
		if err != nil {
			f.Close()
			panic(err)
		} else {
			setDefaultConfig()
			f.Write([]byte("[server]\nport=" + port + "\nendpoint=" + endpoint + "\npooltime=" + strconv.Itoa(pooltime)))
			f.Close()
		}

	} else {
		cfg, err := ini.Load("config.ini")
		if err != nil {
			log.Println("Can't load config file. Error: ", err.Error())
			setDefaultConfig()
		} else {
			port = cfg.Section("server").Key("port").String()
			endpoint = cfg.Section("server").Key("endpoint").String()
			v, err := cfg.Section("server").Key("pooltime").Int()
			if err != nil {
				pooltime = 10
			} else {
				pooltime = v
			}
		}
	}

	url = os.Getenv("KTA_URL")
	if url == "" {
		log.Println("Environment Variable 'KTA_URL' not founded")
		os.Exit(1)
	}
}

func main() {
	checkIniConfig()

	s := server{URL: url}
	s.StartServer()
	s.KeepAlive()
	s.Wait()
}
