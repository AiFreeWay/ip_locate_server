package provider

import (
  "net/http"
  "io/ioutil"
  "encoding/json"
  "github.com/go-redis/redis"
  "strconv"
  "time"
)

func GetProviderByName(providerName string) Provider {
  switch providerName {
    case "ipstack.com":
      redisClient := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
		    Password: "",
		    DB:       0})
      var ipStackProvider IpstackProvider = IpstackProvider{10000, redisClient}
      return ipStackProvider
    case "geoip.nekudo.com":
      return new(GeoipNekudoProvider)
    default: return nil
  }
}

// Providers

type Provider interface {

  RequestLocate(ip string) (*http.Response, error)
  GetCountryCodeFromResponse(response *http.Response) string
  IsLimitAvailable() bool
}

type IpstackProvider struct {
  queryLimit int
  redisClient *redis.Client
}

type GeoipNekudoProvider struct {}

// IpstackProvider

func (p IpstackProvider) RequestLocate(ip string) (*http.Response, error) {
  resp, err := http.Get("http://api.ipstack.com/"+ip+"?access_key=e82eb693d199cf92dc2326df48d69da5")

  currentLimit, _ := p.redisClient.Get("ipstack_com_limit").Result()
  currentLimitInt, _ := strconv.Atoi(currentLimit)
  newLimit := currentLimitInt+1
  p.redisClient.Set("ipstack_com_limit", string(newLimit), 0).Err()

  return resp, err
}

func (p IpstackProvider) GetCountryCodeFromResponse(response *http.Response) string {
  body, err := ioutil.ReadAll(response.Body)

  if err != nil {
    return ""
  }

  var result map[string]interface{}
  json.Unmarshal([]byte(string(body)), &result)

  country_code := result["country_code"]

  return country_code.(string)
}

func (p IpstackProvider) IsLimitAvailable() bool {
  currentLimit, _ := p.redisClient.Get("ipstack_com_limit").Result()
  currentLimitInt, _ := strconv.Atoi(currentLimit)

  if currentLimitInt < p.queryLimit {
    return true
  } else {

   limitEstimate, _ := p.redisClient.Get("ipstack_com_limit_estimate").Result()

    if limitEstimate == "" {
      p.redisClient.Set("ipstack_com_limit_estimate", time.Now().Local().String(), 0).Err()

      return false
    } else {
      tamplate := "2006-01-02T15:04:05.000Z"
      estimateStr, _ := p.redisClient.Get("ipstack_com_limit_estimate").Result()
      parsedEstimate, _ := time.Parse(tamplate, estimateStr)
      now := time.Now().Local()
      diff := (parsedEstimate.Sub(now).Hours())/24/31

      if (diff >= 31) {
        p.redisClient.Set("ipstack_com_limit", string(0), 0).Err()
        p.redisClient.Set("ipstack_com_limit_estimate", "", 0).Err()

        return true
      }
    }

    return false
  }
}

// GeoipNekudoProvider

func (p GeoipNekudoProvider) RequestLocate(ip string) (*http.Response, error){
  resp, err := http.Get("http://geoip.nekudo.com/api/"+ip)
  return resp, err
}

func (p GeoipNekudoProvider) GetCountryCodeFromResponse(response *http.Response) string {
  body, err := ioutil.ReadAll(response.Body)

  if err != nil {
    return ""
  }

  var result map[string]interface{}
  json.Unmarshal([]byte(string(body)), &result)

  country := result["country"].(map[string]interface{})
  country_code := country["code"]

  return country_code.(string)
}

func (p GeoipNekudoProvider) IsLimitAvailable() bool {
  return true
}
