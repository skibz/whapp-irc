package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"mime"
	"regexp"
	"strconv"
	"time"

	"github.com/h2non/filetype"
	"github.com/mozillazg/go-unidecode"
	"github.com/wangii/emoji"
)

var (
	unsafeRegex = regexp.MustCompile(`(?i)[^a-z\d+]`)
	identifiers = make(map[string]int)
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

func IrcSafeString(str string) string {
	emojiTagged := emoji.UnicodeToEmojiTag(str)
	decoded := unidecode.Unidecode(emojiTagged)
	ircSafe := unsafeRegex.ReplaceAllLiteralString(decoded, "")

	if ircSafe == "" {
		return ensureIdentifierIsDistinct("x" + hex.EncodeToString([]byte(str)))
	}
	return ensureIdentifierIsDistinct(ircSafe)
}

func ensureIdentifierIsDistinct(identity string) string {
	if _, exists := identifiers[identity]; exists {
		identifiers[identity]++
		counter := identifiers[identity]
		return fmt.Sprintf("%s_%d", identity, counter)
	}
	identifiers[identity] = 1
	return identity
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
