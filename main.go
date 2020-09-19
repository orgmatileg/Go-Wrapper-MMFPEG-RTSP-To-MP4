package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const (
	CameraURL1 = "add camera rtsp"
	CameraURL2 = "add camera rtsp"
)

func main() {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
	}()

	cmd1 := exec.Command("ffmpeg", "-i", CameraURL1, "-acodec", "aac", "-vcodec", "copy", "-f", "mp4", "-y", "camera1.mp4")
	cmd1.Dir = "/home/hakim/project/siswamedia/go-rstp-to-mp4"

	cmd2 := exec.Command("ffmpeg", "-i", CameraURL2, "-acodec", "aac", "-vcodec", "copy", "-f", "mp4", "-y", "camera2.mp4")
	cmd2.Dir = "/home/hakim/project/siswamedia/go-rstp-to-mp4"

	err := cmd1.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	err = cmd2.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	<-done
	cmd1.Wait()
	cmd2.Wait()

}
