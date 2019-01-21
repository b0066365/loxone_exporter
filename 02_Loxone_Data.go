package main

import (
  "fmt"
  "log"
  "net/http"
  "io/ioutil"
  "strings"
  "encoding/xml"
  "time"
  "strconv"

  "github.com/influxdata/influxdb/client/v2"
)
var LOXONE_IP = "10.0.0.77"
var INFLUXDB_IP = "10.0.0.51"
var INFLUXDB_DB = "loxone"

type Airdevice struct {
    Name    string   `xml:"Name,attr"`
  	Type    string   `xml:"Type,attr"`
  	Place    string   `xml:"Place,attr"`
    Hops    int   `xml:"Hops,attr"`
    RoundTripTime   int   `xml:"RoundTripTime,attr"`
    Battery int   `xml:"Battery,attr"`
}

type Miniserver struct {
	XMLName xml.Name `xml:"Status"`
  AirDevice []Airdevice   `xml:"Miniserver>Extension>AirDevice"`
}

type Loxone_SSA struct {
	XMLName xml.Name `xml:"LL"`
  Value   string   `xml:"value,attr"`
}


func init(){
  log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

func LOXONE_GET(url string, user string, pwd string) (string){
  client := &http.Client{}
  req, err := http.NewRequest("GET", url, nil)
  req.SetBasicAuth(user, pwd)
  //log.Println("HTTP GET Request to URL:", url)
  resp, err := client.Do(req)
  if err != nil{
      log.Fatal(err)
  }
  bodyText, err := ioutil.ReadAll(resp.Body)
  s := string(bodyText)
  return s
}

func INFLUXDB_WRITE(Measure string, AirDevice string, Value string) {
  // Make client
  c, err := client.NewHTTPClient(client.HTTPConfig{
    Addr: "http://"+INFLUXDB_IP+":8086",
  })
  if err != nil {
    fmt.Println("Error creating InfluxDB Client: ", err.Error())
  }
  defer c.Close()

  // Create a new point batch
  bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
    Database:  INFLUXDB_DB,
    Precision: "s",
  })

  // Create a point and add to batch
  tags := map[string]string{"device": AirDevice}
  new_value, err := strconv.ParseFloat(Value, 10)
  if err != nil {
    fmt.Println("Error: ", err.Error())
  }
  fields := map[string]interface{}{ "value": new_value }
  fmt.Println(fields)
  pt, err := client.NewPoint(Measure, tags, fields, time.Now())
  if err != nil {
    fmt.Println("Error: ", err.Error())
  }
  bp.AddPoint(pt)
  //fmt.Println(bp)

  // Write the batch
  c.Write(bp)
}

func LOXONE_Power(AirDevice string) (string){
  resp := LOXONE_GET("http://"+LOXONE_IP+"/dev/sps/io/"+AirDevice+" Leistung/state", "dirk", "$cisco12")
  var ssa Loxone_SSA
  err := xml.Unmarshal([]byte(resp), &ssa)
  if err != nil{
      log.Fatal(err)
  }
  return strings.TrimSuffix(ssa.Value, "kW")
  //fmt.Println("Power:", strings.TrimSuffix(ssa.Value, "kW"))

}

func LOXONE_Temperature(AirDevice string) (string){
  resp := LOXONE_GET("http://"+LOXONE_IP+"/dev/sps/io/"+AirDevice+" Temperatur/state", "dirk", "$cisco12")
  var ssa Loxone_SSA
  err := xml.Unmarshal([]byte(resp), &ssa)
  if err != nil{
      log.Fatal(err)
  }
  return strings.TrimSuffix(ssa.Value, "°")
  //fmt.Println("Power:", strings.TrimSuffix(ssa.Value, "kW"))

}

func LOXONE_Humidity(AirDevice string) (string){
  resp := LOXONE_GET("http://"+LOXONE_IP+"/dev/sps/io/"+AirDevice+" Luftfeuchte/state", "dirk", "$cisco12")
  var ssa Loxone_SSA
  err := xml.Unmarshal([]byte(resp), &ssa)
  if err != nil{
      log.Fatal(err)
  }
  return strings.TrimSuffix(ssa.Value, "%")
  //fmt.Println("Power:", strings.TrimSuffix(ssa.Value, "kW"))

}

//------------------------------------------------------------------------------

func main() {
    log.Println("LOXONE_Exporter started...")

    resp := LOXONE_GET("http://"+LOXONE_IP+"/data/status", "dirk", "$cisco12")
    log.Println(strings.Count(resp, "<Miniserver"), "Miniserver found")
    log.Println(strings.Count(resp, "<AirDevice"), "AirDevices found")

    var miniserver Miniserver
    err := xml.Unmarshal([]byte(resp), &miniserver)
    if err != nil{
        log.Fatal(err)
    }

    for i := 0; i < len(miniserver.AirDevice); i++ {
        fmt.Println("Device Name:", miniserver.AirDevice[i].Name)
        fmt.Println("Device Type:", miniserver.AirDevice[i].Type)
        if miniserver.AirDevice[i].Type == "Smart Socket Air" {
          fmt.Println("Power:", LOXONE_Power(miniserver.AirDevice[i].Name))
          INFLUXDB_WRITE("Test_Power", miniserver.AirDevice[i].Name, LOXONE_Power(miniserver.AirDevice[i].Name))
          fmt.Println("Temperatur:", LOXONE_Temperature(miniserver.AirDevice[i].Name))
          INFLUXDB_WRITE("Test_Temperatur", miniserver.AirDevice[i].Name, LOXONE_Temperature(miniserver.AirDevice[i].Name))

        }

        if miniserver.AirDevice[i].Type == "Temperatur- und Feuchtefühler Air" {
          fmt.Println("Temperatur:", LOXONE_Temperature(miniserver.AirDevice[i].Name))
          fmt.Println("Humidity:", LOXONE_Humidity(miniserver.AirDevice[i].Name))
        }


        fmt.Println("Device Location:", miniserver.AirDevice[i].Place)
        fmt.Println("Device Battery:", miniserver.AirDevice[i].Battery)
        fmt.Println("Device RTT:", miniserver.AirDevice[i].RoundTripTime)
        fmt.Println("#-----------")
    }
}
