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

var gctcmd string

var gctuser string
var gctpassword string

var pguser string
var pgpassword string
var pgdb string

var tbtoken string

type Parm struct {
	key string
	intp int64
	floatp float64
	stringp string
	datep time.Time
	timep time.Time
	timestampp time.Time
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
			calculateAdvice()
			calculateLimit()
			sendAdvice()
			os.Exit(0)
        	}
                if a1 == "climit" {
                        calculateLimit()
                        os.Exit(0)
                }
                if a1 == "telegram" {
                        sendTelegram()
                        os.Exit(0)
                }
                if a1 == "advice" {
                        sendAdvice()
                        os.Exit(0)
                }
                if a1 == "trend7" {
                        trend7()
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
	(max(close) - min(close)) * 80 / min(close) as potwin
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

func calculateAdvice() {
        fmt.Println("Calculate advice")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
	with subquery as (
	select
	pair,
	case
	when (max - (max - min)/15) < current
	then 'sell'
	when (min + (max - min)/15) > current
	then 'buy'
	else 'no action'
	end as advice
	from yourlimits
	)
        UPDATE yourlimits l
        SET advice = subquery.advice
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
        case
        when (min + (max - min)/10) > current
        then (min + (max - min)/10)
        else (max - (max - min)/10)
        end as limit
        from yourlimits)
        UPDATE yourlimits l
        SET "limit" = subquery.limit
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

func getAdvice(advice string) {
	var wr bool = false

        fmt.Printf("Get advice %v\n",advice)

        f, err := os.OpenFile(pipeFile, os.O_WRONLY|syscall.O_NONBLOCK, 0644)
        if err != nil {
                fmt.Printf("open: %v\n", err)
        }
        defer f.Close()

        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, pguser, pgpassword, pgdb)

        db, err := sql.Open("postgres", psqlconn)
        CheckError(err)

        defer db.Close()

        sqlStatement := `
        select pair, "limit"  from yourlimits
        where advice = $1;`

        rows, err := db.Query(sqlStatement, advice)
        if err != nil {
                fmt.Printf("SQL error: %v\n",err)
        }
	defer rows.Close()

	for rows.Next(){
		var pair string
		var limit float64
		if err := rows.Scan(&pair, &limit); err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%v - %f\n", pair, limit)
		_, err := f.WriteString(fmt.Sprintf("%v %v at %f|",advice,pair,limit))
		wr = true
		if err != nil {
			fmt.Println(err)
		}
	}
	if err := rows.Err(); err != nil {
    		fmt.Println(err)
	}
	if wr {
                f.WriteString("\n")
	}
}

func sendAdvice() {
	_, err := getParms("Nobuyinfo")
	if err != nil {
		getAdvice("buy")
	}
        _, err = getParms("Nosellinfo")
        if err != nil {
		getAdvice("sell")
	}
//	getAdvice("no action")
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

        do_trace = viper.GetBool("do_trace")

	gctcmd = viper.GetString("gctcmd")

	gctuser = viper.GetString("gctuser")
        gctpassword = viper.GetString("gctpassword")

        pguser = viper.GetString("pguser")
        pgpassword = viper.GetString("pgpassword")
	pgdb = viper.GetString("pgdb")

	tbtoken = viper.GetString("tbtoken")

	if do_trace {
		fmt.Println("do_trace: ",do_trace)
		for i, v := range pairs {
			fmt.Printf("Index: %d, Value: %v\n", i, v )
		}
	}
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("cron        Do regular work")
     fmt.Println("climit      Calculate new limits")
     fmt.Println("telegram    Start telegram daemon")
}

func CheckError(err error) {
    if err != nil {
        panic(err)
    }
}

func trend7() {
	r := new(regression.Regression)
	r.SetObserved("Close")
	r.SetVar(0, "Timestamp")
	r.Train(
		regression.DataPoint(10, []float64{1}),
	)
	r.Train(
		regression.DataPoint(11, []float64{2}),
		regression.DataPoint(12, []float64{3}),
		regression.DataPoint(13, []float64{4}),
		regression.DataPoint(14, []float64{5}),
		regression.DataPoint(17, []float64{6}),
	)
	r.Train(
		regression.DataPoint(18, []float64{7}),
		regression.DataPoint(20, []float64{8}),
		regression.DataPoint(22, []float64{9}),
		regression.DataPoint(23, []float64{10}),
	)
	r.Run()

	fmt.Printf("Regression formula:\n%v\n", r.Formula)
	fmt.Printf("Coeff: %f\n", r.Coeff(0))
        fmt.Printf("Coeff: %f\n", r.Coeff(1))
}
