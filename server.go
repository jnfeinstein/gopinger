package main

import (
	"database/sql"
	"fmt"
	"github.com/coopernurse/gorp"
	"github.com/go-av/curl"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"gopinger/config"
	"net/http"
	"time"
)

type Database struct {
	Connection *gorp.DbMap
}

func makeDb() (*Database, error) {
	postgres_args := config.PostgresArgs()
	if len(postgres_args) <= 0 {
		return nil, nil // Disables persisting
	}
	db, err := sql.Open("postgres", postgres_args)
	if err != nil {
		return nil, err
	}
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.AddTableWithName(Site{}, "sites").SetKeys(true, "id")
	return &Database{dbmap}, nil
}

func (d *Database) Add(s *Site) error {
	return d.Connection.Insert(s)
}

func (d *Database) Remove(s *Site) error {
	_, err := d.Connection.Delete(s)
	return err
}

func (d *Database) GetAll() ([]Site, error) {
	var sites []Site
	_, err := d.Connection.Select(&sites, "SELECT * FROM sites")
	if err != nil {
		return sites, err
	}
	return sites, nil
}

type Site struct {
	Id      int       `db:"id"`
	Ip      string    `db:"ip"`
	Exit    chan bool `db:"-"`
	Success int       `db:"-"`
	Error   int       `db:"-"`
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

func (s *SiteMap) AddSite(db *Database, ip string) *Site {
	if site, ok := (*s)[ip]; ok {
		return site
	}
	site := &Site{0, ip, make(chan bool), 0, 0}
	(*s)[ip] = site
	if db != nil {
		err := db.Add(site)
		if err != nil {
			fmt.Printf("Error adding to database - %s\n", site.Ip)
			fmt.Printf("%s\n", err.Error())
		}
	}
	go site.Ping()
	return site
}

func (s *SiteMap) RemoveSite(db *Database, ip string) *Site {
	if site, ok := (*s)[ip]; ok {
		site.Exit <- true
		delete((*s), ip)
		if db != nil {
			fmt.Printf("%d\n", site.Id)
			err := db.Remove(site)
			if err != nil {
				fmt.Printf("Error removing from database - %s\n", site.Ip)
				fmt.Printf("%s\n", err.Error())
			}
		}
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
	db, err := makeDb()
	if err != nil {
		fmt.Printf("Error opening database\n")
		fmt.Printf("%s\n", err.Error())
		return
	}

	persistedSites, err := db.GetAll()
	if err != nil {
		fmt.Printf("Error selecting sites from database\n")
		fmt.Printf("%s\n", err.Error())
	}

	for _, s := range persistedSites {
		s.Exit = make(chan bool)
		sites[s.Ip] = &s
		go s.Ping()
	}

	m.Get("/", func(r render.Render, req *http.Request) {
		r.HTML(200, "index", req.Host)
	})

	m.Get("/add/:ip", func(p martini.Params) string {
		ip := p["ip"]
		sites.AddSite(db, ip)
		return "OK\n"
	})

	m.Get("/remove/:ip", func(p martini.Params) string {
		ip := p["ip"]
		sites.RemoveSite(db, ip)
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

	sites.AddSite(nil, config.Url())

	m.Run()
}
