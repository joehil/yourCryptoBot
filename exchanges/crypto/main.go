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
	flag "github.com/spf13/pflag"
	"time"
	"strings"
//  	"strconv"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
/*	"syscall"
	"bytes"
	"math"
	"io/ioutil" */
	"encoding/json" 
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

var pipeFile = "/tmp/yourpipe"

var do_trace bool = true

var exchange_name string

var pairs []string
var tradepairs []string

var gctcmd string
var bkpcmd string

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
var otpFlag string

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
flag.StringVarP(&otpFlag, "otp" , "", "", "OTP Value")

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
		exc := "crypto"
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
                        if v == "getticker" {
                                getTicker()
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
                        if v == "getinstruments" {
                                getInstruments()
                                os.Exit(0)
                        }
		}

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
        bkpcmd = viper.GetString("bkpcmd")

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

func getSignature(request string) string {
  mac := hmac.New(sha256.New, []byte(apisecret))
  mac.Write([]byte(request))
  macsum := mac.Sum(nil)
  dst := make([]byte, hex.EncodedLen(len(macsum)))
  hex.Encode(dst, macsum)
  return string(dst)
}

func convCurr(curr string) string {
var c string = curr
if curr == "ETHEUR" {
        c ="XETHZEUR"
}
if curr == "XRPEUR" {
        c ="XXRPZEUR"
}
return c
}

func myUsage() {

}

func getCandles() {
var data map[string]interface{}
var candles []interface{}
var layout string = "2006-01-02 15:04:05 MST"
var pair string = strings.ToUpper(pFlag)
var docomma bool = false

var out string

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", "_"))

payload := url.Values{}
payload.Add("instrument_name",pFlag)
payload.Add("timeframe","15m")
payload.Add("depth","4")
payload.Add("id",timest)
payload.Add("nonce",timest)
payload.Add("timestamp",timest)

// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetQueryString(payload.Encode()).
      SetHeader("Accept", "application/json").
      Get("https://api.crypto.com/v2/public/get-candlestick")

if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors 
	fmt.Printf("JSON error: %v\n", err)
	fmt.Printf("JSON input: %v\n", resp.Body())
	return
}

//fmt.Println(data)

result := data["result"].(map[string]interface{})
candles = result["data"].([]interface{})

//fmt.Println(candles)

pairs := strings.Split(pair, "-")

out = "{\n"
out += " \"exchange\": \""+exchange_name+"\",\n"
out += " \"pair\": {\n"
out += "  \"delimiter\": \"-\",\n"
out += "  \"base\": \""+pairs[0]+"\",\n"
out += "  \"quote\": \""+pairs[1]+"\"\n"
out += " },\n"
out += " \"interval\": \""+iFlag+"\",\n"
out += " \"candle\": [\n"

for i, cndl := range candles {
	candle := cndl.(map[string]interface{})
	if candle != nil && i < 3 {
		if docomma {
			out += ",\n"
		}
		tim := candle["t"].(float64)
                open := candle["o"].(float64)
      	        high := candle["h"].(float64)
               	low := candle["l"].(float64)
       	        close := candle["c"].(float64)
                volume := candle["v"].(float64)
		t := time.Unix(int64(tim/1000),0)
		out += "  {\n"
		out += "   \"time\": \""+t.Format(layout)+"\",\n"
                out += "   \"low\": "+fmt.Sprintf("%f",low)+",\n"
                out += "   \"high\": "+fmt.Sprintf("%f",high)+",\n"
                out += "   \"open\": "+fmt.Sprintf("%f",open)+",\n"
                out += "   \"close\": "+fmt.Sprintf("%f",close)+",\n"
                out += "   \"volume\": "+fmt.Sprintf("%f",volume)+"\n"
		out += "  }\n"
		docomma = true
		} 
} 

out += " ]\n"
out += "}\n"

fmt.Print(out)
}

func getTicker() {
var data map[string]interface{}

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", "_"))

payload := url.Values{}
payload.Add("instrument_name",pFlag)
payload.Add("id",timest)
payload.Add("nonce",timest)
payload.Add("timestamp",timest)

// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetQueryString(payload.Encode()).
      SetHeader("Accept", "application/json").
      Get("https://api.crypto.com/v2/public/get-ticker")

if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors 
	fmt.Printf("JSON error: %v\n", err)
	fmt.Printf("JSON input: %v\n", resp.Body())
	return
}

//fmt.Println(data)

result := data["result"].(map[string]interface{})
ticker := result["data"].(map[string]interface{})
//fmt.Println(ticker)

if ticker["a"] != nil {
	price := ticker["a"].(float64)
        fmt.Println("{\"last\": "+fmt.Sprintf("%f",price)+",\n\"exchange\": \""+exchange_name+"\"}")
} else {
        fmt.Println("{\"status\": \"invalid\",\n")
        fmt.Println("\"message\": \""+resp.String()+"\"}")
}

}

func getAccount() {
var data map[string]interface{}
var account map[string]interface{}

// Create a Resty Client
client := resty.New()

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

method := "private/get-account-summary"

sigpayload := method + timest + apikey + timest

signature := getSignature(sigpayload)

payload := `{
"id": `+timest+`,
"method": "`+method+`",
"api_key": "`+apikey+`",
"params": {
},
"nonce": `+timest+`,
"sig": "`+signature+`"
}`

//fmt.Println(payload)

resp, err := client.R().
        SetBody(payload).
	SetHeader("Accept", "application/json").
        SetHeader("Content-Type", "application/json").
      	Post("https://api.crypto.com/v2/"+method)
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors
        fmt.Printf("JSON error: %v\n", err)
        fmt.Printf("JSON input: %v\n", resp.Body())
        return
}

//fmt.Println(data)

result := data["result"].(map[string]interface{})
accounts := result["accounts"].([]interface{})

for _, acc := range accounts {
	account = acc.(map[string]interface{})
	asset := account["currency"].(string)
        free := account["balance"].(float64)
	if free > 0 {
		fmt.Printf("\"currency\": \"%s\",\n",asset)
       		fmt.Printf("\"total_value\": %f,\n",free)
	}
}

}

func getOrders() {
var data map[string]interface{}
var price float64
var amount float64
var out string
var pair string
var typ string
var order_side string
var tim time.Time = time.Now()

pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", "_"))

// Create a Resty Client
client := resty.New()

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

method := "private/get-open-orders"

sigpayload := method + timest + apikey + "instrument_name" + pFlag + timest

signature := getSignature(sigpayload)

payload := `{
"id": `+timest+`,
"method": "`+method+`",
"api_key": "`+apikey+`",
"params": {
"instrument_name": "` + pFlag + `"
},
"nonce": `+timest+`,
"sig": "`+signature+`"
}`

//fmt.Println(payload)

resp, err := client.R().
        SetBody(payload).
	SetHeader("Accept", "application/json").
        SetHeader("Content-Type", "application/json").
      	Post("https://api.crypto.com/v2/"+method)
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors
        fmt.Printf("JSON error: %v\n", err)
        fmt.Printf("JSON input: %v\n", resp.Body())
        return
}

//fmt.Println(data)

result := data["result"].(map[string]interface{})
orders := result["order_list"].([]interface{})

out = "{\n"
out += " \"orders\": [\n"

for _, order := range orders {
	var ord map[string]interface{}
	ord = order.(map[string]interface{})

	pair = ord["instrument_name"].(string)

	pairs := strings.Split(pair, "_")

	id := ord["order_id"].(string)
	price = ord["price"].(float64)
	amount = ord["quantity"].(float64)
	order_side = strings.ToUpper(ord["side"].(string))
        typ = strings.ToUpper(ord["type"].(string))
	if pair == pFlag {
        	out += "   {\n"
        	out += "   \"exchange\": \""+exchange_name+"\",\n"
        	out += "   \"id\": \""+id+"\",\n"
        	out += "   \"base_currency\": \""+pairs[0]+"\",\n"
        	out += "   \"quote_currency\": \""+pairs[1]+"\",\n"
        	out += "   \"asset_type\": \"SPOT\",\n"
        	out += "   \"order_side\": \""+order_side+"\",\n"
        	out += "   \"order_type\": \""+typ+"\",\n"
        	out += "   \"creation_time\": "+fmt.Sprintf("%d",tim.Unix())+",\n"
        	out += "   \"update_time\": "+fmt.Sprintf("%d",tim.Unix())+",\n"
        	out += "   \"status\": \"NEW\",\n"
        	out += "   \"price\": "+fmt.Sprintf("%f",price)+",\n"
        	out += "   \"amount\": "+fmt.Sprintf("%f",amount)+",\n"
        	out += "   \"open_volume\": "+fmt.Sprintf("%f",amount)+"\n"
        	out += "   }\n"
	}
}

out += " ]\n"
out += "}\n"
 
fmt.Println(out)
}

func getOrder() {
var data map[string]interface{}
var order map[string]interface{}
var status string = "invalid"
var out string

pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", "_"))

// Create a Resty Client
client := resty.New()

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

method := "private/get-order-detail"

sigpayload := method + timest + apikey + "order_id" + oFlag + timest

signature := getSignature(sigpayload)

payload := `{
"id": `+timest+`,
"method": "`+method+`",
"api_key": "`+apikey+`",
"params": {
"order_id": "` + oFlag + `"
},
"nonce": `+timest+`,
"sig": "`+signature+`"
}`

//fmt.Println(payload)

resp, err := client.R().
        SetBody(payload).
	SetHeader("Accept", "application/json").
        SetHeader("Content-Type", "application/json").
      	Post("https://api.crypto.com/v2/"+method)
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors
        fmt.Printf("JSON error: %v\n", err)
        fmt.Printf("JSON input: %v\n", resp.Body())
        return
}

//fmt.Println(data)

result := data["result"].(map[string]interface{})

if result["order_info"] != nil {
	order = result["order_info"].(map[string]interface{})

	if order["status"] != nil {
		status = order["status"].(string)
		if status == "ACTIVE" {
			status = "OPEN"
		}
	}
}

out = "{\n"

if status != "invalid" {
        out += "   \"exchange\": \""+exchange_name+"\",\n"
	out += "   \"id\": \""+oFlag+"\",\n"
        out += "   \"status\": \""+strings.ToUpper(status)+"\"\n"
} else {
        out += "   \"exchange\": \""+exchange_name+"\",\n"
        out += "   \"status\": \""+strings.ToUpper(status)+"\"\n"
}

out += "}"
 
fmt.Println(out) 
}

func submitOrder(){
var order map[string]interface{}
pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", ""))

sFlag = strings.ToLower(sFlag)

// Create a Resty Client
client := resty.New()

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

payload := url.Values{}
payload.Add("symbol",pFlag)
payload.Add("side",sFlag)
payload.Add("type",tFlag)
payload.Add("price",prFlag)
payload.Add("quantity",aFlag)
payload.Add("timeInForce","GTC")
payload.Add("recvWindow","5000")
payload.Add("timestamp",timest)

//fmt.Println(payload.Encode())

signature := getSignature(payload.Encode())

payload.Add("signature",signature)

resp, err := client.R().
        SetBody(payload.Encode()).
	SetHeader("Accept", "application/json").
	SetHeader("X-MBX-APIKEY", apikey).
	SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
	Post("https://api.binance.com/api/v3/order")
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &order)
if err != nil { // Handle JSON errors
        fmt.Printf("JSON error: %v\n", err)
        fmt.Printf("JSON input: %v\n", resp.Body())
        return
}

if order["orderId"] != nil {
	id := order["orderId"].(float64)
        fmt.Println("{\"id\": \""+fmt.Sprintf("%.0f",id)+"\"}")
} else { 
	fmt.Println("{\"status\": \"invalid\",\n")
        fmt.Println("\"message\": \""+resp.String()+"\"}")
}

}

func cancelOrder(){
var out string
var order map[string]interface{}

pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", ""))

// Create a Resty Client
client := resty.New()

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

payload := url.Values{}
payload.Add("symbol",pFlag)
payload.Add("orderId",oFlag)
payload.Add("recvWindow","5000")
payload.Add("timestamp",timest)

//fmt.Println(payload.Encode())

signature := getSignature(payload.Encode())

payload.Add("signature",signature)

resp, err := client.R().
        SetQueryString(payload.Encode()).
	SetHeader("Accept", "application/json").
	SetHeader("X-MBX-APIKEY", apikey).
	Delete("https://api.binance.com/api/v3/order")
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &order)
if err != nil { // Handle JSON errors
        fmt.Printf("JSON error: %v\n", err)
        fmt.Printf("JSON input: %v\n", resp.Body())
        return
}

out = "{\n"

if order["orderId"] != nil {
        id := order["orderId"].(float64)
        out += "   \"id\": \""+fmt.Sprintf("%.0f",id)+"\",\n"
        out += "   \"status\": \"success\",\n"
} else {
        out += "   \"status\": \"failed\",\n"
}


out += "   \"exchange\": \""+exchange_name+"\"\n"

out += "}"
 
fmt.Println(out) 
}

func getInstruments() {
var data map[string]interface{}

pFlag = strings.ToUpper(strings.ReplaceAll(pFlag,"-","_"))

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

payload := url.Values{}
payload.Add("id",timest)
payload.Add("nonce",timest)
payload.Add("timestamp",timest)

// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetQueryString(payload.Encode()).
      SetHeader("Accept", "application/json").
      Get("https://api.crypto.com/v2/public/get-instruments")

if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &data)
if err != nil { // Handle JSON errors 
	fmt.Printf("JSON error: %v\n", err)
	fmt.Printf("JSON input: %v\n", resp.Body())
	return
}

//fmt.Println(data)

result := data["result"].(map[string]interface{})
instruments := result["instruments"].([]interface{})

//fmt.Println(instruments)

for _, instrument := range instruments {
	instr := instrument.(map[string]interface{})
	instrname := instr["instrument_name"].(string)
	if instrname == pFlag || pFlag == "ALL" {
		fmt.Println(instrument)
	}
} 

}
