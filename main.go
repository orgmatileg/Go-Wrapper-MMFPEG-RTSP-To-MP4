package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

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
func (c *Camera) Stop() error {
	if err := c.CMD.Process.Signal(syscall.SIGTERM); err != nil {
		log.Println("camera: ", err.Error())
	}
	return c.CMD.Wait()
}

// Config ...
type Config struct {
	RetentionFile  float64   `mapstructure:"retention_file"`
	MaxVideoLength string    `mapstructure:"max_video_length"`
	Cameras        []*Camera `mapstructure:"cameras"`
}

var (
	config          Config
	cwd, _          = os.Getwd()
	getTimestamp, _ = regexp.Compile(".*?_(.*?)[.]")
)

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
	checkOrCreateDataDir()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	ticker := time.NewTicker(viper.GetDuration("max_video_length"))
	quitTicker := make(chan struct{})

	startAllCamera()
	go func() {
		for {
			select {
			case <-ticker.C:
				stopAllCamera()
				startAllCamera()
				fileRotation()
				// do stuff
			case <-quitTicker:
				ticker.Stop()
				return
			}
		}
	}()
	<-done
	quitTicker <- struct{}{}
	stopAllCamera()
}

func startAllCamera() {
	for i, c := range config.Cameras {
		t := time.Now()
		filename := fmt.Sprintf("%s/data/%s_%d.mp4", cwd, c.Name, t.Unix())
		config.Cameras[i].CMD = exec.Command("ffmpeg", "-i", c.Server, "-acodec", "aac", "-vcodec", "copy", "-f", "mp4", "-y", filename)
		config.Cameras[i].CMD.Dir = cwd
		if err := c.Start(); err != nil {
			log.Println("camera: " + err.Error())
		}
	}
}

func stopAllCamera() {
	for _, c := range config.Cameras {
		if c.CMD != nil {
			if err := c.Stop(); err != nil {
				log.Println("camera: " + err.Error())
			}
		}
	}
}

func checkOrCreateDataDir() {
	err := os.Mkdir("data", 0755)
	if err != nil && err.Error() != "mkdir data: file exists" {
		panic(err)
	}
}

func fileRotation() {
	var files []string

	root := cwd + "/data"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		ss := strings.Split(file, "_")
		if len(ss) < 2 {
			continue
		}

		ss = strings.Split(ss[1], ".")

		tu, _ := strconv.Atoi(ss[0])
		tf := time.Unix(int64(tu), 0)
		tn := time.Now()
		d := tn.Sub(tf)

		if d.Hours() > config.RetentionFile {
			if err := os.Remove(file); err != nil {
				log.Println("file rotation: " + err.Error())
			}
		}
	}
}
