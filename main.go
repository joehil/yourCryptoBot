package main

import (
	"log"
	"os"
	"os/exec"
	"fmt"
	"io"
	"time"
	"strings"
    	"path"
    	"path/filepath"
    	"sync/atomic"
	"github.com/spf13/viper"
	"github.com/natefinch/lumberjack"
	"github.com/secsy/goftp"
	"crypto/tls"
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
        	if a1 == "backup" {
			backup()
			os.Exit(0)
        	}
                if a1 == "list" {
                        list()
			os.Exit(0)
                }
                if a1 == "fetch" {
                        fetch(os.Args[2])
			os.Exit(0)
                }
                if a1 == "decrypt" {
                        decrypt()
                        os.Exit(0)
                }		
		fmt.Println("parameter invalid")
		os.Exit(-1)
	}
	if len(os.Args) == 1 {
		myUsage()
	}
}


func backup() {
t := time.Now()
tstr := t.Format("20060102")
wd := t.Weekday()
fstr := tstr[0:6] + "01"
tunix := t.Unix()
daynum := t.Day()

var ftplogger io.Writer = nil

if do_trace {
	ftplogger = ownlogger
}

config := goftp.Config{
    User:               ftpsuser,
    Password:           ftpspassword,
    ConnectionsPerHost: 10,
    ActiveTransfers: false,
    DisableEPSV: true,
    Timeout:            100 * time.Second,
    Logger:             ftplogger,
    TLSConfig: &tls.Config{
		InsecureSkipVerify: true,
		Renegotiation: 2,
	},
    TLSMode: 0,
}

client, err := goftp.DialConfig(config, ftpshost)
if err != nil {
    panic(err)
}

    Walk(client, dailydir, func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            // no permissions is okay, keep walking
            if err.(goftp.Error).Code() == 550 {
                return nil
            }
            return err
        }

	fstat, err := client.Stat(fullPath)
        fmt.Println(fstat.Name(),fstat.ModTime().Unix())

	if fstat.ModTime().Unix() < tunix - 86400 * dailykeep {
		log.Println("Delete file ",fullPath)
		client.Delete(fullPath)
	}

        return nil
    })

    Walk(client, weeklydir, func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            // no permissions is okay, keep walking
            if err.(goftp.Error).Code() == 550 {
                return nil
            }
            return err
        }

        fstat, err := client.Stat(fullPath)
        fmt.Println(fstat.Name(),fstat.ModTime().Unix())

        if fstat.ModTime().Unix() < tunix - 86400 * weeklykeep {
		log.Println("Delete file ",fullPath)
		client.Delete(fullPath)
        }

        return nil
    })

    Walk(client, monthlydir, func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            // no permissions is okay, keep walking
            if err.(goftp.Error).Code() == 550 {
                return nil
            }
            return err
        }

        fstat, err := client.Stat(fullPath)
        fmt.Println(fstat.Name(),fstat.ModTime().Unix())

        if fstat.ModTime().Unix() < tunix - 86400 * monthlykeep {
		log.Println("Delete file ",fullPath)
		client.Delete(fullPath)
        }

        return nil
    })

client.Close()

if do_encrypt {
	transferfile = "/autowebbackup." + encryptsuffix
	transfersuffix = encryptsuffix
}

// Loop over directories
	for i, s := range dirs {
		var cmd *exec.Cmd
    		fmt.Println(i, s)
                if daynum == 1 {
			cmd = exec.Command(tarcmd, "-czf", tempdir+"/autowebbackup.tar.gz", s)
		} else {
                        cmd = exec.Command(tarcmd, "-cz", "--newer", fstr, "-f", tempdir+"/autowebbackup.tar.gz", s)
		}
		log.Printf(s)
		err := cmd.Run()
		if err != nil {
			log.Printf("Command finished with error: %v", err)
		}
                if do_encrypt {
                        encrypt()
                }

                sparts := strings.SplitAfter(s, "/")
                spart := sparts[len(sparts)-1]


		if daynum == 1 { 
			client, err := goftp.DialConfig(config, ftpshost)
			if err != nil {
 	                       log.Printf("FTPS connect error: %v", err)
			} else {
				log.Println("FTPS connected successfully")
			}
			bigFile, err := os.Open(tempdir+transferfile)
			if err != nil {
                        	log.Printf("Open file error: %v", err)
			}
			err = client.Store(monthlydir+"/"+spart+"-"+tstr+"."+transfersuffix, bigFile)
			if err != nil {
        	                log.Printf("FTPS store error: %v", err)
			} else {
                	        log.Println("FTPS stored successfully")
                	}
                	client.Close()
                	bigFile.Close()
		} else if wd.String() == "Sunday" {
	                tclient, terr := goftp.DialConfig(config, ftpshost)
        	        if terr != nil {
                	        log.Printf("FTPS connect error: %v", terr)
 	   	        } else {
                	        log.Println("FTPS connected successfully")
           	        }
	                tbigFile, terr := os.Open(tempdir+transferfile)
        	        if terr != nil {
                	        log.Printf("Open file error: %v", terr)
                	}
                	terr = tclient.Store(weeklydir+"/"+spart+"-"+tstr+"."+transfersuffix, tbigFile)
        	        if terr != nil {
                	        log.Printf("FTPS store error: %v", terr)
         	        } else {
                	        log.Println("FTPS stored successfully")
			}
			tclient.Close()
			tbigFile.Close()
                } else {
                        tclient, terr := goftp.DialConfig(config, ftpshost)
                        if terr != nil {
                                log.Printf("FTPS connect error: %v", terr)
                        } else {
                                log.Println("FTPS connected successfully")
                        }
                        tbigFile, terr := os.Open(tempdir+transferfile)
                        if terr != nil {
                                log.Printf("Open file error: %v", terr)
                        }
                        terr = tclient.Store(dailydir+"/"+spart+"-"+tstr+"."+transfersuffix, tbigFile)
                        if terr != nil {
                                log.Printf("FTPS store error: %v", terr)
                        } else {
                                log.Println("FTPS stored successfully")
                        }
                        tclient.Close()
                        tbigFile.Close()
                }

		os.Remove(tempdir+transferfile)
	}

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

// Walk a FTP file tree in parallel with prunability and error handling.
// See http://golang.org/pkg/path/filepath/#Walk for interface details.
func Walk(client *goftp.Client, root string, walkFn filepath.WalkFunc) (ret error) {
    dirsToCheck := make(chan string, 100)

    var workCount int32 = 1
    dirsToCheck <- root

    for dir := range dirsToCheck {
        go func(dir string) {
            files, err := client.ReadDir(dir)

            if err != nil {
                if err = walkFn(dir, nil, err); err != nil && err != filepath.SkipDir {
                    ret = err
                    close(dirsToCheck)
                    return
                }
            }

            for _, file := range files {
                if err = walkFn(path.Join(dir, file.Name()), file, nil); err != nil {
                    if file.IsDir() && err == filepath.SkipDir {
                        continue
                    }
                    ret = err
                    close(dirsToCheck)
                    return
                }

                if file.IsDir() {
                    atomic.AddInt32(&workCount, 1)
                    dirsToCheck <- path.Join(dir, file.Name())
                }
            }

            atomic.AddInt32(&workCount, -1)
            if workCount == 0 {
                close(dirsToCheck)
            }
        }(dir)
    }

    return ret
}

func encrypt() {
    fileSrc, err := os.Open(tempdir+"/autowebbackup.tar.gz")
    if err != nil {
        panic(err)
    }
    defer fileSrc.Close()
    fileDst, err := os.Create(tempdir+"/autowebbackup."+encryptsuffix)
    if err != nil {
        panic(err)
    }
    defer fileDst.Close()
    aes, err := NewAes(32, encryptpassw[0:31])
    if err != nil {
        panic(err)
    }
    err = aes.EncryptStream(fileSrc, fileDst)
    if err != nil {
        panic(err)
    }
    os.Remove(tempdir+"/autowebbackup.tar.gz")
    log.Println("File successfully encrypted")
}

func decrypt() {
    fileSrc, err := os.Open(tempdir+"/autowebbackup."+encryptsuffix)
    if err != nil {
        panic(err)
    }
    defer fileSrc.Close()
    fileDst, err := os.Create(tempdir+"/autowebbackup.tar.gz")
    if err != nil {
        panic(err)
    }
    defer fileDst.Close()
    aes, err := NewAes(32, encryptpassw[0:31])
    if err != nil {
        panic(err)
    }
    err = aes.DecryptStream(fileSrc, fileDst)
    if err != nil {
        panic(err)
    }
    fmt.Println("File successfully decrypted")
}

func myUsage() {
     fmt.Printf("Usage: %s argument\n", os.Args[0])
     fmt.Println("Arguments:")
     fmt.Println("backup        Backup the directories mentioned in the config file")
     fmt.Println("list          List all backups")
     fmt.Println("fetch         Fetch backup from server")
     fmt.Println("decrypt       Decrypt backup")
}

func list() {
config := goftp.Config{
    User:               ftpsuser,
    Password:           ftpspassword,
    ConnectionsPerHost: 10,
    ActiveTransfers: false,
    DisableEPSV: true,
    Timeout:            100 * time.Second,
    Logger:             nil,
    TLSConfig: &tls.Config{
		InsecureSkipVerify: true,
		Renegotiation: 2,
	},
    TLSMode: 0,
}

client, err := goftp.DialConfig(config, ftpshost)
if err != nil {
    panic(err)
}

    Walk(client, dailydir, func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            // no permissions is okay, keep walking
            if err.(goftp.Error).Code() == 550 {
                return nil
            }
            return err
        }
        fmt.Println(fullPath)
        return nil
    })

    Walk(client, weeklydir, func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            // no permissions is okay, keep walking
            if err.(goftp.Error).Code() == 550 {
                return nil
            }
            return err
        }
        fmt.Println(fullPath)
        return nil
    })

    Walk(client, monthlydir, func(fullPath string, info os.FileInfo, err error) error {
        if err != nil {
            // no permissions is okay, keep walking
            if err.(goftp.Error).Code() == 550 {
                return nil
            }
            return err
        }
        fmt.Println(fullPath)
        return nil
    })

client.Close()
}

func fetch(filename string) {
config := goftp.Config{
    User:               ftpsuser,
    Password:           ftpspassword,
    ConnectionsPerHost: 10,
    ActiveTransfers: false,
    DisableEPSV: true,
    Timeout:            100 * time.Second,
    Logger:             nil,
    TLSConfig: &tls.Config{
		InsecureSkipVerify: true,
		Renegotiation: 2,
	},
    TLSMode: 0,
}

client, err := goftp.DialConfig(config, ftpshost)
if err != nil {
    panic(err)
}

bigFile, err := os.Create(tempdir+"/autowebbackup."+encryptsuffix)
if err != nil {
    panic(err)
}

err = client.Retrieve(filename, bigFile)
if err != nil {
    fmt.Printf("FTPS retrieve error: %v", err)
} else {
    fmt.Println("FTPS retrieved successfully")
}

client.Close()
}
