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
//	"io"
	"time"
	"strings"
	"strconv"
	"syscall"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"encoding/json" 
	"github.com/spf13/viper"
	"github.com/sajari/regression"
    	"database/sql"
    	_ "github.com/lib/pq"
)

var pipeFile = "/tmp/yourpipe"

var do_trace bool = true

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
			calculateLimit()
			calculateTrends()
			deleteAccounts()
			readAccount()
			readOrders()
			processOrders()
			deleteOrders()
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
                if a1 == "telegram" {
                        sendTelegram()
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
                if a1 == "trend7" {
                        trend7("BNB-EUR")
                        os.Exit(0)
                }
                if a1 == "trend24" {
                        trend24("BNB-EUR")
                        os.Exit(0)
                }
                if a1 == "trend" {
                        trend24("ADA-EUR")
			trend7("ADA-EUR")
			trend4("ADA-EUR")
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
				storeAccount("binance", cur, total)
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
	insert into yourlimits (pair, min,avg,max,count,potwin)
	(select pair, min(close) as min, avg(close) as avg, max(close) as max, count(close) as count,
	(max(close) - min(close)) * 80 / min(close) as potwin
	from yourcandle 
	where "timestamp" > current_timestamp - interval ` + intv +  `
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

func calculateLimit() {
        fmt.Println("Calculate limit")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        with subquery as (
        select
        pair,
        (min + (max - min)/10) as limitbuy,
        (max - (max - min)/10) as limitsell
        from yourlimits
	)
        UPDATE yourlimits l
        SET "limitbuy" = subquery.limitbuy,
	    "limitsell" = subquery.limitsell
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

func deleteAccounts() {
        fmt.Println("Delete accounts")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        delete from youraccount;`
        _, err = db.Exec(sqlStatement)
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
	where status = 'CANCELLED';`
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

func getAccount() []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getaccountinfo",
        "--exchange","binance","--asset","SPOT").Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
        return out
}

func getOrder(pair string,id string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getorder",
        "--exchange","binance","--asset","SPOT","--pair",pair,"--order_id",id).Output()
        if err != nil {
                fmt.Printf("Command finished with error: %v", err)
        }
        return out
}

func getOrders(pair string) []byte {
        out, err := exec.Command(gctcmd, "--rpcuser", gctuser, "--rpcpassword", gctpassword, "getorders",
        "--exchange","binance","--asset","SPOT","--pair",pair).Output()
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
				if  update.Message.Text == "Nobuyinfo" {
                                	insertParms("Nobuyinfo", 0, 0, "", time.Now(), time.Now(), time.Now())
					mess = "command successful"
				}
                                if  update.Message.Text == "Nosellinfo" {
                                        insertParms("Nosellinfo", 0, 0, "", time.Now(), time.Now(), time.Now())
                                        mess = "command successful"
                                }
                                if  update.Message.Text == "Buyinfo" {
					deleteParms("Nobuyinfo")
                                        mess = "command successful"
                                }
                                if  update.Message.Text == "Sellinfo" {
					deleteParms("Nosellinfo")
                                        mess = "command successful"
                                }
                                if  argParts[0] == "Adviceperiod" {
                                  //      deleteParms("Nosellinfo")
					period, err := strconv.Atoi(argParts[1])
					if err == nil {
						fmt.Println(period)
						deleteParms("AdvicePeriod")
 						insertParms("AdvicePeriod", int64(period), 0, "", time.Now(), time.Now(), time.Now())
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

	gctcmd = viper.GetString("gctcmd")

	gctuser = viper.GetString("gctuser")
        gctpassword = viper.GetString("gctpassword")

        pguser = viper.GetString("pguser")
        pgpassword = viper.GetString("pgpassword")
	pgdb = viper.GetString("pgdb")

	tbtoken = viper.GetString("tbtoken")

	limit_depth = viper.GetInt("limit_depth")

        parm,err := getParms("limit_depth")
        if err == nil {
                limit_depth = int(parm.intp)
        }

	if limit_depth > 90 || limit_depth < 10 {
		limit_depth = 80
	}

	if do_trace {
		fmt.Println("do_trace: ",do_trace)
                fmt.Printf("limit_depth: %d\n",limit_depth)
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
                        o.asset = ord["asset_type"].(string)
                        o.order_side = ord["order_side"].(string) 
                        o.order_type = ord["order_type"].(string)
                        o.creation_time = ord["creation_time"].(float64)
                        o.update_time = ord["update_time"].(float64)
                        o.status = ord["status"].(string)
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

func trend7(pair string) {
        var wr bool = false
	var tm1 int64
	var coeff float64
	var cls float64

        fmt.Printf("Calculate trend7 %v\n",pair)

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
        and "timestamp" > current_timestamp - interval '7 days'
	order by "timestamp";`

        rows, err := db.Query(sqlStatement, pair)
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
                SET trend7 = $1
                WHERE pair = $2;`
                _, err = db.Exec(sqlStatement,coeff,pair)
                if err != nil {
                        fmt.Printf("SQL error: %v\n",err)
                }
        }
}

func trend24(pair string) {
        var wr bool = false
        var tm1 int64
	var coeff float64
	var cls float64

        fmt.Printf("Calculate trend24 %v\n",pair)

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
        and "timestamp" > current_timestamp - interval '24 hours'
        order by "timestamp";`

        rows, err := db.Query(sqlStatement, pair)
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
                SET trend24 = $1
                WHERE pair = $2;`
                _, err = db.Exec(sqlStatement,coeff,pair)
                if err != nil {
                        fmt.Printf("SQL error: %v\n",err)
                }
        }
}

func trend4(pair string) {
	var wr bool = false
        var tm1 int64
	var coeff float64
	var cls float64

        fmt.Printf("Calculate trend4 %v\n",pair)

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
        and "timestamp" > current_timestamp - interval '4 hours'
        order by "timestamp";`

        rows, err := db.Query(sqlStatement, pair)
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
        	SET trend4 = $1
        	WHERE pair = $2;`
        	_, err = db.Exec(sqlStatement,coeff,pair)
        	if err != nil {
                	fmt.Printf("SQL error: %v\n",err)
        	}
	}
}

func calculateTrends() {
        for _, v := range pairs {
		trend7(v)
		trend24(v)
		trend4(v)
	}
}

func processOrders() {
        fmt.Println("Process orders")

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select pair, id  from yourorder
        where status = 'NEW'
        order by "timestamp";`

        rows, err := db.Query(sqlStatement)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
        defer rows.Close()

        for rows.Next(){
                var id string
		var pair string
        	var order map[string]interface{}
		var o Order
                if err := rows.Scan(&pair, &id); err != nil {
                        fmt.Println(err)
                } else {
			fmt.Printf("Pair: %s, OrderID: %s\n",pair, id)
			out :=getOrder(pair, id)
                	err := json.Unmarshal(out, &order)
                	if err != nil { // Handle JSON errors
                        	fmt.Printf("JSON error: %v\n", err)
                        	fmt.Printf("JSON input: %v\n",string(out))
                        	continue
                	}
			o.exchange = strings.ToLower(order["exchange"].(string))
                        o.id = order["id"].(string)
                        o.base_currency = order["base_currency"].(string)
                        o.quote_currency = order["quote_currency"].(string)
                        o.asset = order["asset_type"].(string)
                        o.order_side = order["order_side"].(string) 
                        o.order_type = order["order_type"].(string)
                        o.creation_time = order["creation_time"].(float64)
                        o.update_time = order["update_time"].(float64)
                        o.status = order["status"].(string)
                        o.price = order["price"].(float64)
                        o.amount = order["amount"].(float64)
			if order["cost"] != nil {
                        	o.cost = order["cost"].(float64)
			}
                        fmt.Println(o)

			storeOrder(o.exchange,o.id,o.base_currency+"-"+o.quote_currency,o.asset,o.order_side,o.order_type,o.update_time,o.status,o.price,o.amount)
			
			if o.status == "FILLED" && o.order_side == "BUY" {
				insertPositions(o.exchange,o.base_currency+"-"+o.quote_currency,o.order_side,time.Unix(int64(o.update_time), 0),o.price,o.amount)
                        	submitTelegram("Position: "+o.base_currency+"-"+o.quote_currency+" bought")
			}
                        if o.status == "FILLED" && o.order_side == "SELL" {
            //                    insertPositions(o.exchange,o.base_currency+"-"+o.quote_currency,o.order_side,time.Unix(int64(o.update_time), 0),o.price,o.amount)
                                submitTelegram("Position: "+o.base_currency+"-"+o.quote_currency+" sold")
                        }
		}
        }
}
