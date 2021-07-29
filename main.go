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
	"bufio"
	"time"
	"strings"
	"strconv"
	"syscall"
	"bytes"
	"math"
	"image/png"
	"io/ioutil"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"encoding/json" 
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/spf13/viper"
	"github.com/sajari/regression"
    	"database/sql"
    	_ "github.com/lib/pq"
)

var pipeFile = "/tmp/yourpipe"

var do_trace bool = true
var sleepbeforerun int

var exchange_name string

var pairs []string
var tradepairs []string

var gctcmd string
var wwwpath string

var gctuser string
var gctpassword string

var pguser string
var pgpassword string
var pgdb string

var tbtoken string
var limit_depth int
var invest_amount int
var minwin int
var useticker bool = false

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
	fmt.Printf("%v\n", time.Now().String())

// Set location of config 
	dirname, err := os.UserHomeDir()
    	if err != nil {
        	fmt.Println( err )
    	}

//	viper.SetConfigName("yourCryptoBot") // name of config file (without extension)
	viper.AddConfigPath(dirname+"/.yourCryptoBot/")   // path to look for the config file in

// Get commandline args
	if len(os.Args) > 2 {
			exc := os.Args[1]
			viper.SetConfigName(exc) // name of config file (name of exchange)

// Read config
			read_config()

			if (exc != exchange_name) {
				panic("Wrong exchange")
			}

        	a1 := os.Args[2]
        	if a1 == "cron" {
                        time.Sleep(time.Duration(sleepbeforerun) * time.Second)
			getCandles()
			deleteCandles()
			deleteStats()
			insertStats()
			updateStats()
			processMinMax()
			calculateLimit()
			calculateTrends()
			deleteAccounts()
			readAccount()
			readOrders()
			processOrders()
			deleteOrders()
                        writeCharts()
			buyOrders()
			sellOrders()
                        deactivatePositions()
			activatePositions()
			os.Exit(0)
        	}
                if a1 == "sell" {
                        sellOrders()
                        os.Exit(0)
                }
                if a1 == "jhtest" {
                        jhtest()
                        os.Exit(0)
                }
                if a1 == "butsell" {
                        time.Sleep(15 * time.Second)
                        getCandles()
                        deleteCandles()
                        deleteStats()
                        insertStats()
                        updateStats()
			processMinMax()
                        calculateLimit()
                        calculateTrends()
                        writeCharts()
                        os.Exit(0)
                }
                if a1 == "candles" {
			time.Sleep(15 * time.Second)
                        getCandles()
                        deleteCandles()
                        os.Exit(0)
                }
                if a1 == "updatestats" {
                        deleteStats()
                        insertStats()
                        updateStats()
                        calculateLimit()
                        calculateTrends()
                        os.Exit(0)
                }
                if a1 == "limits" {
                        calculateLimit()
                        os.Exit(0)
                }
                if a1 == "doorder" {
                        buyOrders()
			sellOrders()
                        os.Exit(0)
                }
                if a1 == "forceorder" {
                        forceOrder(os.Args[3],os.Args[4],os.Args[5])
                        os.Exit(0)
                }
                if a1 == "ticker" {
                        runTicker(os.Args[3],os.Args[4],os.Args[5],os.Args[6],os.Args[7])
                        os.Exit(0)
                }
                if a1 == "telegram" {
                        sendTelegram()
                        os.Exit(0)
                }
                if a1 == "totp" {
                        genTotp()
                        os.Exit(0)
                }
                if a1 == "testtotp" {
                        testTotp(os.Args[3])
                        os.Exit(0)
                }
                if a1 == "allowtrade" {
                        allowTrade(os.Args[3])
                        os.Exit(0)
                }
                if a1 == "account" {
			deleteAccounts()
                        readAccount()
                        os.Exit(0)
                }
                if a1 == "orders" {
                        readOrders()
			processOrders()
			deleteOrders()
                        os.Exit(0)
                }
                if a1 == "trend" {
                        calculateTrends()
                        os.Exit(0)
                }
                if a1 == "chart" {
                        writeCharts()
                        os.Exit(0)
                }
                if a1 == "positions" {
                        deactivatePositions()
                        activatePositions()
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
		fmt.Printf("Index: %d, Value: %v\n", i, v )
		out:=getPair(v,limit1+":00",limit2+":00")
		if out == nil {
			fmt.Println("Result empty")
			continue
		}
		err := json.Unmarshal(out, &cand)
	        if err != nil { // Handle JSON errors 
        	        fmt.Printf("JSON error: %v\n", err)
			fmt.Printf("JSON input: %v\n",string(out))
			continue
        	}
		if cand["start"] != nil {
			start = cand["start"].(string)
		}
		if cand["end"] != nil {
                	end = cand["end"].(string)
		}
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
				if cn["volume"] != nil {
                	        	volume = cn["volume"].(float64)
				} else {
					volume = 0
				}
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

func readAccount() {
	var cur string

	output := getAccount()
	str := string(output)
	lines := strings.Split(str, "\n")
	for _, value := range lines {
		value = strings.TrimSpace(value)
		words := strings.Split(value, " ")
		if strings.Contains(words[0], "currency"){
			cur = strings.Trim(words[1], "\",")
		}
                if strings.Contains(words[0], "total"){
                        tot := strings.Trim(words[1], "\",")
			if total, err := strconv.ParseFloat(tot, 64); err == nil {
    				fmt.Printf("Curr: %s, Total: %f\n",cur,total)
				storeAccount(exchange_name, cur, total)
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
        UPDATE yourcandle
        set open = $1, high = $2, low = $3, close = $4, volume = $5
        where exchange = $6 and pair = $7 and asset = $8 and interval = $9 and timestamp = $10`
        info, err := db.Exec(sqlStatement, open, high, low, close, volume, exchange, pair, asset, interval, timest)
        count, err := info.RowsAffected()
        if err != nil {
                panic(err)
        }
        if count == 0 {
		sqlStatement := `
		INSERT INTO yourcandle (
		exchange, pair, interval, timestamp, open, high, low, close, volume, asset)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
		_, err = db.Exec(sqlStatement, exchange, pair, interval, timest, open, high, low, close, volume, asset)
		if err != nil {
  			fmt.Printf("SQL error: %v\n",err)
		}
	}
}

func insertPositions(exchange string, pair string, trtype string, timest time.Time, rate float64, amount float64) {
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        INSERT INTO yourposition (exchange, pair, trtype, timestamp, rate, amount)
        VALUES ($1, $2, $3, $4, $5, $6)`
        _, err = db.Exec(sqlStatement, exchange, pair, trtype, timest, rate, amount)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func deletePositions(exchange string, pair string) {
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        DELETE FROM yourposition
        WHERE exchange = $1 and pair = $2 and active = true;`
        _, err = db.Exec(sqlStatement, exchange, pair)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func storeAccount(exchange string, currency string, amount float64) {
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        UPDATE youraccount
	set amount = $1
        where exchange = $2 and currency = $3`
        info, err := db.Exec(sqlStatement, amount, exchange, currency)
	count, err := info.RowsAffected()
    	if err != nil {
        	panic(err)
    	}
        if count == 0 { 
	        sqlStatement := `
        	INSERT INTO youraccount (exchange, currency, amount)
        	VALUES ($1, $2, $3)`
        	_, err = db.Exec(sqlStatement, exchange, currency, amount)
        	if err != nil {
                	fmt.Printf("SQL error: %v\n",err)
        	}
        }
}

func storeOrder(exchange string,id string,pair string,asset string,side string,otype string,
		timest float64,status string,price float64,amount float64) {

	var tst time.Time = time.Unix(int64(timest), 0)

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        UPDATE yourorder
        set pair = $3, asset = $4, price = $5, amount = $6, side = $7, timestamp = $8, order_type = $9, status = $10
        where exchange = $1 and id = $2`
        info, err := db.Exec(sqlStatement, exchange, id, pair, asset, price ,amount, side, tst, otype, status)
        if err != nil {
                panic(err)
        }
        count, err := info.RowsAffected()
        if err != nil {
                panic(err)
        }
        if count == 0 {
                sqlStatement := `
                INSERT INTO yourorder (exchange, id, pair, asset, price ,amount, side, timestamp, order_type, status)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8 ,$9, $10)`
                _, err = db.Exec(sqlStatement, exchange, id, pair, asset, price, amount, side, tst, otype, status)
                if err != nil {
                        fmt.Printf("SQL error: %v\n",err)
                }
        }
}

func insertStats() {
	var advicePeriod int64 = 7 * 24
        var intv string = "'168 hours'"

	parm,err := getParms("AdvicePeriod")
	if err == nil {
		advicePeriod = parm.intp
		intv = fmt.Sprintf("'%d hours'",advicePeriod)
	} 

	fmt.Printf("Insert statistics %s\n",intv)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
	insert into yourlimits (exchange,pair, min,avg,max,count,potwin)
	(select exchange,pair, min(close) as min, avg(close) as avg, max(close) as max, count(close) as count,
	(max(close) - min(close)) * $1 / min(close) as potwin
	from yourcandle 
	where "timestamp" > current_timestamp - interval ` + intv +  `
	AND LOWER(exchange) = $2
	group by exchange,pair
	order by pair);`
        _, err = db.Exec(sqlStatement,limit_depth,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func calculateMinMax(pair string) {
        var advicePeriod int64 = 7 * 24
        var intv string = "'168 hours'"
	var min float64
	var max float64
	var last float64

        parm,err := getParms("AdvicePeriod")
        if err == nil {
                advicePeriod = parm.intp
                intv = fmt.Sprintf("'%d hours'",advicePeriod)
        }

        fmt.Printf("Calculate MinMax for pair %s\n",pair)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select close
        from yourcandle
        where "timestamp" > current_timestamp - interval ` + intv +  `
        AND LOWER(exchange) = $1
	AND pair = $2
        order by close;`
        rows, err := db.Query(sqlStatement,exchange_name,pair)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

        var i int = 0;
        for rows.Next(){
		var cls float64
                if err := rows.Scan(&cls); err != nil {
                        fmt.Println(err)
                } else {
			if i == 1 {
				min = cls
			}
                        if i > 0 {
                                max = last
                        }
			last = cls
		}
		i++
        }
        if err := rows.Err(); err != nil {
                fmt.Println(err)
        }
	fmt.Printf("Min: %f, Max: %f\n",min,max)

        sqlStatement = `
        UPDATE yourlimits 
	SET min = $1, max = $2
        where pair = $3
        AND LOWER(exchange) = $4;`
        _, err = db.Exec(sqlStatement,min,max,pair,exchange_name)
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
	select close, pair, ((close - open) * 100 / open) as lastcandle from yourcandle y 
	where timestamp =
	(select max(timestamp) from yourcandle where LOWER(exchange) = $1)
	AND LOWER(exchange) = $1
	)
	UPDATE yourlimits l
	SET current = subquery.close,
	lastcandle = subquery.lastcandle
	FROM subquery
	WHERE l.pair = subquery.pair
	AND LOWER(l.exchange) = $1;`
        _, err = db.Exec(sqlStatement,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func calculateLimit() {
	var depth int = (100 - limit_depth)/2

        fmt.Println("Calculate limit")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        with subquery as (
        select
        pair,
        (min + (max - min)*$1/100) as limitbuy,
        (max - (max - min)*$1/100) as limitsell
        from yourlimits where LOWER(exchange) = $2
	)
        UPDATE yourlimits l
        SET "limitbuy" = subquery.limitbuy,
	    "limitsell" = subquery.limitsell
        FROM subquery
        WHERE l.pair = subquery.pair
	AND LOWER(l.exchange) = $2;`
        _, err = db.Exec(sqlStatement,depth,exchange_name)
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
        delete from yourlimits
	WHERE LOWER(exchange) = $1;`
        _, err = db.Exec(sqlStatement,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func deleteAccounts() {
        fmt.Println("Delete accounts")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        delete from youraccount
	WHERE exchange = $1;`
        _, err = db.Exec(sqlStatement,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func deleteOrders() {
        fmt.Println("Delete orders")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        delete from yourorder
	where status in ('CANCELLED','CANCELED')
	AND LOWER(exchange) = $1;`
        _, err = db.Exec(sqlStatement,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func deleteCandles() {
        fmt.Println("Delete candles")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        delete from yourcandle
        where "timestamp" < current_timestamp - interval '30 days'`
        _, err = db.Exec(sqlStatement)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func getPair(p string, s string, e string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "gethistoriccandlesextended",
        "-e",exchange_name,"-a","SPOT","-p",p,"-i","900",
        "--start",s,"--end",e).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
	return out
}

func getAccount() []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getaccountinfo",
        "--exchange",exchange_name,"--asset","SPOT").Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
        return out
}

func getOrder(pair string,id string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getorder",
        "--exchange",exchange_name,"--asset","SPOT","--pair",pair,"--order_id",id).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
        return out
}

func getOrders(pair string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getorders",
        "--exchange",exchange_name,"--asset","SPOT","--pair",pair).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
        return out
}

func submitOrder(pair string,side string,otype string,amount float64,price float64,clientid string) []byte {
        var resp map[string]interface{}
	var stramount string
	var strprice string 
	var aformat string = "%."+amountcomma[strings.ToLower(pair)]+"f"
        var pformat string = "%."+pricecomma[strings.ToLower(pair)]+"f"
	apot, _ := strconv.ParseInt(amountcomma[strings.ToLower(pair)], 10, 32)
	fac := math.Pow(float64(10),float64(apot))
	amount = float64(math.Floor(amount*fac)/fac)

	stramount = fmt.Sprintf(aformat,amount)
	strprice = fmt.Sprintf(pformat,price)

	fmt.Printf("A: %s, P: %s\n",stramount,strprice)

        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "submitorder",
        "--exchange",exchange_name,"--asset","SPOT","--pair",pair,"--side",side,"--type",otype,
	"--amount",stramount,"--price",strprice,"--client_id",clientid).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        } else {
                err := json.Unmarshal(out, &resp)
                if err != nil { // Handle JSON errors
                	fmt.Printf("JSON error: %v\n", err)
                	fmt.Printf("JSON input: %v\n",string(out))
                } else {
//                	status = strings.ToLower(resp["exchange"].(string))
			if resp["order_id"] != nil {
                		id := resp["order_id"].(string)
  	                	storeOrder(exchange_name,id,pair,"SPOT",side,otype,float64(time.Now().Unix()),"NEW",price,amount)
			} else if resp["id"] != nil {
                                id := resp["id"].(string)
                                storeOrder(exchange_name,id,pair,"SPOT",side,otype,float64(time.Now().Unix()),"NEW",price,amount)
                        } else {
				fmt.Printf("Submit error: P: %s, S: %s, T: %s, A: %s, Pr: %s\n",pair,side,otype,stramount,strprice)
			}
		}
	}
        return out
}

func cancelOrder(pair string,oid string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "cancelorder",
        "--exchange",exchange_name,"--asset","SPOT","--pair",pair,"--order_id",oid).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
        return out
}

func insertParms(key string, intp int64, floatp float64, stringp string, datep time.Time, timep time.Time, timestampp time.Time ) {
        fmt.Printf("Insert parm %v\n",key)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        insert into yourparameter (key, "int", "float", "string", "date", "time", "timestamp")
	values ($1,$2,$3,$4,$5,$6,$7);`
        _, err = db.Exec(sqlStatement,key,intp,floatp,stringp,datep,timep,timestampp)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func getParms(key string) (parms Parm, err error) {
        fmt.Printf("Get parm %v\n",key)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()
 
	parms.key = ""

        sqlStatement := `
        select "key", "int", "float", "string", "date", "time", "timestamp"  from yourparameter 
	where key = $1;`

	err = db.QueryRow(sqlStatement, key).Scan(&parms.key,&parms.intp,&parms.floatp,&parms.stringp,&parms.datep,&parms.timep,&parms.timestampp)
	if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
	return
}

func getBuyPriceNew(pair string) (price float64, amount float64, err error) {
	var current float64
	var limit float64
	var min float64
	var potwin float64
	var trend1 float64
        var trend2 float64
        var trend3 float64

	price = 0
	amount = 0

        fmt.Printf("Get buy price %v\n",pair)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select l.limitbuy, l.current, l.min, l.potwin, l.trend1, l.trend2, l.trend3 from yourlimits l 
        where l.pair = $1
	AND LOWER(l.exchange) = $2
        and l.pair not in
	(select distinct p.pair from yourposition p, yourlimits x
	where p.pair = $1
	AND LOWER(p.exchange) = $2
	and LOWER(x.exchange) = LOWER(p.exchange)
	and x.pair = p.pair
	and p.notrade = false
	AND 
	(p.active = true
	or 
	p.rate > (x.current * 0.98) and 
	p.rate < (x.current * 1.02)
	)
        );`

        err = db.QueryRow(sqlStatement, pair, exchange_name).Scan(&limit,&current,&min,&potwin,&trend1,&trend2,&trend3)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }

	fmt.Printf("C: %f, L: %f, M: %f, P: %f, T1: %f, T2: %f, T3: %f\n",current,limit,min,potwin,trend1,trend2,trend3)

	if (current < limit) && (potwin > float64(minwin) + 1) {
		var dobuy bool = false
		if (current >= min) && (trend2 < -1) && (trend1 > 0.1) {
			dobuy = true
			fmt.Println("Rule 1")
		} else if (current >= min) && (trend3 < -0.3) && (trend2 < -0.1) && (trend2 > -0.5) && (trend1 > 0.1) {
                        dobuy = true
                        fmt.Println("Rule 2")
                } else if (current >= min) && (trend2 < 0.7) && (trend2 > -0.7) && (trend1 > 0.1) {
                        dobuy = true
                        fmt.Println("Rule 3")
                } else {
                        fmt.Println("wait due to trend")
			if useticker {
        			_ = exec.Command(os.Args[0], exchange_name, "ticker", pair,
        				"BUY",fmt.Sprintf("%f",limit),fmt.Sprintf("%f",current),"0").Start()
				fmt.Printf("BUY ticker started -- P: %s, L: %f, C: %f\n", pair, limit, current)
			}
		} 
		if dobuy {
			price = limit
			amount = float64(invest_amount)/price
                	fmt.Printf("Price: %f\n",limit)
        	} 
	}

	fmt.Printf("A: %f, P: %f\n",amount,price)

        return
}


func getSellPrice(pair string) (price float64, amount float64, err error) {
	var rate float64
	var current float64
	var limit float64
	var amnt float64
	var max float64
	var trend1 float64
        var trend2 float64
        var trend3 float64
	var winrate float64

	price = 0
	amount = 0

        fmt.Printf("Get sell price %v\n",pair)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select l.limitsell, l.current, l.max, l.trend1, l.trend2, l.trend3, p.rate, p.amount*0.995 as amount from yourlimits l, yourposition p 
        where l.pair = $1
	AND LOWER(l.exchange) = $2
        and p.pair = $1
	AND LOWER(p.exchange) = $2
	AND p.active = true
	AND notrade = false;`

        err = db.QueryRow(sqlStatement, pair, exchange_name).Scan(&limit,&current,&max,&trend1,&trend2,&trend3,&rate,&amnt)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }

	winrate = rate * ((100 + float64(minwin)) / 100)

	fmt.Printf("C: %f, L: %f, M: %f, R: %f, T1: %f, T2: %f, T3: %f\n",current,limit,max,winrate,trend1,trend2,trend3)

	if winrate > limit {
		limit = winrate
	}

	if current > limit {
                var dosell bool = false
                if trend1 < 0.1 {
                        dosell = true
                        fmt.Println("Rule 1")
                }
                if dosell {
                        amount = amnt
                        price = limit
                        fmt.Printf("Price: %f\n",limit)
                } else {
                        fmt.Println("Wait due to trend")
                        if useticker {
                                _ = exec.Command(os.Args[0], exchange_name, "ticker", pair,
                                        "SELL",fmt.Sprintf("%f",limit),fmt.Sprintf("%f",current),fmt.Sprintf("%f",amnt)).Start()
                                fmt.Printf("Sell ticker started -- P: %s, L: %f, C: %f\n", pair, limit, current)
                        }
		}
        }

	fmt.Printf("A: %f, P: %f\n",amount,price)

        return
}

func submitTelegram(msg string) {
        f, err := os.OpenFile(pipeFile, os.O_WRONLY|syscall.O_NONBLOCK, 0644)
        if err != nil {
                fmt.Printf("open: %v\n", err)
        }
        defer f.Close()

        _, err = f.WriteString(msg+"\n")

        if err != nil {
	        fmt.Println(err)
        }
}

func deleteParms(key string) (parms Parm, err error) {
        fmt.Printf("Delete parm %v\n",key)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        delete from yourparameter
        where key = $1;`

        _, err = db.Exec(sqlStatement, key)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        return
}

func sendTelegram(){
	var mess string

	if err := os.Remove(pipeFile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("remove: %v\n", err)
	}
	if err := syscall.Mkfifo(pipeFile, 0644); err != nil {
		fmt.Printf("mkfifo: %v\n", err)
	}

	f, err := os.OpenFile(pipeFile, os.O_RDONLY|syscall.O_NONBLOCK, 0644)
	if err != nil {
		fmt.Printf("open: %v\n", err)
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	bot, err := tgbotapi.NewBotAPI(tbtoken)
	if err != nil {
		fmt.Printf("Telegram error: %v\n",err)
		return
	}

	bot.Debug = false

	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	parms,err := getParms("ChatID")
	if err == nil {
		msg := tgbotapi.NewMessage(parms.intp, "The bot is active")
		bot.Send(msg)
	}

	for true {
		mess = ""
                line, err := reader.ReadBytes('\n')
                if err == nil {
			m := string(line)
			m = strings.ReplaceAll(m, "|", "\n")
                        fmt.Printf("%v Message: %s", time.Now().String(), m)
                	msg := tgbotapi.NewMessage(parms.intp, m)
                	bot.Send(msg)
                }
       		select {
 			case update := <-updates:
  				fmt.Printf("%v [%s] %s\n", time.Now().String(), update.Message.From.UserName, update.Message.Text)
                		insertParms("ChatID", update.Message.Chat.ID, 0, "", time.Now(), time.Now(), time.Now())
				argParts := strings.Split(update.Message.Text, " ")
                                if  argParts[0] == "Adviceperiod" {
					var passcode string
					period, err := strconv.Atoi(argParts[1])
					if len(argParts) > 2 {
                                        passcode = argParts[2]
					} else {
						passcode = ""
					}
                                        valid := totp.Validate(passcode, key_secret)
					if err == nil && valid {
						fmt.Println(period)
						deleteParms("AdvicePeriod")
 						insertParms("AdvicePeriod", int64(period), 0, "", time.Now(), time.Now(), time.Now())
						mess = "command successful"
					} else {
                                        	mess = "command failed"
					}
                                }
                                if  argParts[0] == "Limitdepth" {
                                        var passcode string
                                        period, err := strconv.Atoi(argParts[1])
                                        if len(argParts) > 2 {
                                        passcode = argParts[2]
                                        } else {
                                                passcode = ""
                                        }
                                        valid := totp.Validate(passcode, key_secret)
                                        if err == nil && valid {
                                                fmt.Println(period)
                                                deleteParms("limit_depth")
                                                insertParms("limit_depth", int64(period), 0, "", time.Now(), time.Now(), time.Now())
                                                mess = "command successful"
                                        } else {
                                                mess = "command failed"
                                        }
                                }
                                if  argParts[0] == "Stoptrade" {
                                        var passcode string
                                        if len(argParts) > 1 {
                                        passcode = argParts[1]
                                        } else {
                                                passcode = ""
                                        }
        				valid := totp.Validate(passcode, key_secret)
                                        if valid {
                                                deleteParms("DoTrade")
                                                mess = "command successful"
                                        } else {
                                                mess = "command failed"
                                        }
                                }
				if  mess != "" {
                			msg := tgbotapi.NewMessage(update.Message.Chat.ID, mess)
                			msg.ReplyToMessageID = update.Message.MessageID
		                	bot.Send(msg)
					fmt.Printf("%s Sent: %s\n", time.Now().String(),mess)
				}
 			default:
 		}
		time.Sleep(30 * time.Second)
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
	sleepbeforerun = viper.GetInt("sleepbeforerun")

        exchange_name = viper.GetString("exchange_name")

	gctcmd = viper.GetString("gctcmd")

        wwwpath = viper.GetString("wwwpath")

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

	useticker = viper.GetBool("useticker")

        parm,err := getParms("limit_depth")
        if err == nil {
                limit_depth = int(parm.intp)
        }

	if limit_depth > 90 || limit_depth < 10 {
		limit_depth = 80
	}

        if invest_amount > 500 || limit_depth < 50 {
                invest_amount = 100
        }

	if do_trace {
		fmt.Println("do_trace: ",do_trace)
                fmt.Printf("sleepbeforerun: %d\n",sleepbeforerun)
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

func readOrders() {
        var pack map[string]interface{}
	var o Order

        for i, v := range tradepairs {
                fmt.Printf("%d, Value: %v\n", i, v )
		out := getOrders(v)

		if len(string(out)) < 10 {
			fmt.Println("No open order")
			continue
		}

                err := json.Unmarshal(out, &pack)
                if err != nil { // Handle JSON errors
                        fmt.Printf("JSON error: %v\n", err)
                        fmt.Printf("JSON input: %v\n",string(out))
                        continue
                }
		orders := pack["orders"].([]interface{})
                for _, order := range orders {
			ord := order.(map[string]interface{})
			o.exchange = ord["exchange"].(string)
                        o.id = ord["id"].(string)
                        o.base_currency = ord["base_currency"].(string)
                        o.quote_currency = ord["quote_currency"].(string)
			if ord["asset_type"] != nil { 
				o.asset = ord["asset_type"].(string)
			}
                        o.order_side = ord["order_side"].(string) 
                        o.order_type = ord["order_type"].(string)
                        o.creation_time = ord["creation_time"].(float64)
			if ord["update_time"] != nil {
                        	o.update_time = ord["update_time"].(float64)
			}
			if ord["status"] != nil {
                        	o.status = ord["status"].(string)
			}
                        o.price = ord["price"].(float64)
                        o.amount = ord["amount"].(float64)
                        o.open_volume = ord["open_volume"].(float64)
                        fmt.Println(o)

			storeOrder(o.exchange,o.id,o.base_currency+"-"+o.quote_currency,o.asset,o.order_side,o.order_type,o.update_time,o.status,o.price,o.amount) 
		} 
        }
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("cron         Do regular work")
     fmt.Println("limits       Calculate new limits")
     fmt.Println("orders       Get open orders")
     fmt.Println("telegram     Start telegram daemon")
     fmt.Println("updatestats  Update statistics")
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

func trend1(pair string) {
        var wr bool = false
	var tm1 int64
	var coeff float64
	var cls float64

        fmt.Printf("Calculate trend1 %v\n",pair)

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        r := new(regression.Regression)
        r.SetObserved("Close")
        r.SetVar(0, "Timestamp")

        sqlStatement := `
        select "timestamp", "close"  from yourcandle
        where pair = $1
	AND LOWER(exchange) = $2
        and "timestamp" > current_timestamp - interval '1 hour'
	order by "timestamp";`

        rows, err := db.Query(sqlStatement, pair, exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

	var i int = 0;
        for rows.Next(){
                var tmstamp time.Time
		var tmu int64
                if err := rows.Scan(&tmstamp, &cls); err != nil {
                        fmt.Println(err)
                }
                tmu = tmstamp.Unix()
		if i==0 {
			tm1 = tmu
		}
		tmu = tmu - tm1
		i++
	        r.Train(
			regression.DataPoint(cls, []float64{float64(tmu)}),
		)
                wr = true
                if err != nil {
                        fmt.Println(err)
                }
        }
        if err := rows.Err(); err != nil {
                fmt.Println(err)
        }
        if wr {
                r.Run()
                fmt.Printf("Regression formula:\n%v\n", r.Formula)
                fmt.Printf("Coeff: %f\n", r.Coeff(0))
                fmt.Printf("Coeff: %f\n", r.Coeff(1))
                coeff = r.Coeff(1) * 360000 / cls
                sqlStatement = `
                UPDATE yourlimits
                SET trend1 = $1
                WHERE pair = $2
		AND LOWER(exchange) = $3;`
                _, err = db.Exec(sqlStatement,coeff,pair,exchange_name)
                if err != nil {
                        fmt.Printf("SQL error: %v\n",err)
                }
        }
}

func trend3(pair string) {
        var wr bool = false
        var tm1 int64
	var coeff float64
	var cls float64

        fmt.Printf("Calculate trend3 %v\n",pair)

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        r := new(regression.Regression)
        r.SetObserved("Close")
        r.SetVar(0, "Timestamp")

        sqlStatement := `
        select "timestamp", "close"  from yourcandle
        where pair = $1
	AND LOWER(exchange) = $2
        and "timestamp" > current_timestamp - interval '4 hours'
        and "timestamp" <= current_timestamp - interval '1 hours'
        order by "timestamp";`

        rows, err := db.Query(sqlStatement, pair, exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

        var i int = 0;
        for rows.Next(){
                var tmstamp time.Time
                var tmu int64
                if err := rows.Scan(&tmstamp, &cls); err != nil {
                        fmt.Println(err)
                }
                tmu = tmstamp.Unix()
                if i==0 {
                        tm1 = tmu
                }
                tmu = tmu - tm1
                i++
                r.Train(
                        regression.DataPoint(cls, []float64{float64(tmu)}),
                )
                wr = true
                if err != nil {
                        fmt.Println(err)
                }
        }
        if err := rows.Err(); err != nil {
                fmt.Println(err)
        }
        if wr {
                r.Run()
                fmt.Printf("Regression formula:\n%v\n", r.Formula)
                fmt.Printf("Coeff: %f\n", r.Coeff(0))
                fmt.Printf("Coeff: %f\n", r.Coeff(1))
                coeff = r.Coeff(1) * 360000 / cls
                sqlStatement = `
                UPDATE yourlimits
                SET trend3 = $1
                WHERE pair = $2
		AND LOWER(exchange) = $3;`
                _, err = db.Exec(sqlStatement,coeff,pair,exchange_name)
                if err != nil {
                        fmt.Printf("SQL error: %v\n",err)
                }
        }
}

func trend2(pair string) {
	var wr bool = false
        var tm1 int64
	var coeff float64
	var cls float64

        fmt.Printf("Calculate trend2 %v\n",pair)

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        r := new(regression.Regression)
        r.SetObserved("Close")
        r.SetVar(0, "Timestamp")

        sqlStatement := `
        select "timestamp", "close"  from yourcandle
        where pair = $1
	AND LOWER(exchange) = $2
        and "timestamp" > current_timestamp - interval '2 hours'
	and "timestamp" <= current_timestamp - interval '1 hours'
        order by "timestamp";`

        rows, err := db.Query(sqlStatement, pair, exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

        var i int = 0;
        for rows.Next(){
                var tmstamp time.Time
                var tmu int64
                if err := rows.Scan(&tmstamp, &cls); err != nil {
                        fmt.Println(err)
                }
                tmu = tmstamp.Unix()
                if i==0 {
                        tm1 = tmu
                }
                tmu = tmu - tm1
                i++
                r.Train(
                        regression.DataPoint(cls, []float64{float64(tmu)}),
                )
		wr = true
                if err != nil {
                        fmt.Println(err)
                }
        }
        if err := rows.Err(); err != nil {
                fmt.Println(err)
        }
        if wr {
                r.Run()
                fmt.Printf("Regression formula:\n%v\n", r.Formula)
                fmt.Printf("Coeff: %f\n", r.Coeff(0))
                fmt.Printf("Coeff: %f\n", r.Coeff(1))
		coeff = r.Coeff(1) * 360000 / cls
	        sqlStatement = `
        	UPDATE yourlimits 
        	SET trend2 = $1
        	WHERE pair = $2
		AND LOWER(exchange) = $3;`
        	_, err = db.Exec(sqlStatement,coeff,pair,exchange_name)
        	if err != nil {
                	fmt.Printf("SQL error: %v\n",err)
        	}
	}
}

func calculateTrends() {
        for _, v := range pairs {
		trend1(v)
		trend2(v)
		trend3(v)
	}
}

func processOrders() {
        fmt.Println("Process orders")

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select id, pair, asset, price, amount, side, timestamp, order_type from yourorder
        where status in ('NEW','OPEN')
	AND LOWER(exchange) = $1
        order by "timestamp";`

        rows, err := db.Query(sqlStatement,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

        for rows.Next(){
		var pair string
		var crtime time.Time
        	var order map[string]interface{}
		var o Order
                if err := rows.Scan(&o.id, &pair, &o.asset, &o.price, &o.amount, &o.order_side, &crtime, &o.order_type); err != nil {
                        fmt.Println(err)
                } else {
			fmt.Printf("Pair: %s, OrderID: %s\n",pair, o.id)
			out :=getOrder(pair, o.id)
			currencies := strings.Split(pair, "-")
			o.base_currency = currencies[0]
                        o.quote_currency = currencies[1]
			o.update_time = float64(crtime.Unix())
                	err := json.Unmarshal(out, &order)
                	if err != nil { // Handle JSON errors
                        	fmt.Printf("JSON error: %v\n", err)
                        	fmt.Printf("JSON input: %v\n",string(out))
                        	continue
                	}
			if order["exchange"] != nil {
				o.exchange = strings.ToLower(order["exchange"].(string))
			}
                        o.id = order["id"].(string)
			if order["base_currency"] != nil {
                        	o.base_currency = order["base_currency"].(string)
			}
			if order["quote_currency"] != nil {
                        	o.quote_currency = order["quote_currency"].(string)
			}
                        if order["asset_type"] != nil {
				o.asset = order["asset_type"].(string)
			}
			if order["order_side"] != nil {
                        	o.order_side = order["order_side"].(string)
			} 
			if order["order_type"] != nil {
                        	o.order_type = order["order_type"].(string)
			}
			if order["creation_time"] != nil {
                        	o.creation_time = order["creation_time"].(float64)
			}
			if order["update_time"] != nil {
	                        o.update_time = order["update_time"].(float64)
			}
                        o.status = order["status"].(string)
			if order["price"] != nil {
                        	o.price = order["price"].(float64)
			}
			if order["amount"] != nil {
                        	o.amount = order["amount"].(float64)
			}
			if order["cost"] != nil {
                        	o.cost = order["cost"].(float64)
			}
                        fmt.Println(o)

			storeOrder(o.exchange,o.id,o.base_currency+"-"+o.quote_currency,o.asset,o.order_side,o.order_type,o.update_time,o.status,o.price,o.amount)

			if (o.status == "FILLED" || o.status == "CLOSED" || o.status == "FINISHED") && o.order_side == "BUY" {
				insertPositions(o.exchange,o.base_currency+"-"+o.quote_currency,o.order_side,time.Unix(int64(o.update_time), 0),o.price,o.amount)
                        	submitTelegram("Position: "+o.base_currency+"-"+o.quote_currency+" bought")
			}
                        if (o.status == "FILLED" || o.status == "CLOSED" || o.status == "FINISHED") && o.order_side == "SELL" {
				deletePositions(o.exchange, o.base_currency+"-"+o.quote_currency)
                                submitTelegram("Position: "+o.base_currency+"-"+o.quote_currency+" sold")
                        }
		}
        }
}

func buyOrders() {
        var pack map[string]interface{}
        var resp map[string]interface{}
        var o Order

        parm,err := getParms("DoTrade")
        if err != nil {
                fmt.Println("Trades not allowed")
                return
        }
        if parm.intp != 13579 {
                fmt.Println("Trades not allowed")
                return
        }

        for i, v := range tradepairs {
                fmt.Printf("%d, Value: %v\n", i, v )
                out := getOrders(v)

                if len(string(out)) < 25 {
                        fmt.Println("No open order")
                        newprice,newamount,err := getBuyPriceNew(v)
                        if err != nil {
                                fmt.Printf("Price error: %v\n", err)
                                continue
                        }
                        if newprice > 0 {
                                fmt.Printf("Price: %f, Amount: %f\n",newprice,float64(invest_amount)/newprice)
                                out := submitOrder(v,"BUY","LIMIT",newamount,newprice,"automatic-new")
                                fmt.Println(string(out))
                        }
                        continue
                }

                err := json.Unmarshal(out, &pack)
                if err != nil { // Handle JSON errors
                        fmt.Printf("JSON error: %v\n", err)
                        fmt.Printf("JSON input: %v\n",string(out))
                        continue
                }
                orders := pack["orders"].([]interface{})
                for _, order := range orders {
                        ord := order.(map[string]interface{})
                        o.exchange = ord["exchange"].(string)
                        o.id = ord["id"].(string)
                        o.order_side = ord["order_side"].(string)
                        o.price = ord["price"].(float64)
			if o.order_side == "BUY" {
	                        fmt.Printf("E: %v, ID: %v, P: %f\n",o.exchange,o.id,o.price)
				out := cancelOrder(v,o.id)
				fmt.Println(string(out))
				err := json.Unmarshal(out, &resp)
		                if err != nil { // Handle JSON errors
                		        fmt.Printf("JSON error: %v\n", err)
                        		fmt.Printf("JSON input: %v\n",string(out))
                        		continue
                		}
				status := resp["status"].(string)
				if status == "success" {
		                        newprice,newamount,err := getBuyPriceNew(v)
                		        if err != nil {
                                		fmt.Printf("Price error: %v\n", err)
                                		continue
                        		}
                        		if newprice > 0 {
                                		fmt.Printf("Price: %f, Amount: %f\n",newprice,float64(invest_amount)/newprice)
                                		out := submitOrder(v,"BUY","LIMIT",newamount,newprice,"automatic-new")
                                		fmt.Println(string(out))
                        		}
				}
			}
                }
        }
}

func sellOrders() {
        var pack map[string]interface{}
        var resp map[string]interface{}
        var o Order

        parm,err := getParms("DoTrade")
        if err != nil {
		fmt.Println("Trades not allowed 1")
                return
        }
        if parm.intp != 13579 {
                fmt.Println("Trades not allowed 2")
                return
        }

        for i, v := range tradepairs {
                fmt.Printf("%d, Value: %v\n", i, v )
                out := getOrders(v)

                if len(string(out)) < 25 {
                        fmt.Println("No open order")
                        newprice,newamount,err := getSellPrice(v)
                        if err != nil {
                                fmt.Printf("Price error: %v\n", err)
                                continue
                        }
			if newprice > 0 {
                        	fmt.Printf("Price: %f, Amount: %f\n",newprice,float64(invest_amount)/newprice)
                        	out := submitOrder(v,"SELL","LIMIT",newamount,newprice,"automatic-new")
                        	fmt.Println(string(out))
			}
                        continue
                }

                err := json.Unmarshal(out, &pack)
                if err != nil { // Handle JSON errors
                        fmt.Printf("JSON error: %v\n", err)
                        fmt.Printf("JSON input: %v\n",string(out))
                        continue
                }
                orders := pack["orders"].([]interface{})
                for _, order := range orders {
                        ord := order.(map[string]interface{})
                        o.exchange = ord["exchange"].(string)
                        o.id = ord["id"].(string)
                        o.order_side = ord["order_side"].(string)
                        o.price = ord["price"].(float64)
                        if o.order_side == "SELL" {
                                fmt.Printf("E: %v, ID: %v, P: %f\n",o.exchange,o.id,o.price)
                                out := cancelOrder(v,o.id)
                                fmt.Println(string(out))
                                err := json.Unmarshal(out, &resp)
                                if err != nil { // Handle JSON errors
                                        fmt.Printf("JSON error: %v\n", err)
                                        fmt.Printf("JSON input: %v\n",string(out))
                                        continue
                                }
                                status := resp["status"].(string)
                                if status == "success" {
                                        newprice,newamount,err := getSellPrice(v)
                                        if err != nil {
                                                fmt.Printf("Price error: %v\n", err)
                                                continue
                                        }
                                        fmt.Printf("Price: %f, Amount: %f\n",newprice,newamount)
                                        out := submitOrder(v,"SELL","LIMIT",newamount,newprice,"automatic-update")
                                        fmt.Println(string(out))
                                }
                        }
                }
        }
}

func genTotp() {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      key_issuer,
		AccountName: key_account,
	})
	if err != nil {
		panic(err)
	}
	// Convert TOTP key into a PNG
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		panic(err)
	}
	png.Encode(&buf, img)
}

func display(key *otp.Key, data []byte) {
	fmt.Printf("Issuer:       %s\n", key.Issuer())
	fmt.Printf("Account Name: %s\n", key.AccountName())
	fmt.Printf("Secret:       %s\n", key.Secret())
	fmt.Println("Writing PNG to qr-code.png....")
	ioutil.WriteFile("/tmp/qr-code.png", data, 0644)
}

func testTotp(passcode string) {
	// Now Validate that the user's successfully added the passcode.
	fmt.Println("Validating TOTP...")
	valid := totp.Validate(passcode, key_secret)
	if valid {
		println("Valid passcode!")
	} else {
		println("Invalid passcode!")
	}
}

func allowTrade(passcode string) {
        valid := totp.Validate(passcode, key_secret)
        if valid {
        	insertParms("DoTrade", 13579, 0, "", time.Now(), time.Now(), time.Now())
                println("Trade allowed")
        } else {
                println("Wrong passcode!")
        }
}

func writeChart(pair string) {
        var tmstp time.Time
        var value float64

        fmt.Printf("Write chart %s\n",pair)

	f, err := os.Create(wwwpath+"/"+exchange_name+"-"+pair+".html")

	if err != nil {
        	panic(err)
	}

	defer f.Close()

	header := `
<!doctype html>
<html>
<head>
  <title>Timeline</title>
  <script type="text/javascript" src="https://unpkg.com/vis-timeline@latest/standalone/umd/vis-timeline-graph2d.min.js"></script>
  <link href="https://unpkg.com/vis-timeline@latest/styles/vis-timeline-graph2d.min.css" rel="stylesheet" type="text/css" />
  <style type="text/css">
    #visualization {
      width: 1800px;
      height: 1200px;
      border: 1px solid lightgray;
    }
  </style>
</head>
<body>
<div id="visualization"></div>
<script type="text/javascript">
  var container = document.getElementById('visualization');
  var items = new vis.DataSet(
[
	`

	_, err = f.WriteString(header)

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select "timestamp", "close"  from yourcandle
        where pair = $1
	and LOWER(exchange) = $2
        and "timestamp" > current_timestamp - interval '30 days'
        order by "timestamp";`

        rows, err := db.Query(sqlStatement,pair,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

        for rows.Next(){
                if err := rows.Scan(&tmstp, &value); err != nil {
                        fmt.Println(err)
                }
		_, err = f.WriteString(fmt.Sprintf("{x: '%s', y: %f},\n",tmstp.Format("2006-01-02T15:04:05"),value)) 
        }

        footer := `
]
  );
  var options = {
    width:  '100%',
    height: '800px'
  };
  var graph2d = new vis.Graph2d(container, items, options);
</script>
</body>
</html>
        `
        _, err = f.WriteString(footer)
}

func writeCharts() {
        for _, v := range pairs {
                writeChart(v)
        }
}

func processMinMax() {
        for _, v := range pairs {
                calculateMinMax(v)
        }
}

func calculateProfit(){
	var price float64
	var amount float64
	var side string

        fmt.Println("Calculate profit")

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select price, amount, side from yourorder
        where LOWER(exchange) = $1
	AND status = 'FILLED'
        order by pair, "timestamp";`

        rows, err := db.Query(sqlStatement, exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

//        var i int = 0;
        for rows.Next(){
                if err := rows.Scan(&price, &amount, &side); err != nil {
                        fmt.Println(err)
                }
        }
        if err := rows.Err(); err != nil {
                fmt.Println(err)
        }
}

func forceOrder(pair string, price string, passcode string) {
	var newamount float64

        valid := totp.Validate(passcode, key_secret)
        if valid {
        	newprice, err := strconv.ParseFloat(price, 64)
		if err != nil {
			panic(err)
		}
		newamount = float64(invest_amount)/newprice
                fmt.Printf("Price: %f, Amount: %f\n",newprice,newamount)
                out := submitOrder(pair,"BUY","LIMIT",newamount,newprice,"force-order")
                fmt.Println(string(out))
        } else {
                println("Invalid passcode!")
        }
}

func activatePosition(pair string) {
        fmt.Printf("Activate position: %s\n", pair)
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
	update yourposition
	set active = true
	where rid =
	(select rid from yourposition p
	where LOWER(exchange) = $1
	and pair = $2
	and pair not in 
	(select pair from yourposition x
	where active = true
	AND LOWER(exchange) = $1)
	and notrade = false
	and rate < 
	(select limitsell from yourlimits l
	where LOWER(l.exchange) = $1
	and l.pair = $2)
	order by rate desc
	limit 1);`
        _, err = db.Exec(sqlStatement,exchange_name,pair)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
}

func activatePositions() {
        for _, v := range tradepairs {
                activatePosition(v)
        }
}

func deactivatePositions() {
        fmt.Println("Dectivate position")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        update yourposition
        set active = false
        where LOWER(exchange) = $1
	and notrade = false;`
        _, err = db.Exec(sqlStatement,exchange_name)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }

}

func runTicker(pair string, side string, limitstr string, currentstr string, amountstr string){
	var level float64
	var tmthen int64 = time.Now().Unix()
	var tmnow int64 = tmthen
	var done bool = false

	current, _ := strconv.ParseFloat(currentstr, 64)
        limit, _ := strconv.ParseFloat(limitstr, 64)
        amount, _ := strconv.ParseFloat(amountstr, 64)
	side = strings.ToUpper(side)
        pair = strings.ToUpper(pair)

	level = limit + ((current - limit) / 2)

	var i int64 = 0
	for i < 860 && done == false {
	        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getticker",
	        "--exchange",exchange_name,"--asset","SPOT","--pair",pair).Output()
	        if err != nil {
	                fmt.Printf("Command finished with error: %v", err)
	        } else {
//			fmt.Println(string(out))
			outstr := string(out)
			pos := strings.Index(outstr, "\"last\"") + 8
			laststr := string(outstr[pos:pos+20])
			lastarr := strings.Split(laststr, ",")
			laststr = lastarr[0]
			last, _ := strconv.ParseFloat(laststr, 64)
//			submitTelegram(fmt.Sprintf("Last: %f, Current: %f, Limit: %f, Level: %f\n",last, current, limit, level))
			if side == "BUY" && last > level {
				amount := float64(invest_amount)/limit
                                _ = submitOrder(pair,"BUY","LIMIT",amount,limit,"ticker")
				submitTelegram(fmt.Sprintf("Ticker bought %f %s at %f\n",amount,pair,limit))
				done = true
			}
                        if side == "SELL" && last < level {
                                _ = submitOrder(pair,"SELL","LIMIT",amount,limit,"ticker")
                                submitTelegram(fmt.Sprintf("Ticker sold %f %s at %f\n",amount,pair,limit))
				done = true
                        }
		} 
		time.Sleep(15 * time.Second)
		tmnow = time.Now().Unix()
		i = tmnow - tmthen 
	}
}

func jhtest() {
        err := exec.Command(os.Args[0], exchange_name, "ticker", "ETH-EUR",
        "BUY","2300","2400","0").Start()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        } else {
                fmt.Println("Everything ok")

	}
}
