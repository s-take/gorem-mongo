package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"math"
	"os"
	"strings"
	"time"
)

var (
	host    = flag.String("host", "localhost", "connect host")
	delay   = flag.Int("d", 1, "delay get value")
	getRepl = flag.Bool("r", false, "get rs.status")
)

var keys = []string{"pid", "uptime"}

func get_value(data map[string]interface{}, keyparts []string) string {
	if len(keyparts) > 1 {
		subdata, _ := data[keyparts[0]].(map[string]interface{})
		return get_value(subdata, keyparts[1:])
	} else if v, ok := data[keyparts[0]]; ok {
		switch v.(type) {
		case nil:
			return ""
		case float64:
			f, _ := v.(float64)
			if math.Mod(f, 1.0) == 0.0 {
				return fmt.Sprintf("%d", int(f))
			} else {
				return fmt.Sprintf("%f", f)
			}
		default:
			return fmt.Sprintf("%+v", v)
		}
	}

	return ""
}

func writecsv(result map[string]interface{}, keys []string, printHeader bool) {
	//file, _ := os.Create("result.csv")
	//defer file.Close()
	//writer := csv.NewWriter(file)
	writer := csv.NewWriter(os.Stdout)
	if printHeader {
		writer.Write(keys)
	}

	var expanded_keys [][]string
	for _, key := range keys {
		expanded_keys = append(expanded_keys, strings.Split(key, "."))
	}

	var record []string
	for _, expanded_key := range expanded_keys {
		record = append(record, get_value(result, expanded_key))
	}
	writer.Write(record)
	defer writer.Flush()
}

func main() {
	session, err := mgo.Dial(*host)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	result := bson.M{}

	for cnt := 0; ; {
		cnt++
		if err := session.DB("admin").Run(bson.D{{"serverStatus", 1}}, &result); err != nil {
			panic(err)
		} else {
			if cnt == 1 {
				writecsv(result, keys, true)
			} else {
				writecsv(result, keys, false)
			}
			time.Sleep(time.Duration(*delay) * time.Second)
		}
	}
	/*
		if err := session.DB("test").Run("dbstats", &result); err != nil {
			panic(err)
		} else {
			fmt.Println(result)
		}
		if err := session.DB("admin").Run("replSetGetStatus", &result); err != nil {
			panic(err)
		} else {
			fmt.Println(result)
		}
	*/
}
