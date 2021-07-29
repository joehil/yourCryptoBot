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
//	"os/exec"
	"fmt"
	"io"
	flag "github.com/spf13/pflag"
//	"bufio"
	"time"
	"strings"
  	"strconv"
	"crypto/hmac"
	"crypto/sha256"
/*	"syscall"
	"bytes"
	"math"
	"io/ioutil" */
	"encoding/json" 
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"github.com/google/uuid"
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

var apikey string
var apisecret string
var apiclient string

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

var pFlag string
var iFlag string
var limitFlag string
var rpcuserFlag string
var rpcpasswordFlag string
var excFlag string
var assetFlag string
var startFlag string
var endFlag string
var oFlag string
var sFlag string
var tFlag string
var aFlag string
var prFlag string
var cFlag string

func init(){
flag.StringVarP(&pFlag, "pair", "p", "ETH-EUR", "Currency pair")
flag.StringVarP(&iFlag, "interval", "i", "900", "Interval")
flag.StringVarP(&limitFlag, "limit" , "l", "1", "Limit")
flag.StringVarP(&rpcuserFlag, "rpcuser" , "u", "admin", "RPC user")
flag.StringVarP(&rpcpasswordFlag, "rpcpassword" , "w", "password", "RPC password")
flag.StringVarP(&excFlag, "exchange" , "e", "bitstamp", "Exchange name")
flag.StringVarP(&assetFlag, "asset" , "a", "SPOT", "Asset")
flag.StringVarP(&startFlag, "start" , "s", "", "Start time")
flag.StringVarP(&endFlag, "end" , "x", "", "End time")
flag.StringVarP(&oFlag, "order_id" , "o", "", "Order ID")
flag.StringVarP(&sFlag, "side" , "", "", "Order side")
flag.StringVarP(&tFlag, "type" , "", "", "Order type")
flag.StringVarP(&aFlag, "amount" , "", "", "Order amount")
flag.StringVarP(&prFlag, "price" , "", "", "Order price")
flag.StringVarP(&cFlag, "client_id" , "c", "", "Order client ID")

flag.Parse()
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
                        if v == "getaccountinfo" {
                                getAccount()
                                os.Exit(0)
                        }
                        if v == "getorders" {
                                getOrders()
                                os.Exit(0)
                        }
                        if v == "getorder" {
                                getOrder()
                                os.Exit(0)
                        }
                        if v == "submitorder" {
                                submitOrder()
                                os.Exit(0)
                        }
                        if v == "cancelorder" {
                                cancelOrder()
                                os.Exit(0)
                        }
		}

/*		argsWithoutProg := os.Args[1:]
		fmt.Println(argsWithoutProg)
		fmt.Println(gctcmd)

		out, err := exec.Command(gctcmd, argsWithoutProg...).Output()
        	if err != nil {
                	fmt.Printf("Command finished with error: %v", err)
        	}
		fmt.Println(string(out))
                os.Exit(0) */
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

        apikey = viper.GetString("apikey")
        apiclient = viper.GetString("apiclient")
        apisecret = viper.GetString("apisecret")

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
var data map[string]interface{}
var candle map[string]interface{}
var ohlc []interface{}
var layout string = "2006-01-02 15:04:05 MST"
var open string
var close string
var low string
var high string
var volume string
var itime int64

var out string

var tmst time.Time = time.Now()
var tst int64 = tmst.Unix() - 900

pFlag = strings.ToLower(strings.ReplaceAll(pFlag, "-", ""))

// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetQueryString("limit="+limitFlag+"&step="+iFlag+"&start="+fmt.Sprintf("%d",tst)).
      SetHeader("Accept", "application/json").
      Get("https://www.bitstamp.net/api/v2/ohlc/"+pFlag+"/")

if err != nil {
	fmt.Println(err)
	return
}

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors 
	fmt.Printf("JSON error: %v\n", err)
	fmt.Printf("JSON input: %v\n", resp.Body())
	return
}

candle = data["data"].(map[string]interface{})
pair := candle["pair"].(string)
ohlc = candle["ohlc"].([]interface{})

pairs := strings.Split(pair, "/")

out = "{\n"
out += " \"exchange\": \""+exchange_name+"\",\n"
out += " \"pair\": {\n"
out += "  \"delimiter\": \"-\",\n"
out += "  \"base\": \""+pairs[0]+"\",\n"
out += "  \"quote\": \""+pairs[1]+"\"\n"
out += " },\n"
out += " \"interval\": \""+iFlag+"\",\n"
out += " \"candle\": [\n"

for _, cndl := range ohlc {
	cn := cndl.(map[string]interface{})
	if cn != nil {
		open = cn["open"].(string)
                close = cn["close"].(string)
		if cn["volume"] != nil {
      	        	volume = cn["volume"].(string)
		} else {
			volume = "0"
		}
               	low = cn["low"].(string)
       	        high = cn["high"].(string)
		itime,err = strconv.ParseInt(cn["timestamp"].(string),10,64)
		t := time.Unix(itime,0)
	        if err != nil {
       			fmt.Printf("Time conversion error: %v", err)
 		}
		out += "  {\n"
		out += "   \"time\": \""+t.Format(layout)+"\",\n"
                out += "   \"low\": "+low+",\n"
                out += "   \"high\": "+high+",\n"
                out += "   \"open\": "+open+",\n"
                out += "   \"close\": "+close+",\n"
                out += "   \"volume\": "+volume+"\n"
		out += "  }\n"
		}
} 

out += " ]\n"
out += "}\n"

fmt.Print(out)

}

func getAccount() {
var data map[string]interface{}

// Create a Resty Client
client := resty.New()

for _, v := range pairs {
	pFlag = strings.ToLower(strings.ReplaceAll(v, "-", ""))
	currencies := strings.Split(v, "-") 
	timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)
	nonce := uuid.New().String()
	var toSign string = "BITSTAMP "+apikey+"POST"+"www.bitstamp.net"+"/api/v2/balance/"+pFlag+"/"+""+
                      ""+nonce+timest+"v2"
	hash := hmac.New(sha256.New, []byte(apisecret))
	io.WriteString(hash, toSign)
	signature := fmt.Sprintf("%x", hash.Sum(nil))
	resp, err := client.R().
      		SetHeader("Accept", "application/json").
      		SetHeader("X-Auth", "BITSTAMP "+apikey).
      		SetHeader("X-Auth-Signature", signature).
      		SetHeader("X-Auth-Nonce", nonce).
      		SetHeader("X-Auth-Timestamp", timest).
      		SetHeader("X-Auth-Version", "v2").
      		Post("https://www.bitstamp.net/api/v2/balance/"+pFlag+"/")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(resp.Body(), &data)
	if err != nil { // Handle JSON errors
        	fmt.Printf("JSON error: %v\n", err)
        	fmt.Printf("JSON input: %v\n", resp.Body())
        	return
	}

	c1 := data[strings.ToLower(currencies[0])+"_balance"].(string)
        c2 := data[strings.ToLower(currencies[1])+"_balance"].(string)

	fmt.Printf("\"currency\": \"%s\",\n",currencies[0])
        fmt.Printf("\"total_value\": %s,\n",c1)
        fmt.Printf("\"currency\": \"%s\",\n",currencies[1])
        fmt.Printf("\"total_value\": %s,\n",c2)
}

}

func getOrders() {
var orders []interface{}
var tm string
var id string
var price string
var amount string
var typ string
var out string
var tim time.Time

// Create a Resty Client
client := resty.New()

currencies := strings.Split(pFlag, "-")

pFlag = strings.ToLower(strings.ReplaceAll(pFlag, "-", ""))
timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)
nonce := uuid.New().String()
var toSign string = "BITSTAMP "+apikey+"POST"+"www.bitstamp.net"+"/api/v2/open_orders/"+pFlag+"/"+""+
                    ""+nonce+timest+"v2"
hash := hmac.New(sha256.New, []byte(apisecret))
io.WriteString(hash, toSign)
signature := fmt.Sprintf("%x", hash.Sum(nil))
resp, err := client.R().
	SetHeader("Accept", "application/json").
	SetHeader("X-Auth", "BITSTAMP "+apikey).
	SetHeader("X-Auth-Signature", signature).
	SetHeader("X-Auth-Nonce", nonce).
	SetHeader("X-Auth-Timestamp", timest).
	SetHeader("X-Auth-Version", "v2").
	Post("https://www.bitstamp.net/api/v2/open_orders/"+pFlag+"/")
if err != nil {
	fmt.Println(err)
	return
}

err = json.Unmarshal(resp.Body(), &orders)
if err != nil { // Handle JSON errors
       	fmt.Printf("JSON error: %v\n", err)
       	fmt.Printf("JSON input: %v\n", resp.Body())
       	return
}

out = "{\n"
out += " \"orders\": [\n"

for _, order := range orders {
        od := order.(map[string]interface{})
        if od != nil {
                id = od["id"].(string)
                tm = od["datetime"].(string)
		tim, _ = time.Parse("2006-01-02 15:04:05", tm)
                amount = od["amount"].(string)
                price = od["price"].(string)
		typ = od["type"].(string)
		out += "   {\n"
                out += "   \"exchange\": \""+exchange_name+"\",\n"
                out += "   \"id\": \""+id+"\",\n"
                out += "   \"base_currency\": \""+strings.ToUpper(currencies[0])+"\",\n"
                out += "   \"quote_currency\": \""+strings.ToUpper(currencies[1])+"\",\n"
                out += "   \"asset_type\": \"spot\",\n"
		if typ == "1" {
                	out += "   \"order_side\": \"SELL\",\n"
		} else {
                        out += "   \"order_side\": \"BUY\",\n"
		}
                out += "   \"order_type\": \"LIMIT\",\n"
                out += "   \"creation_time\": "+fmt.Sprintf("%d",tim.Unix())+",\n"
                out += "   \"update_time\": "+fmt.Sprintf("%d",tim.Unix())+",\n"
                out += "   \"status\": \"NEW\",\n"
                out += "   \"price\": "+price+",\n"
                out += "   \"amount\": "+amount+",\n"
                out += "   \"open_volume\": "+amount+"\n"
                out += "   }\n"
	}
}

out += " ]\n"
out += "}\n"
 
fmt.Println(out)
}

func getOrder() {
var order map[string]interface{}
var status string
var id float64
var out string

// Create a Resty Client
client := resty.New()

//currencies := strings.Split(pFlag, "-")

pFlag = strings.ToLower(strings.ReplaceAll(pFlag, "-", ""))
timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)
nonce := uuid.New().String()
var query = `id=`+oFlag
var toSign string = "BITSTAMP "+apikey+"POST"+"www.bitstamp.net"+"/api/v2/order_status/"+""+
                    "application/x-www-form-urlencoded"+nonce+timest+"v2"+query
hash := hmac.New(sha256.New, []byte(apisecret))
io.WriteString(hash, toSign)
signature := fmt.Sprintf("%x", hash.Sum(nil))
resp, err := client.R().
        SetHeader("Accept", "application/json").
	SetHeader("Content-Type", "application/x-www-form-urlencoded").
	SetHeader("X-Auth", "BITSTAMP "+apikey).
	SetHeader("X-Auth-Signature", signature).
	SetHeader("X-Auth-Nonce", nonce).
	SetHeader("X-Auth-Timestamp", timest).
	SetHeader("X-Auth-Version", "v2").
        SetBody(query).
	Post("https://www.bitstamp.net/api/v2/order_status/")
if err != nil {
	fmt.Println(err)
	return
}

err = json.Unmarshal(resp.Body(), &order)
if err != nil { // Handle JSON errors
       	fmt.Printf("JSON error: %v\n", err)
       	fmt.Printf("JSON input: %v\n", resp.Body())
       	return
}

out = "{\n"

if order != nil {
        id = order["id"].(float64)
        status = order["status"].(string)
        out += "   \"exchange\": \""+exchange_name+"\",\n"
	out += "   \"id\": \""+fmt.Sprintf("%.0f",id)+"\",\n"
        out += "   \"status\": \""+strings.ToUpper(status)+"\"\n"
}

out += "}"
 
fmt.Println(out) 
}

func submitOrder(){
// Create a Resty Client
client := resty.New()

pFlag = strings.ToLower(strings.ReplaceAll(pFlag, "-", ""))
sFlag = strings.ToLower(sFlag)
timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)
nonce := uuid.New().String()

var query = `amount=`+aFlag+`&price=`+prFlag

var toSign string = "BITSTAMP "+apikey+"POST"+"www.bitstamp.net"+"/api/v2/"+sFlag+"/"+pFlag+"/"+""+
                    "application/x-www-form-urlencoded"+nonce+timest+"v2"+query
hash := hmac.New(sha256.New, []byte(apisecret))
io.WriteString(hash, toSign)
signature := fmt.Sprintf("%x", hash.Sum(nil))
resp, err := client.R().
        SetHeader("Accept", "application/json").
	SetHeader("Content-Type", "application/x-www-form-urlencoded").
	SetHeader("X-Auth", "BITSTAMP "+apikey).
	SetHeader("X-Auth-Signature", signature).
	SetHeader("X-Auth-Nonce", nonce).
	SetHeader("X-Auth-Timestamp", timest).
	SetHeader("X-Auth-Version", "v2").
        SetBody(query).
	Post("https://www.bitstamp.net/api/v2/"+sFlag+"/"+pFlag+"/")
if err != nil {
	fmt.Println(err)
	return
}

fmt.Println(resp.String())
}

func cancelOrder(){
var order map[string]interface{}
var id float64
var out string

// Create a Resty Client
client := resty.New()

//currencies := strings.Split(pFlag, "-")

pFlag = strings.ToLower(strings.ReplaceAll(pFlag, "-", ""))
timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)
nonce := uuid.New().String()
var query = `id=`+oFlag
var toSign string = "BITSTAMP "+apikey+"POST"+"www.bitstamp.net"+"/api/v2/cancel_order/"+""+
                    "application/x-www-form-urlencoded"+nonce+timest+"v2"+query
hash := hmac.New(sha256.New, []byte(apisecret))
io.WriteString(hash, toSign)
signature := fmt.Sprintf("%x", hash.Sum(nil))
resp, err := client.R().
        SetHeader("Accept", "application/json").
	SetHeader("Content-Type", "application/x-www-form-urlencoded").
	SetHeader("X-Auth", "BITSTAMP "+apikey).
	SetHeader("X-Auth-Signature", signature).
	SetHeader("X-Auth-Nonce", nonce).
	SetHeader("X-Auth-Timestamp", timest).
	SetHeader("X-Auth-Version", "v2").
        SetBody(query).
	Post("https://www.bitstamp.net/api/v2/cancel_order/")
if err != nil {
	fmt.Println(err)
	return
}

err = json.Unmarshal(resp.Body(), &order)
if err != nil { // Handle JSON errors
       	fmt.Printf("JSON error: %v\n", err)
       	fmt.Printf("JSON input: %v\n", resp.Body())
       	return
}

out = "{\n"

if order != nil {
	if order["id"] != nil {
        	id = order["id"].(float64)
        	out += "   \"id\": \""+fmt.Sprintf("%.0f",id)+"\",\n"
        	out += "   \"status\": \"success\",\n"
	} else {
                out += "   \"status\": \"failed\",\n"
	}
        out += "   \"exchange\": \""+exchange_name+"\"\n"
}

out += "}"
 
fmt.Println(out) 
}
