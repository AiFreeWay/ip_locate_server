package main

import (
  "./config"
  "./provider"
  "net/http"
  _ "github.com/lib/pq"
  "database/sql"
  "net"
  "fmt"
  "log"
)

var dataBase *sql.DB

func main() {
  config.Init()
  initDb()
  initWebServer()
}

func initDb() {
  db, dbOpenErr := sql.Open("postgres", "postgres://"+config.GetDbUser()+":"+config.GetDbPass()+"@localhost/iplocatedb")

  if dbOpenErr != nil {
    panic(dbOpenErr)
  } else {
    dataBase = db
  }
}

func initWebServer() {
  http.HandleFunc("/", detectLocateByIp)
  log.Println("Listening 8217 port...")
  http.ListenAndServe(":8217", nil)
}

func detectLocateByIp(responseWriter http.ResponseWriter, request *http.Request) {
  ip := getIp(request)
  ignoredIps := config.GetIgnoredIps()

  for _, ignoreIp := range ignoredIps {
    if ip == ignoreIp {
      fmt.Fprintf(responseWriter, "You have ignored ip")
      return
    }
  }

  provider := selectProvider()

  if provider == nil {
    fmt.Fprintf(responseWriter, "All providers unavailable")
    log.Println("All providers unavailable")
    return
  }

  locateResponse, locateRespErr := provider.RequestLocate(ip)

  if locateRespErr == nil {
    defer locateResponse.Body.Close()
    countryCode := provider.GetCountryCodeFromResponse(locateResponse)

    cacheIpLocate(ip, countryCode)
    fmt.Fprintf(responseWriter, "Hello! You country is "+countryCode)
    log.Println("Connected by ip "+ip)

  } else {
    fmt.Fprintf(responseWriter, "Internal server error")
    log.Println("Internal server error")
  }
}

func getIp(request *http.Request) string {
  ip, _, _ := net.SplitHostPort(request.RemoteAddr)
  return net.ParseIP(ip).String()
}

func selectProvider() provider.Provider {
  for _, provider := range config.GetProviders() {
    if provider.IsLimitAvailable() {
      return provider
    }
  }

  return nil
}

func cacheIpLocate(ip, countryCode string) {
  dataBase.Exec("INSERT INTO users_ip_locate VALUES ($1, $2)", ip, countryCode)
}
