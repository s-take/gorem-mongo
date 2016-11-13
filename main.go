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
	host    = flag.String("h", "localhost", "ip of MongoDB")
	port    = flag.String("p", "27017", "port of MongoDB")
	delay   = flag.Int("d", 1, "delay get value")
	getRepl = flag.Bool("r", false, "get rs.status")
)

var keys = []string{
	"backgroundFlushing.flushes",
	"backgroundFlushing.total_ms",
	"backgroundFlushing.average_ms",
	"backgroundFlushing.last_ms",
	"backgroundFlushing.last_finished",
	"connections.current",
	"connections.available",
	"connections.totalCreated",
	"cursors.clientCursors_size",
	"cursors.totalOpen",
	"cursors.pinned",
	"cursors.totalNoTimeout",
	"cursors.timedOut",
	"dur.commits",
	"dur.journaledMB",
	"dur.writeToDataFilesMB",
	"dur.compression",
	"dur.commitsInWriteLock",
	"dur.earlyCommits",
	"dur.timeMs.dt",
	"dur.timeMs.prepLogBuffer",
	"dur.timeMs.writeToJournal",
	"dur.timeMs.writeToDataFiles",
	"extra_info.heap_usage_bytes",
	"extra_info.page_faults",
	"globalLock.totalTime",
	"globalLock.lockTime",
	"globalLock.currentQueue.total",
	"globalLock.currentQueue.readers",
	"globalLock.currentQueue.writers",
	"globalLock.activeClients.total",
	"globalLock.activeClients.readers",
	"globalLock.activeClients.writers",
	"indexCounters.accesses",
	"indexCounters.hits",
	"indexCounters.misses",
	"indexCounters.resets",
	"indexCounters.missRatio",
	"locks.admin.timeLockedMicros.r",
	"locks.admin.timeLockedMicros.w",
	"locks.admin.timeAcquiringMicros.r",
	"locks.admin.timeAcquiringMicros.w",
	"network.bytesIn",
	"network.bytesOut",
	"network.numRequests",
	"network.bytesIn",
	"opcounters.insert",
	"opcounters.query",
	"opcounters.update",
	"opcounters.delete",
	"opcounters.getmore",
	"opcounters.command",
	"mem.bits",
	"mem.resident",
	"mem.virtual",
	"mem.mapped",
	"mem.mappedWithJournal",
}

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

func writecsv(data map[string]interface{}, keys []string, printHeader bool) {
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
		record = append(record, get_value(data, expanded_key))
	}

	writer.Write(record)
	defer writer.Flush()
}

func main() {
	flag.Parse()
	con := *host + ":" + *port
	session, err := mgo.Dial(con)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	//result := bson.M{}
	var result map[string]interface{}

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
