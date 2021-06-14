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
//	"time"
//	"strings"
//    	"path"
//    	"path/filepath"
//    	"sync/atomic"
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
			cron()
			os.Exit(0)
        	}
		fmt.Println("parameter invalid")
		os.Exit(-1)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}

func cron() {
	var start string
	var end string
	var exchange string
	var interval string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)
 
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
 
	defer db.Close()

	var cand map[string]interface{}

	for i, v := range pairs {
		log.Printf("Index: %d, Value: %v\n", i, v )
		out:=getPair(v,"2021-06-13 16:00:00","2021-06-13 16:15:00")
		fmt.Println(string(out))
		err := json.Unmarshal(out, &cand)
	        if err != nil { // Handle JSON errors 
        	        fmt.Printf("JSON error: %v", err)
        	}
		start = fmt.Sprintf("%v",cand["start"])
                end = fmt.Sprintf("%v",cand["end"])
                exchange = fmt.Sprintf("%v",cand["exchange"])
                interval = fmt.Sprintf("%v",cand["interval"])
		fmt.Printf("S: %s, E: %s, Ex: %s, I: %s\n",start,end,exchange,interval)
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
