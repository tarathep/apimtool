package engine

import (
	"net/url"
	"os"
	"regexp"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
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
			color.New(color.FgYellow).Println("directory " + checkdir + " not found!")
			log.Logger.Warn().Msg("directory " + checkdir + " not found!")
			return false
		}
	}
	return true
}

func urlParse(url_ string) {
	_, err := url.Parse(url_)
	if err != nil {
		color.New(color.FgRed).Println("URL Parsing Error")
		log.Logger.Err(err).Msg("URL Parsing Error")
	}

}
