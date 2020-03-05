package speed

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"time"
)

var (
	Cmd = &cobra.Command{
		Use:   "speed",
		Short: "speed test between client and server",
		Long:  "this is a speed test tool",
	}

	ServerCmd = &cobra.Command{
		Use:   "server",
		Short: "running test server",
		RunE:  runServer,
	}

	ClientCmd = &cobra.Command{
		Use:   "client",
		Short: "running test client",
		RunE:  runClient,
	}

	argServerListenAddress string

	argClientServerAddress string
	argSeconds             int64

	argUp   bool
	argDown bool
)

func init() {
	ServerCmd.Flags().StringVar(&argServerListenAddress, "addr", ":9014", "server listen address")
	ClientCmd.Flags().StringVar(&argClientServerAddress, "addr", ":9014", "remote server address")
	ClientCmd.Flags().Int64Var(&argSeconds, "t", 60, "request time")
	ClientCmd.Flags().BoolVar(&argUp, "up", true, "test upload speed")
	ClientCmd.Flags().BoolVar(&argDown, "down", true, "test down speed")
	Cmd.AddCommand(ServerCmd, ClientCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	l, err := net.Listen("tcp", argServerListenAddress)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go handleConn(conn)
	}
}

const (
	commandSend = 0x01
	commandRecv = 0x02
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	var b = make([]byte, 1)
	_, err := conn.Read(b)
	if err != nil {
		log.Println("read failed")
		return
	}

	switch b[0] {
	case commandRecv:
		var t = time.Now()
		var totalSize int
		for {
			var maxBuf = make([]byte, 32*32*1024)
			n, err := conn.Read(maxBuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("read failed:", err)
				break
			}
			totalSize += n
		}
		t2 := time.Now()
		speed := float64(totalSize) / float64(t2.Sub(t).Nanoseconds()) / 1024 * 1e9
		if speed > 1024 {
			speed = speed / 1024
			log.Println(fmt.Sprintf("upload speed is : %.2f mb/s, total size: %.2f mb", speed, float64(totalSize)/1024/1024))
		} else {
			log.Println(fmt.Sprintf("upload speed is : %.2f kb/s, total size: %.2f mb", speed, float64(totalSize)/1024/1024))
		}
		return
	case commandSend:
		var t = time.Now()
		var totalSize int
		var i = 1
		for {
			var maxBuf = bytes.Repeat([]byte{'x'}, i*1024)
			n, err := conn.Write(maxBuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("write failed:", err)
				break
			}
			totalSize += n
			i++
		}
		t2 := time.Now()
		speed := float64(totalSize) / float64(t2.Sub(t).Nanoseconds()) / 1024 * 1e9
		if speed > 1024 {
			speed = speed / 1024
			log.Println(fmt.Sprintf("upload speed is: %s", LogSpeed(speed)))
		} else {
			log.Println(fmt.Sprintf("upload speed is : %s", LogSpeed(speed)))
		}
		return
	}

}

func runClient(cmd *cobra.Command, args []string) error {

	var upSeppd float64
	var downSpeed float64

	if argUp {
		// test upload.
		conn, err := net.Dial("tcp", argClientServerAddress)
		if err != nil {
			return err
		}
		_, err = conn.Write([]byte{commandRecv})
		if err != nil {
			return err
		}

		var t = time.Now()
		var totalSize int
		var i = 1
		var seTk = time.NewTicker(time.Second * 1)
		var tick = time.NewTicker(time.Duration(argSeconds) * time.Second)
		for {
			var maxBuf = bytes.Repeat([]byte{'x'}, 1024*i)
			n, err := conn.Write(maxBuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("write failed:", err)
				break
			}
			totalSize += n
			i = i * 2
			select {
			case <-tick.C:
				tick.Stop()
				seTk.Stop()
				conn.Close()
				goto result
			default:
			}

			select {
			case <-seTk.C:
				x := float64(time.Now().Sub(t).Nanoseconds())
				speed := float64(totalSize) / x / 1024 * 1e9
				log.Println(fmt.Sprintf("upload speed is :  %s", LogSpeed(speed)))
			default:
			}
		}
	result:
		t2 := time.Now()
		speed := float64(totalSize) / float64(t2.Sub(t).Nanoseconds()) / 1024 * 1e9 / 1024
		log.Println(fmt.Sprintf("finished -> upload speed is :  %s", LogSpeed(speed)))
		upSeppd = speed
	}

	if argDown {
		log.Printf("start %d seconds download", argSeconds)
		var tick = time.NewTicker(time.Duration(argSeconds) * time.Second)
		// test upload.
		conn, err := net.Dial("tcp", argClientServerAddress)
		if err != nil {
			return err
		}
		_, err = conn.Write([]byte{commandSend})
		if err != nil {
			return err
		}

		var t = time.Now()
		var totalSize int
		var seTk = time.NewTicker(1 * time.Second)
		for {
			var maxBuf = make([]byte, 32*1024)
			n, err := conn.Read(maxBuf)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Println("read failed:", err)
				break
			}
			totalSize += n
			select {
			case <-tick.C:
				tick.Stop()
				seTk.Stop()
				conn.Close()
				goto res2
			default:
			}

			select {
			case <-seTk.C:
				speed := float64(totalSize) / float64(time.Now().Sub(t).Nanoseconds()) / 1024 * 1e9
				log.Println(fmt.Sprintf("download speed is :  %s", LogSpeed(speed)))
			default:
			}
		}
	res2:
		t2 := time.Now()
		speed := float64(totalSize) / float64(t2.Sub(t).Nanoseconds()) / 1024 * 1e9
		downSpeed = speed
	}

	if argUp {
		log.Println(fmt.Sprintf("finished: upload speed is : %f mb/s", upSeppd/1024))
	}
	if argDown {
		log.Println(fmt.Sprintf("finished: download speed is : %f mb/s", downSpeed/1024))
	}
	return nil
}

func LogSpeed(speed float64) string {
	var unit = "kb/s"
	if speed/1024 > 1024 {
		speed = speed / 1024
		unit = "mb/s"
	}
	if speed/1024 > 1024 {
		speed = speed / 1024
		unit = "gb/s"
	}
	return fmt.Sprintf("%.2f %s", speed, unit)
}
