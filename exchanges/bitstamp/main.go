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
	"os"
	"os/exec"
	"fmt"
	"flag"
//	"bufio"
	"time"
	"strings"
//  	"strconv"
/*	"syscall"
	"bytes"
	"math"
	"io/ioutil"
	"encoding/json" */ 
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

var pipeFile = "/tmp/yourpipe"

var do_trace bool = true

var exchange_name string

var pairs []string
var tradepairs []string

var gctcmd string

var gctuser string
var gctpassword string

var pguser string
var pgpassword string
var pgdb string

var tbtoken string
var limit_depth int
var invest_amount int
var minwin int

var amountcomma map[string]string
var pricecomma map[string]string

var key_issuer string
var key_account string
var key_secret string

type Parm struct {
	key string
	intp int64
	floatp float64
	stringp string
	datep time.Time
	timep time.Time
	timestampp time.Time
}

type Order struct {
        exchange string
        id string
        base_currency string
        quote_currency string
        asset string
        order_side string
        order_type string
        creation_time float64
        update_time float64
        status string
        price float64
        amount float64
        open_volume float64
	cost float64
}

func main() {
// Set location of config 
	dirname, err := os.UserHomeDir()
    	if err != nil {
        	fmt.Println( err )
    	}

//	viper.SetConfigName("yourCryptoBot") // name of config file (without extension)
	viper.AddConfigPath(dirname+"/.yourCryptoBot/")   // path to look for the config file in

// Get commandline args
	if len(os.Args) > 1 {
		exc := "bitstamp"
		viper.SetConfigName(exc) // name of config file (name of exchange)
// Read config
		read_config()

		if (exc != exchange_name) {
			panic("Wrong exchange")
		}

		for _, v := range os.Args {
		    	if v == "gethistoriccandlesextended" {
				getCandles()
				os.Exit(0)
    			}
		}

		argsWithoutProg := os.Args[1:]
		fmt.Println(argsWithoutProg)
		fmt.Println(gctcmd)

		out, err := exec.Command(gctcmd, argsWithoutProg...).Output()
        	if err != nil {
                	fmt.Printf("Command finished with error: %v", err)
        	}
		fmt.Println(string(out))
                os.Exit(0)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}

func read_config() {
        err := viper.ReadInConfig() // Find and read the config file
        if err != nil { // Handle errors reading the config file
                fmt.Printf("Config file not found: %v", err)
        }

        pairs = viper.GetStringSlice("pairs")
        tradepairs = viper.GetStringSlice("tradepairs")

        do_trace = viper.GetBool("do_trace")

        exchange_name = viper.GetString("exchange_name")

	gctcmd = viper.GetString("gctcmd")

	gctuser = viper.GetString("gctuser")
        gctpassword = viper.GetString("gctpassword")

        pguser = viper.GetString("pguser")
        pgpassword = viper.GetString("pgpassword")
	pgdb = viper.GetString("pgdb")

	tbtoken = viper.GetString("tbtoken")

        key_issuer = viper.GetString("key_issuer")
        key_account = viper.GetString("key_account")
        key_secret = viper.GetString("key_secret")

	amountcomma = viper.GetStringMapString("amountcomma")
        pricecomma = viper.GetStringMapString("pricecomma")

	limit_depth = viper.GetInt("limit_depth")
        invest_amount = viper.GetInt("invest_amount")
        minwin = viper.GetInt("minwin")

        if invest_amount > 500 || limit_depth < 50 {
                invest_amount = 100
        }

	if do_trace {
		fmt.Println("do_trace: ",do_trace)
                fmt.Printf("exchange_name: %s\n",exchange_name)
                fmt.Printf("limit_depth: %d\n",limit_depth)
                fmt.Printf("invest_amount: %d\n",invest_amount)
                fmt.Printf("minwin: %d\n",minwin)
		fmt.Println(amountcomma)
                fmt.Println(pricecomma)
		for i, v := range pairs {
			fmt.Printf("Index: %d, Value: %v\n", i, v )
		}
	}
}

func myUsage() {

}

func getCandles() {
var pFlag string
var iFlag string
var limitFlag string

flag.StringVar(&pFlag, "p" , "ETH-EUR", "Currency pair")
flag.StringVar(&iFlag, "i" , "900", "Interval")
flag.StringVar(&limitFlag, "limit" , "1", "Limit")

flag.Parse()

pFlag = strings.ToLower(strings.ReplaceAll(pFlag, "-", ""))

fmt.Printf("i: %s, p: %s\n",iFlag,pFlag)

// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetQueryString("limit="+limitFlag+"&step="+iFlag).
      SetHeader("Accept", "application/json").
      Get("https://www.bitstamp.net/api/v2/ohlc/"+pFlag+"/")

if err != nil {
	fmt.Println(err)
} else {
	fmt.Println(string(resp.Body()))
}

}
