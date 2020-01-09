package main

import (
	"github.com/skyduy/mysocks/cipher"
	"github.com/skyduy/mysocks/core"
	"golang.org/x/net/proxy"
	"io"
	"log"
	"math/rand"
	"net"
	"reflect"
	"sync"
	"testing"
	"time"
)

const (
	MaxPackSize = 1024 * 1024 * 5 // 5Mb
	GoogleAddr  = "127.0.0.1:3453"
	LocalAddr   = "127.0.0.1:8448"
	ServerAddr  = "127.0.0.1:8449"
)

var (
	chromeDialer proxy.Dialer
)

func init() {
	log.SetFlags(log.Lshortfile)
	go fakeGoogle()
	go runTunnel()
	var err error
	time.Sleep(time.Second)
	chromeDialer, err = proxy.SOCKS5("tcp", LocalAddr, nil, proxy.Direct)
	if err != nil {
		log.Fatalln(err)
	}
}

// 启动echo server
func fakeGoogle() {
	listener, err := net.Listen("tcp", GoogleAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("echoServer connect Accept")
		go func() {
			defer func() {
				_ = conn.Close()
				log.Println("echoServer connect Close")
			}()
			_, _ = io.Copy(conn, conn)
		}()
	}
}

func runTunnel() {
	password := cipher.RandPassword()
	serverS, err := core.NewLocal(password, LocalAddr, ServerAddr)
	if err != nil {
		log.Fatalln(err)
	}
	localS, err := core.NewServer(password, ServerAddr)
	if err != nil {
		log.Fatalln(err)
	}
	go func() {
		log.Fatalln(serverS.Run())
	}()
	log.Fatalln(localS.Run())
}

// 发生一次连接测试经过代理后的数据传输的正确性
// packSize 代表这个连接发生数据的大小
func testConnect(packSize int) {
	// 随机生产 MaxPackSize byte的[]byte
	data := make([]byte, packSize)
	_, err := rand.Read(data)

	// 连接
	conn, err := chromeDialer.Dial("tcp", GoogleAddr)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// 写
	go func() {
		_, _ = conn.Write(data)
	}()

	// 读
	buf := make([]byte, len(data))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		log.Fatalln(err)
	}
	if !reflect.DeepEqual(data, buf) {
		log.Fatalln("代理传输得到的数据前后不一致")
	} else {
		log.Println("数据传输一致")
	}
}

func TestProxy(t *testing.T) {
	testConnect(rand.Intn(MaxPackSize))
}

// 获取并发发送 data 到 echo server 并且收到全部返回 所花费到时间
func benchmarkProxy(concurrenceCount int) {
	wg := sync.WaitGroup{}
	wg.Add(concurrenceCount)
	for i := 0; i < concurrenceCount; i++ {
		go func() {
			testConnect(rand.Intn(MaxPackSize))
			wg.Done()
		}()
	}
	wg.Wait()
}

// 获取 发送 data 到 echo server 并且收到全部返回 所花费到时间
func BenchmarkProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		benchmarkProxy(10)
		b.StopTimer()
	}
}
