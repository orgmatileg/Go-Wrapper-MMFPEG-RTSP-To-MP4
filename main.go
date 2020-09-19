package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
)

// Camera ...
type Camera struct {
	Name   string `mapstructure:"name"`
	Server string `mapstructure:"server"`
	CMD    *exec.Cmd
}

// Start ...
func (c *Camera) Start() error { return c.CMD.Start() }

// Stop ...
func (c *Camera) Stop() error { return c.CMD.Wait() }

// Config ...
type Config struct {
	Cameras []*Camera `mapstructure:"cameras"`
}

var config Config

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("viper: %s", err.Error()))
	}

	if err := viper.Unmarshal(&config); err != nil {
		panic(fmt.Errorf("viper: %s", err.Error()))
	}
}

func main() {

	initConfig()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	for i, c := range config.Cameras {
		filename := fmt.Sprintf("%s.mp4", c.Name)
		config.Cameras[i].CMD = exec.Command("ffmpeg", "-i", c.Server, "-acodec", "aac", "-vcodec", "copy", "-f", "mp4", "-y", filename)
		cwd, _ := os.Getwd()
		config.Cameras[i].CMD.Dir = cwd

		if err := c.Start(); err != nil {
			log.Println("camera: " + err.Error())
		}
	}

	<-done
	for _, c := range config.Cameras {
		if err := c.Stop(); err != nil {
			log.Println("camera: " + err.Error())
		}
	}
}
