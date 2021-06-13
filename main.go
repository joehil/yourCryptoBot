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
//	"os/exec"
	"fmt"
	"io"
//	"time"
//	"strings"
//    	"path"
//    	"path/filepath"
//    	"sync/atomic" 
	"github.com/spf13/viper"
	"github.com/natefinch/lumberjack"
)

var do_trace bool = true

var ownlog string

var dirs []string

var tarcmd string

var ftpsuser string
var ftpspassword string
var ftpshost string

var dailydir string
var weeklydir string
var monthlydir string
var tempdir string

var dailykeep int64
var weeklykeep int64
var monthlykeep int64

var do_encrypt bool = true
var encryptsuffix string
var encryptpassw string

var transferfile string = "/autowebbackup.tar.gz"
var transfersuffix string = "tar.gz"

var ownlogger io.Writer

func main() {
// Set location of config 
	viper.SetConfigName("autowebbackup") // name of config file (without extension)
	viper.AddConfigPath("/etc/")   // path to look for the config file in

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
}

func read_config() {
        err := viper.ReadInConfig() // Find and read the config file
        if err != nil { // Handle errors reading the config file
                log.Fatalf("Config file not found: %v", err)
        }

        ownlog = viper.GetString("own_log")
        if ownlog =="" { // Handle errors reading the config file
                log.Fatalf("Filename for ownlog unknown: %v", err)
        }
// Open log file
        ownlogger = &lumberjack.Logger{
                Filename:   ownlog,
                MaxSize:    5, // megabytes
                MaxBackups: 3,
                MaxAge:     28, //days
                Compress:   true, // disabled by default
        }
//        defer ownlogger.Close()
        log.SetOutput(ownlogger)

        dirs = viper.GetStringSlice("dirs")

        do_trace = viper.GetBool("do_trace")

	tarcmd = viper.GetString("tarcmd")

	ftpsuser = viper.GetString("ftpsuser")
        ftpspassword = viper.GetString("ftpspassword")
        ftpshost = viper.GetString("ftpshost")

        dailydir = viper.GetString("dailydir")
        weeklydir = viper.GetString("weeklydir")
        monthlydir = viper.GetString("monthlydir")
        tempdir = viper.GetString("tempdir")

        dailykeep = viper.GetInt64("dailykeep")
        weeklykeep = viper.GetInt64("weeklykeep")
        monthlykeep = viper.GetInt64("monthlykeep")

	do_encrypt = viper.GetBool("do_encrypt")
	encryptsuffix = viper.GetString("encryptsuffix")
	encryptpassw = viper.GetString("encryptpassw")

	if do_trace {
		log.Println("do_trace: ",do_trace)
		log.Println("own_log; ",ownlog)
		for i, v := range dirs {
			log.Printf("Index: %d, Value: %v\n", i, v )
		}
	}
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("backup        Backup the directories mentioned in the config file")
}

