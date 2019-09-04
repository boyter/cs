package snippet

import (
	"io/ioutil"
	"math/rand"
	"strings"
	"testing"
)

func TestExtractRelevant(t *testing.T) {
	relevant := ExtractRelevant("this is some text (╯°□°）╯︵ ┻━┻) the thing we want is here", []LocationType{
		{
			Term:     "the",
			Location: 31,
		},
	}, 30, 20, "...")

	if len(relevant) == 0 {
		t.Error("Expected some value")
	}
}

func TestExtractLocation(t *testing.T) {
	content, _ := ioutil.ReadFile("blns.json")

	for i := 0; i < 10000; i++ {
		location := ExtractLocation(RandStringBytes(rand.Intn(2)), string(content), 50)

		for l := range location {
			if l > len([]rune(string(content))) {
				t.Error("Should not be longer")
			}
		}
	}
}

func TestExtractLocations(t *testing.T) {
	content, _ := ioutil.ReadFile("blns.json")

	locations := ExtractLocations([]string{"test", "something", "other"}, string(content))

	if len(locations) == 0 {
		t.Error("Expected at least one location")
	}
}

func TestGetPrevCount(t *testing.T) {
	got := GetPrevCount(300)

	if got != 50 {
		t.Error("Expected 50 got", got)
	}
}

// Designed to catch out any issues with unicode and the like
func TestFuzzy(t *testing.T) {
	content, _ := ioutil.ReadFile("blns.json")

	split := strings.Split("a b c d e f g h i j k l m n o p q r s t u v w x y z", " ")

	for i, t := range split {
		ExtractRelevant(string(content), []LocationType{
			{
				Term:     t,
				Location: i,
			},
		}, 300, 50, "...")
	}

	for i := 0; i < 10000; i++ {
		ExtractRelevant(string(content), []LocationType{
			{
				Term:     RandStringBytes(rand.Intn(10)),
				Location: rand.Intn(len(content)),
			},
		}, 300, 50, "...")
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ~!@#$%^&*()_+`1234567890-=[]\\{}|;':\",./<>?"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
