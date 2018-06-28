/*
 This program for HTTP video streaming.
 This file defines main entry points of the program.

 Copyleft 2016  Viacheslav Dubrovskyi <dubrsl@gmail.com>

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.

*/

/* TODO

1. Конфиг
2. HLS downloader
3. HLS creator
4. Chunk analizer

*/

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	err error
	// Initiate logging system
	log = logrus.New()
	//Debug if true run with debug
	debug bool
	//ConfigFile path to config file
	configFile string
	//path to folder with logs
	logFolder string
)

func main() {
	flag.StringVar(&configFile, "config", "config", "Path to the config file")
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.StringVar(&logFolder, "log", "log", "Folder for output logs (default: log)")
	flag.Parse()

	// Improves perf in 1.1.2 linux/amd64
	runtime.GOMAXPROCS(runtime.NumCPU())

	//Read config
	Viper, err := readConfig(configFile, map[string]interface{}{
		"port":     9090,
		"hostname": "localhost",
		"auth": map[string]string{
			"username": "admin",
			"password": "adminpass",
		},
	})
	if err != nil { // Handle errors reading the config file
		log.Panic("Fatal error in read config file: %s")
	}
	if debug == true {
		Viper.Debug()
	}

	//Initiate log system
	if _, err := os.Stat(logFolder); os.IsNotExist(err) {
		os.Mkdir(logFolder, os.FileMode(0755))
	}
	fileOpenMode := os.O_APPEND
	if _, err := os.Stat(logFolder + "/main.log"); os.IsNotExist(err) {
		fileOpenMode = os.O_CREATE
	}
	file, err := os.OpenFile(logFolder+"/main.log", fileOpenMode|os.O_WRONLY, 0644)
	if err == nil {
		log.Out = file
	} else {
		log.Panic("Failed to log to file")
	}

	log.Info("Run program")

	/* Handle SIGNALS */
	k := make(chan os.Signal, 1)
	signal.Notify(k,
		syscall.SIGINT,  // Terminate
		syscall.SIGTERM, // Terminate
		syscall.SIGQUIT, // Stop gracefully
		syscall.SIGHUP,  // Reload config
		syscall.SIGUSR1, // Reopen log files
		syscall.SIGUSR2, // Seamless binary upgrade
	)

	go func() {
		for {
			sig := <-k
			switch sig {
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				log.Info("Server stopped gracefully.")
				os.Exit(0)
			case syscall.SIGQUIT:
				log.Info("Waiting for running tasks to finish...")
				os.Exit(0)
			case syscall.SIGHUP:
				log.Info("SIGHUP Received: Reloading configuration...")
				err := Viper.ReadInConfig() // Find and read the config file
				if err != nil {             // Handle errors reading the config file
					log.Panic("Fatal error config file: %s")
				}
			case syscall.SIGUSR1:
				log.Info("SIGUSR1 Received: Re-opening logs...")
			case syscall.SIGUSR2:
				log.Info("SIGUSR2 Received: Seamless binary upgrade...")
			}
		}
	}()

	for streamname, stream := range Viper.GetStringMap("stream") {
		fmt.Println(streamname, stream)
	}

	/* Finally start the HTTP server */
	for {
		log.Info("working")
		time.Sleep(5 * time.Second)

	}

	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Server stopped gracefully.")
	}
	os.Exit(0)

}

func readConfig(filename string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath("/etc/dubreamer/")  // path to look for the config file in
	v.AddConfigPath("$HOME/.dubreamer") // call multiple times to add many search paths
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}
