package funcs

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var (
	numberOpenedFiles = 10
	rowDataChan       = make(chan []*ArchiveTable, 100)
	// DefaultLocalTimeZone string
	fileChan = make(chan *os.File, numberOpenedFiles)
	// concurrentInsertNum  int
	// writeTmpFileDone     = make(chan struct{}, 1)
	isComplete = 0
)

func NotifyJobComplete() {
	isComplete = 1
}

func ConsumeRowDataToFile() {
	InitTmpFiles()
	ctx, cancel := context.WithCancel(context.Background())
	var checkWriteRowStatus func()
	checkWriteRowStatus = func() {
		time.Sleep(50 * time.Second)
		for {
			log.Printf("rowDataChan channel list length --> %d \n", len(rowDataChan))
			if isComplete == 1 {
				time.Sleep(10 * time.Second)
				log.Printf("write data to tmp file done \n")
				isComplete = 0
				break
			}
			time.Sleep(10 * time.Second)
		}
		cancel()
	}
	go checkWriteRowStatus()

	for i := 0; i < numberOpenedFiles; i++ {
		go func(ctx context.Context) {
			// rowString := ""
			file := <-fileChan
			defer file.Close()
			bufferedWriter := bufio.NewWriter(file)
			for {
				select {
				case rows := <-rowDataChan:
					for _, v := range rows {
						rowString := fmt.Sprintf("%s\001%s\001%s\001%d\001%.3f\001%.3f\001%.3f\001%.3f\n", v.Endpoint, v.Metric, v.Tags, v.UnixTime, v.Avg, v.Min, v.Max, v.P95)
						_, err := bufferedWriter.WriteString(rowString)
						if err != nil {
							log.Fatal(err)
						}
					}
					bufferedWriter.Flush()
				case <-ctx.Done():
					break
				}
			}
		}(ctx)
	}
}

func InitTmpFiles() {
	tmp_dir := Config().Hdfs.LocalTempDir
	err := os.MkdirAll(tmp_dir+"/tmp", 0750)
	if err != nil {
		log.Fatal(err)
		return
	}
	maxOpenFile := Config().Hdfs.MaxFileOpen
	openNum, err := strconv.Atoi(maxOpenFile)
	if err != nil {
		log.Fatal(err)
	} else {
		numberOpenedFiles = openNum
	}
	log.Printf("max open file %d", numberOpenedFiles)
	fileChan = make(chan *os.File, numberOpenedFiles)
	for i := 0; i < numberOpenedFiles; i++ {
		t, _ := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+DefaultLocalTimeZone, time.Now().Format("2006-01-02")))
		fileName := fmt.Sprintf("archive_%s_%d", time.Unix(t.Unix(), 0).Format("2006_01_02"), i)
		log.Printf("open local file %s", fileName)
		file, err := os.OpenFile(tmp_dir+"/tmp/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return
		}
		fileChan <- file
	}
}

func getLocalIP() string {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("net.Interfaces failed, err:%v", err.Error())
	}
	localIP := ""
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						localIP = ipnet.IP.String()
						break
					}
				}
			}
		}

	}
	return localIP
}

func MergeTmpFile() string {
	tmp_dir := Config().Hdfs.LocalTempDir
	// tmp files
	// WriteDataToFile()

	// merge tmp files
	t, _ := time.Parse("2006-01-02 15:04:05 MST", fmt.Sprintf("%s 00:00:00 "+DefaultLocalTimeZone, time.Now().Format("2006-01-02")))
	ip := getLocalIP()
	fileName := fmt.Sprintf("archive_%s_%s", time.Unix(t.Unix(), 0).Format("2006_01_02"), ip)

	mergeFile, err := os.OpenFile(tmp_dir+"/"+fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		log.Fatal(err)
		return fileName
	}
	tmpFileList, err := filepath.Glob(tmp_dir + "/tmp/archive_*")
	if err != nil {
		log.Fatal(err)
		return fileName
	}
	for _, v := range tmpFileList {
		file, err := os.OpenFile(v, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Fatal(err)
			return fileName
		}
		// b, err := ioutil.ReadAll(file)
		// if err != nil {
		// 	log.Fatal(err)
		// 	return
		// }
		// mergeFile.Write(b)
		_, err = io.Copy(mergeFile, file)
		if err != nil {
			log.Fatal(err)
			return fileName
		}
		file.Close()
	}
	mergeFile.Close()
	cmd := exec.Command("rm", "-rf", tmp_dir+"/tmp")
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		return fileName
	}

	return fileName
}
