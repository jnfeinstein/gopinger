package main

import (
	"fmt"
	"github.com/go-av/curl"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"gopinger/config"
	"net/http"
	"time"
)

type Site struct {
	Ip      string
	Exit    chan bool
	Success int
	Error   int
}

func (s *Site) Curl() {
	url := fmt.Sprintf("http://%s", s.Ip)
	err, _ := curl.String(url, "timeout=", time.Second*10, "deadline=", time.Second*10)
	if err != nil {
		fmt.Printf("ERROR - %s : %s\n", s.Ip, err.Error())
		s.Error++
	} else {
		fmt.Printf("SUCCESS - %s\n", s.Ip)
		s.Success++
	}
}

func (s *Site) Ping() {
	fmt.Printf("%s\n", s.Ip)
	s.Curl()
	for {
		exit := false

		select {
		case <-s.Exit:
			exit = true
			fmt.Println("exit")
		case <-time.After(time.Minute * 1):
			s.Curl()
		}
		if exit {
			break
		}
	}
}

func (s *Site) Stats() string {
	return fmt.Sprintf("%s - %d SUCCESS %d ERROR", s.Ip, s.Success, s.Error)
}

type SiteMap map[string]*Site

func (s *SiteMap) AddSite(ip string) *Site {
	site := &Site{ip, make(chan bool), 0, 0}
	(*s)[ip] = site
	go site.Ping()
	return site
}

func (s *SiteMap) RemoveSite(ip string) *Site {
	if site, ok := (*s)[ip]; ok {
		site.Exit <- true
		delete((*s), ip)
		return site
	}
	return nil
}

func (s *SiteMap) QuerySite(ip string) string {
	if site, ok := (*s)[ip]; ok {
		return site.Stats()
	} else {
		return "Site not found"
	}
}

func main() {
	m := martini.Classic()

	config.Initialize(m)

	m.Use(render.Renderer())

	sites := make(SiteMap)

	m.Get("/", func(r render.Render, req *http.Request) {
		r.HTML(200, "index", req.Host)
	})

	m.Get("/add/:ip", func(p martini.Params) string {
		ip := p["ip"]
		sites.AddSite(ip)
		return "OK\n"
	})

	m.Get("/remove/:ip", func(p martini.Params) string {
		ip := p["ip"]
		sites.RemoveSite(ip)
		return "OK\n"
	})

	m.Get("/query/:ip", func(p martini.Params) string {
		ip := p["ip"]
		return fmt.Sprintf("%s\n", sites.QuerySite(ip))
	})

	m.Get("/dump", func(p martini.Params) string {
		result := ""
		for _, site := range sites {
			result += fmt.Sprintf("%s\n", site.Stats())
		}
		return result
	})

	sites.AddSite(config.Url())

	m.Run()
}
