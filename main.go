package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// OmdbAPI stores API results 
type OmdbAPI struct {
	Actors     string `json:"Actors"`
	Awards     string `json:"Awards"`
	BoxOffice  string `json:"BoxOffice"`
	Country    string `json:"Country"`
	DVD        string `json:"DVD"`
	Director   string `json:"Director"`
	Genre      string `json:"Genre"`
	Language   string `json:"Language"`
	Metascore  string `json:"Metascore"`
	Plot       string `json:"Plot"`
	Poster     string `json:"Poster"`
	Production string `json:"Production"`
	Rated      string `json:"Rated"`
	Ratings    []struct {
		Source string `json:"Source"`
		Value  string `json:"Value"`
	} `json:"Ratings"`
	Released   string `json:"Released"`
	Error      string `json:"Error"`
	Response   string `json:"Response"`
	Runtime    string `json:"Runtime"`
	Title      string `json:"Title"`
	Type       string `json:"Type"`
	Website    string `json:"Website"`
	Writer     string `json:"Writer"`
	Year       string `json:"Year"`
	ImdbID     string `json:"imdbID"`
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
}

func queryAPI(s string) *OmdbAPI {
	// var data []byte
	jsonData := &OmdbAPI{}
	var URL *url.URL

	URL, err := url.Parse("http://www.omdbapi.com")

	omniAPIKey := os.Getenv("OMNIAPIKEY")
	if omniAPIKey == "" {
		log.Fatal("OMNIAPIKEY not set in env")
	}

	// URL.Path += "/some/path/or/other_with_funny_characters?_or_not/"
	parameters := url.Values{}
	parameters.Add("apikey", omniAPIKey)
	parameters.Add("t", s)
	URL.RawQuery = parameters.Encode()

	// log.Printf("Encoded URL is %q\n", URL.String())
	response, err := http.Get(URL.String())

	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		// fmt.Println(string(data))
		err := json.Unmarshal([]byte(data), jsonData)
		if err != nil {
			log.Fatal("Unmarshal failed", err)
		}
	}

	return jsonData
}

func cleanInputString(s string) string {
	keyWords := []string{"unrated multi", "limited", "proper", "bluray", "x264", "nedivx", "bestdivx", "divx", "xvid", "dvdrip", "repacked", "repack", "ac3"}

	for _, v := range keyWords {
		s = strings.Replace(s, v, "", -1)
		// log.Printf("v: %s", v)
	}
	return s
}

func cleanString(s string) string {
	log.Println("cleanString() - Input:", s)

	cleaned := strings.TrimSpace(strings.ToLower(strings.Replace(s, "\\", "", -1)))
	// cleaned = strings.Replace(cleaned, "_", " ", -1)
	// cleaned = strings.Replace(cleaned, ".", " ", -1)
	//replace words with divx in it, xvid in it, or a number in it
	cleaned = cleanInputString(cleaned)
	cleaned = strings.TrimSpace(strings.Title(cleaned))
	log.Println("Input (Cleaned) Titled:", cleaned)

	var re = regexp.MustCompile(`^(.*?) ([0-9]{3,4})[p]*(.*)$`)
	var replaceString = "$1"
	var matchedString string

	if !re.MatchString(cleaned) {
		// fmt.Println("No Match")
		matchedString = cleaned
	} else {
		matchedString = re.ReplaceAllString(cleaned, replaceString)
	}

	// fmt.Println("Potential Movie Name:", matchedString)
	// From here on end, try to see if we get a movie name
	// Movie names are preceeded by 720p/1080p/etc

	// cleaned = strings.Replace(s, "720p", "", -1)
	// cleaned = strings.Replace(s, "UNRATED MULTI", "", -1)
	// cleaned = strings.Replace(s, "BluRay", "", -1)
	// cleaned = strings.Replace(s, "x264", "", -1)

	// fmt.Println("Out:", cleaned)
	// var m *MyJsonName
	m := queryAPI(matchedString)

	if matchedString != "" {
		if m.Response != "False" {
			return fmt.Sprintf("%s (%s)", strings.TrimSpace(matchedString), m.Year)
		}
	}
	if m.Response == "False" {
		log.Printf("API Search Failed: %s\n", m.Error)
	}
	return matchedString
}

func main() {
	var dryRun bool
	var noop bool

	flag.BoolVar(&dryRun, "dryrun", true, "Display results only")
	flag.BoolVar(&noop, "noop", false, "Display results only")
	flag.Parse()
	log.Println("===+ Movie RENAMER +===\n")

	var re = regexp.MustCompile(`^(.*?) [{(\[]*([0-9]{4})[^p][\])}]*(.*)$`)
	var replaceString = "$1 ($2)"

	if dryRun {
		log.Println("=== Running in DRY MODE ===")
		log.Println("Provide -dryrun=false as an argument to actually make changes")
		log.Println("Press enter to contine...")
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		var renamed string
		if !file.IsDir() {
			continue
		} else {
			// Ignore dot dirs and anyting starting with !
			if strings.HasPrefix(file.Name(), ".")  || strings.HasPrefix(file.Name(), "!") {
				log.Printf("Ignoring: %s", file.Name())
				continue
			}
		}
		var line = file.Name()
		line = strings.Replace(line, "\\", "", -1)
		line = strings.Replace(line, "_", " ", -1)
		line = strings.Replace(line, ".", " ", -1)

		if !re.MatchString(line) {
			log.Printf("NO MATCH  => %s\n", line)
			renamed = cleanString(line)
			log.Printf("Potential => %s\n", renamed)

		} else {
			renamed = re.ReplaceAllString(line, replaceString)
			if line != renamed {
				log.Printf("    From => %s\n", line)
				log.Printf("      To => %s\n", renamed)
			} else {
				if noop {
					log.Printf("    From => %s\n", line)
					log.Printf(" NO OP for %s\n", renamed)
				}
				continue
			}
		}
		if !dryRun {
			log.Printf("Renaming %s to %s\n", file.Name(), renamed)
			if err := os.Rename(file.Name(), renamed); err != nil {
				log.Printf("Aborting: Could not rename: %v\n", file.Name(), err)
				//os.Exit(1)
			}
		}
	}
}
