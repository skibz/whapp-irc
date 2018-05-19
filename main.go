package main

import (
	"log"
	"net"
	"os"
	"whapp-irc/database"

	"github.com/chromedp/chromedp"
)

const (
	defaultHost           = "localhost"
	defaultFileServerPort = "3000"
	defaultIRCPort        = "6060"
)

var fs *FileServer
var userDb *database.Database
var pool *chromedp.Pool

func handleSocket(socket *net.TCPConn) {
	conn, err := MakeConnection()
	if err != nil {
		log.Printf("error while making connection: %s", err)
	}
	go conn.BindSocket(socket)
}

func loadEnvironmentVariables() (host, fileServerPort, ircPort string) {
	host = os.Getenv("HOST")
	if host == "" {
		host = defaultHost
	}

	fileServerPort = os.Getenv("FILE_SERVER_PORT")
	if fileServerPort == "" {
		fileServerPort = defaultFileServerPort
	}

	ircPort = os.Getenv("IRC_SERVER_PORT")
	if ircPort == "" {
		ircPort = defaultIRCPort
	}

	return
}

func startFileServer(host, fileServerPort string) (*FileServer, error) {
	fs, err := MakeFileServer(host, fileServerPort, "files")
	if err != nil {
		return nil, err
	}

	go fs.Start()
	onInterrupt(func() { fs.Stop() })

	return fs, nil
}

func createPool() (*chromedp.Pool, error) {
	pool, err := chromedp.NewPool()
	if err != nil {
		return nil, err
	}
	onInterrupt(func() { pool.Shutdown() })

	return pool, nil
}

func main() {
	host, fileServerPort, ircPort := loadEnvironmentVariables()

	var err error

	userDb, err = database.MakeDatabase("db/users")
	if err != nil {
		panic(err)
	}

	fs, err = startFileServer(host, fileServerPort)
	if err != nil {
		panic(err)
	}

	pool, err = createPool()
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", ":"+ircPort)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	for {
		socket, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("error accepting TCP connection: %#v", err)
			continue
		}

		go handleSocket(socket)
	}
}
