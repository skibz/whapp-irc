package main

import (
	"encoding/hex"
	"github.com/wangii/emoji"
	"strings"
	"testing"
)

const unsafeCharacters = "$&!:;/?^%#*~`"

func TestIrcSafeStringSimpleNoEmoji(t *testing.T) {
	simpleNoEmojiStr := IrcSafeString(unsafeCharacters)
	if simpleNoEmojiStr == "" {
		t.Fatalf("expected simple no emoji string to not be empty string after invoking IrcSafeString but found empty string: %s", simpleNoEmojiStr)
	}
	t.Logf("simple no emoji string is unique after invoking IrcSafeString: %s", simpleNoEmojiStr)
}

func TestIrcSafeStringSimpleNoEmojiDecode(t *testing.T) {
	simpleNoEmojiStr := IrcSafeString(unsafeCharacters)
	if simpleNoEmojiStr == "" {
		t.Fatalf("expected simple no emoji string to not be empty string after invoking IrcSafeString but found empty string: %s", simpleNoEmojiStr)
	}
	parts := strings.Split(simpleNoEmojiStr, "_")
	stripped := parts[0][1:]
	original, err := hex.DecodeString(stripped)
	if err != nil {
		t.Fatalf("error decoding hex string: %e %s %s", err, stripped, original)
	}
	if string(original) != unsafeCharacters {
		t.Fatalf("expected strings to match but found no match: %s %s", stripped, original)
	}
	t.Logf("strings match after decoding back to original string value: %s %s", original, unsafeCharacters)
}

func TestIrcSafeStringSimple(t *testing.T) {
	// a basic emoji string
	simpleEmojiStr := emoji.EmojiTagToUnicode(":ok_hand:" + unsafeCharacters)

	// we invoke the test subject a few times in a
	// row to test the map[string]int side effects
	simpleSafeStr1 := IrcSafeString(simpleEmojiStr)
	simpleSafeStr2 := IrcSafeString(simpleEmojiStr)
	simpleSafeStr3 := IrcSafeString(simpleEmojiStr)

	if simpleSafeStr1 == "" ||
		simpleSafeStr2 == "" ||
		simpleSafeStr3 == "" {
		t.Fatalf("expected simple emoji strings to not be empty strings after invoking IrcSafeString but found empty strings: %s %s %s", simpleSafeStr1, simpleSafeStr2, simpleSafeStr3)
	}

	// ensure none of the strings match
	if simpleSafeStr1 == simpleSafeStr2 ||
		simpleSafeStr1 == simpleSafeStr3 ||
		simpleSafeStr2 == simpleSafeStr3 {
		t.Fatalf("expected simple emoji strings to NOT match after invoking IrcSafeString but found matching strings: %s %s %s", simpleSafeStr1, simpleSafeStr2, simpleSafeStr3)
	}

	// ensure none of them contain the characters we expected to not exist
	if strings.Contains(simpleSafeStr1, unsafeCharacters) ||
		strings.Contains(simpleSafeStr2, unsafeCharacters) ||
		strings.Contains(simpleSafeStr3, unsafeCharacters) {
		t.Fatalf("expected simple emoji strings to NOT contain unsafeCharacters substring after invoking IrcSafeString but found substring match: %s %s %s", simpleSafeStr1, simpleSafeStr2, simpleSafeStr3)
	}

	// ensure the duplicate strings have been made distinct with an incrementing integer
	if !strings.HasSuffix(simpleSafeStr2, "2") {
		t.Fatalf("expected duplicate simple emoji string to end with a disambiguating numeric suffix: %s", simpleSafeStr2)
	}

	// ensure the duplicate strings have been made distinct with an incrementing integer
	if !strings.HasSuffix(simpleSafeStr3, "3") {
		t.Fatalf("expected duplicate simple emoji string to end with a disambiguating numeric suffix: %s", simpleSafeStr3)
	}

	t.Logf("simple emoji strings are unique after invoking IrcSafeString: %s %s %s", simpleSafeStr1, simpleSafeStr2, simpleSafeStr3)
}

func TestIrcSafeStringComplex(t *testing.T) {
	// a much longer one
	complexEmojiStr := emoji.EmojiTagToUnicode(":ok_hand::ok_hand::ok_hand::ok_hand::ok_hand::ok_hand::ok_hand:" + unsafeCharacters)

	// we invoke the test subject a few times in a
	// row to test the map[string]int side effects
	complexSafeStr1 := IrcSafeString(complexEmojiStr)
	complexSafeStr2 := IrcSafeString(complexEmojiStr)
	complexSafeStr3 := IrcSafeString(complexEmojiStr)

	if complexSafeStr1 == "" ||
		complexSafeStr2 == "" ||
		complexSafeStr3 == "" {
		t.Fatalf("expected complex emoji strings to not be empty strings after invoking IrcSafeString but found empty strings: %s %s %s", complexSafeStr1, complexSafeStr2, complexSafeStr3)
	}

	// ensure none of the strings match
	if complexSafeStr1 == complexSafeStr2 ||
		complexSafeStr1 == complexSafeStr3 ||
		complexSafeStr2 == complexSafeStr3 {
		t.Fatalf("expected complex emoji strings to NOT match after invoking IrcSafeString but found matching strings: %s %s %s", complexSafeStr1, complexSafeStr2, complexSafeStr3)
	}

	// ensure none of them contain the characters we expected to not exist
	if strings.Contains(complexSafeStr1, unsafeCharacters) ||
		strings.Contains(complexSafeStr2, unsafeCharacters) ||
		strings.Contains(complexSafeStr3, unsafeCharacters) {
		t.Fatalf("expected complex emoji strings to NOT contain unsafeCharacters substring after invoking IrcSafeString but found substring match: %s %s %s", complexSafeStr1, complexSafeStr2, complexSafeStr3)
	}

	// ensure the duplicate strings have been made distinct with an incrementing integer
	if !strings.HasSuffix(complexSafeStr2, "2") {
		t.Fatalf("expected duplicate complex emoji string to end with a disambiguating numeric suffix: %s", complexSafeStr2)
	}

	// ensure the duplicate strings have been made distinct with an incrementing integer
	if !strings.HasSuffix(complexSafeStr3, "3") {
		t.Fatalf("expected duplicate complex emoji string to end with a disambiguating numeric suffix: %s", complexSafeStr3)
	}

	t.Logf("complex emoji strings are unique after invoking IrcSafeString: %s %s %s", complexSafeStr1, complexSafeStr2, complexSafeStr3)
}
