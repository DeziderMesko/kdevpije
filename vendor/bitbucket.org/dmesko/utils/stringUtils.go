package utils

import (
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
	"io/ioutil"
)

func ReadFileFromWorkingOrHomeDir(filename, homedir string) ([]byte, error) {
	//try working directory first
	b, err := ioutil.ReadFile(filename)
	if err == nil {
		log.Printf("Data successfuly read from working directory: %v", len(b))
		return b, err
	}
	log.Printf("Unable to read file content from working directory: %v", err)
	// otherwise try home directory
	filePath, err := GetHomeDirConfigFileName(filename, homedir)
	if err != nil {
		log.Printf("I have an issue with homedir detection: %v", err)
		return b, err
	}
	b, err = ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Unable to read file content from home directory: %v", err)
		return nil, err
	}
	
	log.Printf("Data successfuly read from homedir: %v", len(b))
	return b, err
}

func GetHomeDirConfigFileName(filename, homedir string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	cfgDir := filepath.Join(usr.HomeDir, homedir)
	err = os.MkdirAll(cfgDir, 0755)
	return filepath.Join(cfgDir, url.QueryEscape(filename)), err
}

func StripDiacritics(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, _ := transform.String(t, s)
	return strings.ToLower(o)
}

func KeysToLower(m map[string][]string) (ret map[string][]string) {
	ret = make(map[string][]string)
	for k, v := range m {
		ret[strings.ToLower(k)] = v
	}
	return
}

func SliceToLower(slice []string) (ret []string) {
	for _, i := range slice {
		ret = append(ret, strings.ToLower(i))
	}
	return
}

func Distance(e string, q string) int {
	lenE := utf8.RuneCountInString(e)
	lenQ := utf8.RuneCountInString(q)

	matrix := make([][]int, lenQ+1)
	for i := 0; i < lenQ+1; i++ {
		matrix[i] = make([]int, lenE+1)
	}
	for i := 1; i < lenQ+1; i++ {
		matrix[i][0] = i
	}
	for j := 1; j < lenE+1; j++ {
		matrix[0][j] = j
	}

	for i, rs := range StripDiacritics(q) {
		i++
		for j, rq := range StripDiacritics(e) {
			j++
			if rs == rq {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min3(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+1)
			}
		}
	}
	return matrix[lenQ][lenE]
}

func min3(a, b, c int) int {
	return min(min(a, b), c)
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
