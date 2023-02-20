package util

import (
	"hash/fnv"
	"math/rand"
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"
)

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func URLtoLatLng(url string) map[string]string {
	re := regexp.MustCompile(`!3d(-?\d+.\d+)`)
	lat := re.FindStringSubmatch(url)
	if len(lat) != 2 {
		return nil
	}

	re = regexp.MustCompile(`!4d(-?\d+.\d+)`)
	lng := re.FindStringSubmatch(url)
	if len(lng) != 2 {
		return nil
	}

	return map[string]string{
		"lat": lat[1],
		"lng": lng[1],
	}
}

// Gathering the long URL from the short google maps url
func GatherLongUrlFromShortUrl(shortURL string) (string, error) {

	client := &http.Client{}
	res, err := client.Get(shortURL)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	defer res.Body.Close()
	fullURL := res.Request.URL.String()
	for res.StatusCode == 302 || res.StatusCode == 301 {
		res, err = client.Get(fullURL)
		if err != nil {
			logrus.Error(err)
			return "", err
		}
		defer res.Body.Close()
		fullURL = res.Request.URL.String()

	}
	return fullURL, nil

}
