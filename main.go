package main

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"whapp-irc/config"
	"whapp-irc/database"
	"whapp-irc/files"
	"whapp-irc/maps"

	"github.com/chromedp/chromedp"
)

var (
	cfg    *config.Config
	fs     *files.FileServer
	userDb *database.Database
	pool   *chromedp.Pool

	listener *net.TCPListener

	mapProvider       maps.Provider
	alternativeReplay bool

	startTime = time.Now()
	commit    string

	upstreamDriver = false
)

func triggerUpstreamIrcConnect() {
	errfmt := "upstream irc server connect error: %s"
	for {
		err := func() error {

			uri, err := url.ParseRequestURI(cfg.UpstreamIRCBaseURI)
			if err != nil {
				return err
			}

			uri.Path = cfg.UpstreamIRCPath
			body := url.Values{}

			for _, channel := range cfg.IRCChannels {
				body.Add("chan", channel)
			}

			body.Add("host", cfg.Hostname)
			body.Add("nick", cfg.IRCNickname)
			body.Add("id", cfg.IRCIdentityID)
			body.Add("hash", cfg.IRCIdentityHash)

			reader := strings.NewReader(body.Encode())
			req, err := http.NewRequest(cfg.UpstreamIRCMethod, uri.String(), reader)
			if err != nil {
				return err
			}

			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			c := &http.Client{}
			res, err := c.Do(req)
			if err != nil {
				return err
			}

			defer res.Body.Close()

			if res.StatusCode >= 400 {
				return errors.New("upstream sent " + res.Status)
			}

			return nil
		}()

		// exit the loop if no error
		if err == nil {
			break
		}

		// otherwise log the error and keep iterating
		log.Printf(errfmt, err)

		time.Sleep(2 * time.Second)
	}
	log.Println("LEAVE triggerUpstreamIrcConnect")
}

func init() {
	var err error

	config, err := config.ReadEnvVars()
	if err != nil {
		panic(err)
	}
	cfg = &config

	mapProvider = cfg.MapProvider
	alternativeReplay = cfg.AlternativeReplay

	userDb, err = database.MakeDatabase("db/users")
	if err != nil {
		panic(err)
	}

	fs, err = files.MakeFileServer(
		cfg.FileServerHost,
		cfg.FileServerPort,
		"files",
		cfg.FileServerHTTPS,
	)
	if err != nil {
		panic(err)
	}
	pool, err = chromedp.NewPool()
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", ":"+cfg.IRCPort)
	if err != nil {
		panic(err)
	}

	listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
}

func main() {

	// boot the file server
	go func() {

		// FIXME
		// crash parent on error

		if err := fs.Start(); err != nil {
			log.Printf("error while starting fileserver: %s", err)
		}
	}()

	// ensure the file server shuts down
	// ensure the chrome pool shuts down
	// ensure the tcp server shuts down
	// FIXME
	// escalate handle any errors (not that they really matter)
	defer func() {
		listener.Close()
		fs.Stop()
		pool.Shutdown()
	}()

	if cfg.UpstreamIRC {
		go triggerUpstreamIrcConnect()
	}

	for {
		if upstreamDriver {
			break
		}
		// try and init a new connection
		socket, err := listener.AcceptTCP()

		// FIXME
		// crash parent on error, rather?
		if err != nil {
			log.Printf("error accepting TCP connection: %s", err)
			continue
		}

		// and asynchronously do irc
		// FIXME
		// crash parent on error, rather?
		go func() {
			if err := BindSocket(socket, cfg); err != nil {
				log.Println(err)
			}
		}()

		// update the exit condition variable
		upstreamDriver = cfg.UpstreamIRC && true
	}

	if cfg.UpstreamIRC {
		log.Println("about to close tcp server")
		// stop accepting new connections
		// now that upstream has connected
		listener.Close()

		log.Println("about to block until panic on main thread")
		for {
			// and wait until something crashes
		}
	}
}
