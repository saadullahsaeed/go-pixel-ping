package main

import (
  "log"
  "fmt"
  "net/http"
  "time"
  "encoding/json"
  "net/url"
  "os"
  "io/ioutil"
)


type Config struct {
  Port int `json:"port"`
  Host string `json:"host"`
  Interval time.Duration `json:"interval"`
  Endpoint string `json:"endpoint"`
  Discard bool `json:"discard"`
}

func (c *Config) Parse(path string) {
  file, err := ioutil.ReadFile(path)
  if err != nil {
     log.Fatal("File error: %v\n", err)
  }
  json.Unmarshal([]byte(file), &c)
}


type Store struct {
  Data map[string]int
}

func (s *Store) reset() {
  if len(s.Data) > 0 {
    for key, _ := range s.Data {
      delete(s.Data, key)
    }
  }
}

func (s *Store) flush() {
  if len(s.Data) > 0 {
    store_json, _ := json.Marshal(s.Data)
    log.Println(string(store_json))

    values := make(url.Values)
    values.Set("json", string(store_json))

    resp, err := http.PostForm(config.Endpoint, values)
    if err != nil {
      log.Println(err)
    }
    defer resp.Body.Close()
  }
}

func (s *Store) register(key string) {
  if s.Data == nil {
    s.Data = make(map[string]int)
  }
  
  if s.Data[key] == 0 {
    s.Data[key] = 1
  } else {
    s.Data[key]++
  }
}


var (
  store Store
  config Config
)


func pingHandler(w http.ResponseWriter, r *http.Request) {  
  key := r.FormValue("key")
  if key == "" {
    return
  }
  store.register(key)
}


func flusher() {
  log.Println(fmt.Sprintf("Flushing hits to %v", config.Endpoint))
  for {
    store.flush()
    store.reset()
    time.Sleep(time.Second * config.Interval)
  }
}


func main() {
  config.Parse(os.Args[1])
  go flusher()
  http.HandleFunc("/", pingHandler)
  http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}