package engine

import (
	"log"
	"net/url"
	"os"
	"regexp"
)

func getQuotedString(s string) []string {
	ms := regexp.MustCompile(`'(.*?)'`).FindAllStringSubmatch(s, -1)
	ss := make([]string, len(ms))

	for i, m := range ms {
		ss[i] = m[1]
	}

	return ss
}

func checkPaths(paths []string) bool {
	for _, checkdir := range paths {
		if _, err := os.Stat(checkdir); os.IsNotExist(err) {
			log.Println("directory " + checkdir + " not found!")
			return false
		}
	}
	return true
}

func urlParse(url_ string) {
	_, err := url.Parse(url_)
	if err != nil {
		log.Println(err)
	}

}
