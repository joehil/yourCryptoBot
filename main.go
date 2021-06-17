/*MIT License

Copyright (c) 2021 Joerg Hillebrand

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE. */

package main

import (
	"log"
	"os"
	"os/exec"
	"fmt"
	"io"
	"time"
//	"strings"
	"encoding/json" 
	"github.com/spf13/viper"
	"github.com/natefinch/lumberjack"
    	"database/sql"
    	_ "github.com/lib/pq"
)

var do_trace bool = true

var ownlog string

var pairs []string

var gctcmd string

var gctuser string
var gctpassword string

var pguser string
var pgpassword string
var pgdb string

var ownlogger io.Writer

func main() {
// Set location of config 
	dirname, err := os.UserHomeDir()
    	if err != nil {
        	log.Fatal( err )
    	}

	viper.SetConfigName("yourCryptoBot") // name of config file (without extension)
	viper.AddConfigPath(dirname+"/.yourCryptoBot/")   // path to look for the config file in

// Read config
	read_config()

// Get commandline args
	if len(os.Args) > 1 {
        	a1 := os.Args[1]
        	if a1 == "cron" {
			getCandles()
			deleteStats()
			insertStats()
			updateStats()
			os.Exit(0)
        	}
		fmt.Println("parameter invalid")
		os.Exit(-1)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}

func getCandles() {
	var start string
	var end string
	var exchange string
	var interval string
	var base string
	var quote string
	var open float64
	var close float64
	var low float64
	var high float64
	var volume float64
	var stime string
	var layout string = "2006-01-02 15:04:05 MST"

	var cand map[string]interface{}
	var pair map[string]interface{}
        var cndls []interface{}

	ti1 := time.Now()
	ti2 := time.Now()

	ti1 = ti1.Add(time.Minute * -15)
	fmt.Println(ti2.String())
	limit1 := ti1.String()[0:16]
	limit2 := ti2.String()[0:16]

	for i, v := range pairs {
		log.Printf("Index: %d, Value: %v\n", i, v )
		out:=getPair(v,limit1+":00",limit2+":00")
		if out == nil {
			continue
		}
		err := json.Unmarshal(out, &cand)
	        if err != nil { // Handle JSON errors 
        	        fmt.Printf("JSON error: %v\n", err)
			fmt.Printf("JSON input: %v\n",string(out))
			continue
        	}
		start = cand["start"].(string)
                end = cand["end"].(string)
                exchange = cand["exchange"].(string)
                interval = cand["interval"].(string)
		pair = cand["pair"].(map[string]interface{})
		base = pair["base"].(string)
		quote = pair["quote"].(string)
                cndls = cand["candle"].([]interface{})
		fmt.Printf("S: %s, E: %s, Ex: %s, I: %s, C:%s-%s\n",start,end,exchange,interval,base,quote)
		for _, cndl := range cndls {
			cn := cndl.(map[string]interface{})
			if cn != nil {
				open = cn["open"].(float64)
        	                close = cn["close"].(float64)
                	        volume = cn["volume"].(float64)
                        	low = cn["low"].(float64)
	           	        high = cn["high"].(float64)
				stime = cn["time"].(string)
				t, err := time.Parse(layout, stime)
			        if err != nil {
                			fmt.Printf("Time conversion error: %v", err)
       		 		}
				fmt.Printf("O: %f, C: %f, H: %f, L: %f, V: %f, T: %s\n",open,close,high,low,volume,t)
				insertCandles(exchange,base+"-"+quote,interval,t,open,high,low,close,volume,"SPOT")
			}
    		}

	}
}

func insertCandles(exchange string, pair string, interval string, timest time.Time, open float64, high float64, low float64,
                   close float64, volume float64, asset string) {
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

	sqlStatement := `
	INSERT INTO yourcandle (exchange_name, pair, interval, timestamp, open, high, low, close, volume, asset)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = db.Exec(sqlStatement, exchange, pair, interval, timest, open, high, low, close, volume, asset)
	if err != nil {
  		fmt.Printf("SQL error: %v\n",err)
	}
}

func insertStats() {
	fmt.Println("Insert statistics")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
	insert into yourlimits (pair, min,avg,max,count,potwin)
	(select pair, min(close) as min, avg(close) as avg, max(close) as max, count(close) as count,
	(max(close) - min(close)) * 100 / min(close) as potwin
	from yourcandle 
	where "timestamp" > current_timestamp - interval '7 days'
	group by pair
	order by pair);`
        _, err = db.Exec(sqlStatement)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func updateStats() {
        fmt.Println("Update statistics")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
	WITH subquery AS (
	select close,pair from yourcandle y 
	where timestamp =
	(select max(timestamp) from yourcandle)
	)
	UPDATE yourlimits l
	SET current = subquery.close
	FROM subquery
	WHERE l.pair = subquery.pair;`
        _, err = db.Exec(sqlStatement)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}


func deleteStats() {
	fmt.Println("Delete statistics")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        delete from yourlimits;`
        _, err = db.Exec(sqlStatement)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func getPair(p string, s string, e string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "gethistoriccandlesextended",
        "-e","binance","-a","SPOT","-p",p,"-i","900",
        "--start",s,"--end",e).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
	return out
}

func read_config() {
        err := viper.ReadInConfig() // Find and read the config file
        if err != nil { // Handle errors reading the config file
                log.Printf("Config file not found: %v", err)
        }

        ownlog = viper.GetString("own_log")
        if ownlog =="" { // Handle errors reading the config file
                fmt.Printf("Filename for ownlog unknown: %v", err)
        }
// Open log file
        ownlogger = &lumberjack.Logger{
                Filename:   ownlog,
                MaxSize:    5, // megabytes
                MaxBackups: 3,
                MaxAge:     28, //days
                Compress:   true, // disabled by default
        }
//        defer log.Close()
        log.SetOutput(ownlogger)

        pairs = viper.GetStringSlice("pairs")

        do_trace = viper.GetBool("do_trace")

	gctcmd = viper.GetString("gctcmd")

	gctuser = viper.GetString("gctuser")
        gctpassword = viper.GetString("gctpassword")

        pguser = viper.GetString("pguser")
        pgpassword = viper.GetString("pgpassword")
	pgdb = viper.GetString("pgdb")

	if do_trace {
		fmt.Println("do_trace: ",do_trace)
		fmt.Println("own_log; ",ownlog)
		for i, v := range pairs {
			fmt.Printf("Index: %d, Value: %v\n", i, v )
		}
	}
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("cron        Do regular work")
     fmt.Println("serv        Do regular work, but without gctcli")
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}
