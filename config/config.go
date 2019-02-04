package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"whapp-irc/maps"
)

// Config contains all the possible configuration options and their values
type Config struct {
	Hostname string

	UpstreamIRC        bool
	UpstreamIRCHTTPS   bool
	UpstreamIRCMethod  string
	UpstreamIRCPort    string
	UpstreamIRCHost    string
	UpstreamIRCPath    string
	UpstreamIRCBaseURI string

	FileServerHost    string
	FileServerPort    string
	FileServerHTTPS   bool
	FileServerBaseURI string

	IRCPort string

	MapProvider maps.Provider

	AlternativeReplay bool

	IRCChannels     []string
	IRCNickname     string
	IRCIdentityHash string
	IRCIdentityID   string
}

func getEnvDefault(env, def string) string {
	res := os.Getenv(env)
	if res == "" {
		return def
	}
	return res
}

func externalIP() (string, error) {
	noIfacesErr := errors.New("are you connected to the network?")

	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {

		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			// interface is down or is loopback interface
			continue
		}

		addrs, err := iface.Addrs()

		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", noIfacesErr
}

func defaultHostAddress() string {
	var hostname string

	// try using the OS hostname if available
	hostname, err := os.Hostname()

	if err != nil {

		// otherwise, try use a non-loopback ethernet interface ip address
		ip, err := externalIP()

		if err != nil {
			// otherwise, just use localhost
			hostname = "localhost"
		} else {
			hostname = ip
		}
	}

	return hostname
}

// ReadEnvVars reads environment variables and returns a Config instance
// containing the parsed values, or an error.
func ReadEnvVars() (Config, error) {

	var hostname string
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	host := getEnvDefault("HOST", defaultHostAddress())

	upstreamIrc := getEnvDefault("UPSTREAM_IRC", "false")
	upstreamIrcHttps := getEnvDefault("UPSTREAM_IRC_HTTPS", "false")
	upstreamIrcMethod := getEnvDefault("UPSTREAM_IRC_METHOD", "POST")
	upstreamIrcPath := getEnvDefault("UPSTREAM_IRC_PATH", "/")
	upstreamIrcHost := getEnvDefault("UPSTREAM_IRC_HOST", "localhost")
	upstreamIrcPort := getEnvDefault("UPSTREAM_IRC_PORT", "80")

	ui, err := strconv.ParseBool(upstreamIrc)

	// FIXME
	// crash parent process on error

	if err != nil {
		return Config{}, err
	}

	var uiUseHttps string
	uiHttps, err := strconv.ParseBool(upstreamIrcHttps)
	if uiHttps {
		uiUseHttps = "s"
	}

	// FIXME
	// crash parent process on error

	if err != nil {
		return Config{}, err
	}

	upstreamIrcBaseUri := fmt.Sprintf("http%s://%s:%s", uiUseHttps, upstreamIrcHost, upstreamIrcPort)

	channels := getEnvDefault("UPSTREAM_IRC_IDENTITY_CHANNELS", "")
	nickname := getEnvDefault("UPSTREAM_IRC_IDENTITY_NICKNAME", "")
	hash := getEnvDefault("UPSTREAM_IRC_IDENTITY_HASH", "")
	id := getEnvDefault("UPSTREAM_IRC_IDENTITY_ID", "")

	missingUpstreamIdentityCfg := nickname == "" || hash == "" || id == ""

	if ui && missingUpstreamIdentityCfg {
		return Config{}, errors.New("missing required configuration options for irc identity")
	}

	fileServerPort := getEnvDefault("FILE_SERVER_PORT", "3000")
	fileServerUseHTTPS := getEnvDefault("FILE_SERVER_HTTPS", "false")
	ircPort := getEnvDefault("IRC_SERVER_PORT", "6060")
	mapProviderRaw := getEnvDefault("MAP_PROVIDER", "google-maps")
	replayMode := getEnvDefault("REPLAY_MODE", "normal")

	fsUseHTTPS, err := strconv.ParseBool(fileServerUseHTTPS)
	if err != nil {
		return Config{}, err
	}
	fileServerBaseURI := fmt.Sprintf("http%s://%s:%s", fileServerUseHTTPS, host, fileServerPort)

	var mapProvider maps.Provider
	switch strings.ToLower(mapProviderRaw) {
	case "openstreetmap", "open-street-map":
		mapProvider = maps.OpenStreetMap
	case "googlemaps", "google-maps":
		mapProvider = maps.GoogleMaps

	default:
		err := fmt.Errorf("no map provider %s found", mapProviderRaw)
		return Config{}, err
	}

	ircChannels := strings.Fields(channels)
	ircNickname := nickname
	ircIdentityHash := hash
	ircIdentityId := id

	return Config{
		Hostname: hostname,

		UpstreamIRC:        ui,
		UpstreamIRCHTTPS:   uiHttps,
		UpstreamIRCMethod:  upstreamIrcMethod,
		UpstreamIRCPort:    upstreamIrcPort,
		UpstreamIRCHost:    upstreamIrcHost,
		UpstreamIRCPath:    upstreamIrcPath,
		UpstreamIRCBaseURI: upstreamIrcBaseUri,

		FileServerHost:    host,
		FileServerPort:    fileServerPort,
		FileServerHTTPS:   fsUseHTTPS,
		FileServerBaseURI: fileServerBaseURI,

		IRCPort: ircPort,

		MapProvider: mapProvider,

		AlternativeReplay: replayMode == "alternative",

		IRCChannels:     ircChannels,
		IRCNickname:     ircNickname,
		IRCIdentityHash: ircIdentityHash,
		IRCIdentityID:   ircIdentityId,
	}, nil
}
