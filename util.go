package main

import (
	"fmt"
	"log"
	"mime"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"

	"github.com/h2non/filetype"
	"github.com/mozillazg/go-unidecode"
	"github.com/wangii/emoji"
)

var (
	identifiers = make(map[string]int)
	unsafeRegex = regexp.MustCompile(`(?i)[^a-z\d+]`)
)

func strTimestamp() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func getExtension(bytes []byte) string {
	typ, err := filetype.Match(bytes)
	if err != nil {
		return ""
	}

	res := typ.Extension
	if res == "unknown" {
		return ""
	}
	return res
}

func getExtensionByMime(typ string) (string, error) {
	extensions, err := mime.ExtensionsByType(typ)
	if err != nil {
		return "", err
	}

	if len(extensions) == 0 {
		return "", nil
	}

	return extensions[0][1:], nil
}

func getExtensionByMimeOrBytes(mime string, bytes []byte) string {
	if res, err := getExtensionByMime(mime); res != "" && err == nil {
		return res
	}

	return getExtension(bytes)
}

// IrcSafeString converts any emoji unicode characters into emojitag, then
// converts any non-ascii characters into their ascii equivalents, then strips
// characters that satisfy unsafeRegex, and finally disambiguates the
// identifier if required
func IrcSafeString(str string) string {
	emojiTagged := emoji.UnicodeToEmojiTag(str)
	decoded := unidecode.Unidecode(emojiTagged)
	ircSafe := unsafeRegex.ReplaceAllLiteralString(decoded, "")
	return ensureIdentifierIsDistinct(ircSafe)
}

func ensureIdentifierIsDistinct(identity string) string {
	_, exists := identifiers[identity]

	// we've encountered this identifier before so
	// increment the counter and append the new count
	// to the identifier we return
	if exists {
		identifiers[identity]++
		counter := identifiers[identity]
		return fmt.Sprintf("%s%d", identity, counter)
	}

	// it's the first time we're encountering this identifier
	// so we initialise the counter
	identifiers[identity] = 1
	return identity
}

func onInterrupt(fn func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fn()
		os.Exit(1)
	}()
}

func plural(count int, singular, plural string) string {
	if count == 1 || count == -1 {
		return singular
	}

	return plural
}

func logMessage(time time.Time, from, to, message string) {
	timeStr := time.Format("2006-01-02 15:04:05")
	log.Printf("(%s) %s->%s: %s", timeStr, from, to, message)
}
