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
	"crypto/sha512"
	"encoding/base64"
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
		exc := "coinbase"
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
/*                        if v == "gettransactions" {
                                getTransactions()
                                os.Exit(0)
                        } */
		}

		fmt.Println("{}")
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

func getKrakenSignature(url_path string, values url.Values, secret []byte) string {

  sha := sha256.New()
  sha.Write([]byte(values.Get("nonce") + values.Encode()))
  shasum := sha.Sum(nil)

  mac := hmac.New(sha512.New, secret)
  mac.Write(append([]byte(url_path), shasum...))
  macsum := mac.Sum(nil)
  return base64.StdEncoding.EncodeToString(macsum)
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
var data []interface{}
var layout string = "2006-01-02T15:04:05.000000Z"
var laytarget string = "2006-01-02 15:04:05 MST"
var open float64
var close float64
var low float64
var high float64
var volume float64
var docomma bool = false

_,offset := time.Now().Zone()

var tmnow int64 = (time.Now().UTC().Unix() / 900) * 900 - int64(offset)
var tmthen int64 = tmnow - 1800

var out string

tmnowtime := time.Unix(tmnow,0)
tmthentime := time.Unix(tmthen,0)
tmnowstr := tmnowtime.Format(layout)
tmthenstr := tmthentime.Format(layout)

//pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", "_"))

currencies := strings.Split(pFlag,"-")

// Create a Resty Client
client := resty.New()

//fmt.Println(tmthenstr)
//fmt.Println(tmnowstr)

resp, err := client.R().
      SetQueryString("granularity=900&start="+tmthenstr+"&end="+tmnowstr).
      SetHeader("Accept", "application/json").
      Get("https://api.pro.coinbase.com/products/"+pFlag+"/candles")

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

out = "{\n"
out += " \"exchange\": \""+exchange_name+"\",\n"
out += " \"pair\": {\n"
out += "  \"delimiter\": \"-\",\n"
out += "  \"base\": \""+currencies[0]+"\",\n"
out += "  \"quote\": \""+currencies[1]+"\"\n"
out += " },\n"
out += " \"interval\": \""+iFlag+"\",\n"
out += " \"candle\": [\n"

for _, cn := range data {
	cndl := cn.(interface {})
//	fmt.Println(cndl)
	if cndl != nil {
		if docomma {
			out += ",\n"
		}
		cn := cndl.([]interface {})
                open = cn[3].(float64)
      	        high = cn[2].(float64)
               	low = cn[1].(float64)
       	        close = cn[4].(float64)
                volume = cn[5].(float64)
		tms := cn[0].(float64)
		t := time.Unix(int64(tms),0)
		out += "  {\n"
		out += "   \"time\": \""+t.Format(laytarget)+"\",\n"
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
var ticker map[string]interface{}
pFlag = strings.ToUpper(strings.ReplaceAll(pFlag, "-", "_"))

// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetHeader("Accept", "application/json").
      Get("https://api.exchange.bitpanda.com/public/v1/market-ticker/"+pFlag)

if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &ticker)
if err != nil { // Handle JSON errors
       	fmt.Printf("JSON error: %v\n", err)
       	fmt.Printf("JSON input: %v\n", resp.Body())
       	return
}

if ticker["last_price"] != nil {
	price := ticker["last_price"].(string)
        fmt.Println("{\"last\": "+price+",\n\"exchange\": \""+exchange_name+"\"}")
} else {
	fmt.Println("{\"status\": \"invalid\",\n")
        fmt.Println("\"message\": \""+resp.String()+"\"}")
}

}

func getAccount() {
var data map[string]interface{}

// Create a Resty Client
client := resty.New()

resp, err := client.R().
	SetHeader("Accept", "application/json").
	SetHeader("Authorization", "Bearer "+apikey).
	Get("https://api.exchange.bitpanda.com/public/v1/account/balances")
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

balances := data["balances"].([]interface{})

//fmt.Println(balances)

for _,balance := range balances {
	bala := balance.(map[string]interface{})
	curr := bala["currency_code"].(string)
	amount := bala["available"].(string)
	fmt.Printf("\"currency\": \"%s\",\n",curr)
        fmt.Printf("\"total_value\": %s,\n",amount)
}

}

func getOrders() {
var orders map[string]interface{}
var out string

client := resty.New()

pFlag = strings.ReplaceAll(pFlag,"-","_")

payload := url.Values{}
payload.Add("instrument_code",pFlag)
payload.Add("with_just_orders","true")

resp, err := client.R().
        SetBody(payload.Encode()).
	SetHeader("Accept", "application/json").
	SetHeader("Authorization", "Bearer "+apikey).
	Get("https://api.exchange.bitpanda.com/public/v1/account/orders")
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

err = json.Unmarshal(resp.Body(), &orders)
if err != nil { // Handle JSON errors
       	fmt.Printf("JSON error: %v\n", err)
       	fmt.Printf("JSON input: %v\n", resp.Body())
       	return
}

orderhist := orders["order_history"].([]interface{})

//fmt.Println(orderhist)

out = "{\n"
out += " \"orders\": [\n"

for _, order := range orderhist {
	ord := order.(map[string]interface{})
	or := ord["order"].(map[string]interface{})
	amount := or["amount"].(string)
        price := or["price"].(string)
        typ := or["type"].(string)
        side := or["side"].(string)
	id := or["order_id"].(string)
        tim := or["time"].(string)
	tmtime,_ := time.Parse("2006-01-02T15:04:05.999999Z",tim)
	pair := or["instrument_code"].(string)
        base := strings.ReplaceAll(pair, "_EUR", "")

	if pair == pFlag { 
        	out += "   {\n"
        	out += "   \"exchange\": \""+exchange_name+"\",\n"
        	out += "   \"id\": \""+id+"\",\n"
        	out += "   \"base_currency\": \""+base+"\",\n"
        	out += "   \"quote_currency\": \"EUR\",\n"
        	out += "   \"asset_type\": \"SPOT\",\n"
        	out += "   \"order_side\": \""+side+"\",\n"
        	out += "   \"order_type\": \""+typ+"\",\n"
        	out += "   \"creation_time\": "+fmt.Sprintf("%d",tmtime.Unix())+",\n"
        	out += "   \"update_time\": "+fmt.Sprintf("%d",tmtime.Unix())+",\n"
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
var status string = "invalid"
var out string

client := resty.New()

resp, err := client.R().
	SetHeader("Accept", "application/json").
	SetHeader("Authorization", "Bearer "+apikey).
	Get("https://api.exchange.bitpanda.com/public/v1/account/orders/"+oFlag)
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

//fmt.Println(order)

ord := order["order"].(map[string]interface{})

status = ord["status"].(string)
filled_amount := ord["filled_amount"].(string)

if filled_amount == "0.0" {
	status = "CANCELED"
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
// Create a Resty Client
client := resty.New()

pFlag = strings.ReplaceAll(pFlag,"-","_")

payload := `{
"instrument_code": "`+pFlag+`",
"side": "`+sFlag+`",
"type": "`+tFlag+`",
"amount": "`+aFlag+`",
"price": "`+prFlag+`"
}`

//fmt.Println(payload)

resp, err := client.R().
        SetBody(payload).
	SetHeader("Accept", "application/json").
	SetHeader("Authorization", "Bearer "+apikey).
	SetHeader("Content-Type", "application/json").
	Post("https://api.exchange.bitpanda.com/public/v1/account/orders")
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

//fmt.Println(order)

if order["order_id"] != nil {
	id := order["order_id"].(string)
	fmt.Println("{\"id\": \""+id+"\"}")
} else {
	fmt.Println("{\"status\": \"invalid\",")
	fmt.Println("\"message\": \""+resp.String()+"\"}")
}

}

func cancelOrder(){
var out string

// Create a Resty Client
client := resty.New()

_, err := client.R().
	SetHeader("Accept", "application/json").
	SetHeader("Authorization", "Bearer "+apikey).
	Delete("https://api.exchange.bitpanda.com/public/v1/account/orders/"+oFlag)
if err != nil {
	fmt.Println(err)
	return
}

//fmt.Println(resp.String())

out = "{\n"
out += "   \"id\": \""+oFlag+"\",\n"
out += "   \"status\": \"success\",\n"
out += "   \"exchange\": \""+exchange_name+"\"\n"
out += "}"
 
fmt.Println(out) 
}

func getTransactions() {
var transactions map[string]interface{}
var out string
var docomma bool = false

// Create a Resty Client
client := resty.New()

timest := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

payload := url.Values{}
payload.Add("nonce",timest)

b64DecodedSecret, _ := base64.StdEncoding.DecodeString(apisecret)

signature := getKrakenSignature("/0/private/TradesHistory", payload, b64DecodedSecret)

resp, err := client.R().
        SetBody(payload.Encode()).
	SetHeader("Accept", "application/json").
	SetHeader("API-Key", apikey).
	SetHeader("API-Sign", signature).
        SetHeader("User-Agent", "yourCryptoBot").
	SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8").
	Post("https://api.kraken.com/0/private/TradesHistory")
if err != nil {
	fmt.Println(err)
	return
}

err = json.Unmarshal(resp.Body(), &transactions)
if err != nil { // Handle JSON errors
        fmt.Printf("JSON error: %v\n", err)
        fmt.Printf("JSON input: %v\n", resp.Body())
        return
}

//fmt.Println(resp.String()) 

out = "[\n"

if transactions["result"] != nil {
	var trad map[string]interface{}
	result := transactions["result"].(map[string]interface{})
	trades := result["trades"].(map[string]interface{})
	for _, tr := range trades {
		trad = tr.(map[string]interface{})
        	if docomma {
                	out += fmt.Sprintln(",")
        	}

	        id := trad["ordertxid"].(string)
        	fee := trad["fee"].(string)
        	amount_quote := trad["cost"].(string)
        	amount := trad["vol"].(string)
        	price := trad["price"].(string)
        	tmst := trad["time"].(float64)
		tmint := int64(tmst)
		tmtim := time.Unix(tmint,int64(0))
		timest := tmtim.Format("2006-01-02 15:04:05.00000")
        	pair := trad["pair"].(string)
		pair = strings.ReplaceAll(pair,"EUR","-EUR")
        	typ := trad["type"].(string)
		if typ != "sell" {
			amount_quote = "-" + amount_quote
		}
                out += fmt.Sprintf("{\"id\": \"%s\", \"fee\": %s, \"amount\": %s, \"amount_quote\": %s, \"price\": %s, \"timestamp\": \"%s\", \"pair\": \"%s\"}\n",
                                id,fee,amount,amount_quote,price,timest,pair)
                docomma = true
	}
}

out += "]"

fmt.Println(out)

}
