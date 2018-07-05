package main

import (
	"fmt"
	"github.com/containous/flaeg"
	"github.com/containous/staert"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Redis  string `description:"Redis server"`
	Prefix string `description:"Download counter key prefix"`
	Bind   string `description:"HTTP bind string"`
}

type CliConfig struct {
	Config
	ConfigFile string `description:"Configuration file to use (TOML)."`
}

type Stat struct {
	status   int
	path     string
	complete string
}

var (
	pool *redis.Pool
	conf Config
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	status := 400
	if stat, err := strconv.Atoi(r.Header.Get("X-Track-Status")); err == nil {
		status = stat
	}

	path := r.Header.Get("X-Track-URI")
	complete := r.Header.Get("X-Track-Complete")

	if 200 <= status && status < 300 && complete == "OK" && !strings.HasSuffix(path, "/") {
		conn := pool.Get()
		defer conn.Close()
		key := conf.Prefix + strings.Replace(path, "/", ":", -1)
		conn.Do("INCR", key)
	}

	w.WriteHeader(204)
}

func run() {
	pool = newPool(conf.Redis)

	http.HandleFunc("/", handler)

	fmt.Println("Binding to", conf.Bind)
	http.ListenAndServe(conf.Bind, nil)
}

func main() {
	config := &CliConfig{
		Config: Config{
			Redis:  ":6379",
			Prefix: "download",
			Bind:   ":8725",
		},
		ConfigFile: "settings",
	}

	rootCmd := &flaeg.Command{
		Name:                  "dl-stats",
		Description:           "dl-stats keeps download stats for nginx post_action notifications",
		Config:                &config,
		DefaultPointersConfig: &config,
		Run: func() error {
			conf = config.Config
			run()
			return nil
		},
	}

	f := flaeg.New(rootCmd, os.Args[1:])
	if _, err := f.Parse(rootCmd); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	s := staert.NewStaert(rootCmd)
	toml := staert.NewTomlSource("settings", []string{config.ConfigFile, "."})
	s.AddSource(toml)
	s.AddSource(f)

	if _, err := s.LoadConfig(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := s.Run(); err != nil {
		fmt.Println(err.Error())
	}
}
