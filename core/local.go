package core

import (
	"github.com/skyduy/mysocks/cipher"
	"log"
	"net"
)

type MyLocal struct {
	Cipher     *cipher.Cipher
	ListenAddr *net.TCPAddr
	RemoteAddr *net.TCPAddr
}

// 新建一个本地端
// 本地端的职责是:
// 1. 监听来自本机浏览器的代理请求
// 2. 转发前加密数据
// 3. 转发socket数据到墙外代理服务端
// 4. 把服务端返回的数据转发给用户的浏览器
func NewLocal(password string, listenAddr, remoteAddr string) (*MyLocal, error) {
	bsPassword, err := cipher.ParsePassword(password)
	if err != nil {
		return nil, err
	}
	structListenAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	structRemoteAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}
	return &MyLocal{
		Cipher:     cipher.NewCipher(bsPassword),
		ListenAddr: structListenAddr,
		RemoteAddr: structRemoteAddr,
	}, nil
}

// 本地端启动监听，接收来自本机浏览器的连接，并进行本次连接处理
func (local *MyLocal) Run() error {
	return ListenSecureTCP(local.ListenAddr, local.Cipher, local.handleConn)
}

func (local *MyLocal) handleConn(userConn *SecureTCPConn) {
	defer userConn.Close()

	proxyServer, err := DialTCPSecure(local.RemoteAddr, local.Cipher)
	if err != nil {
		log.Println(err)
		return
	}
	defer proxyServer.Close()

	// 直接进行转发，socks5协议的local部分，由chrome插件switch Omega控制
	// 从 proxyServer 读取数据发送到 localUser
	go func() {
		err := proxyServer.DecodeCopy(userConn)
		if err != nil {
			// 在 copy 的过程中可能会存在网络超时等 error 被 return，只要有一个发生了错误就退出本次工作
			userConn.Close()
			proxyServer.Close()
		}
	}()
	// 从 localUser 发送数据发送到 proxyServer，这里因为处在翻墙阶段出现网络错误的概率更大
	_ = userConn.EncodeCopy(proxyServer)
}
