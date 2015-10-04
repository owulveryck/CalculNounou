package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type Config struct {
	Contrat struct {
		DateDebutContrat string        `json:"dateDebutContrat"`
		DateFinContrat   string        `json:"dateFinContrat"`
		NombreHeureTotal time.Duration `json:"nombreHeureTotal"`
		SalaireDeBase    float64       `json:"salaireDeBase"`
	} `json:"contrat"`
	Tarifs struct {
		Entretien float64 `json:"entretien"`
		Gouter    float64 `json:"gouter"`
		Repas     float64 `json:"repas"`
	} `json:"tarifs"`
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-api-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	// Read the global config

	// Read the secret file
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	// Getting calendar list
	/*
		calendarList, err := srv.CalendarList.List().Do()
		if err != nil {
			log.Fatalf("Unable to retrieve the calendar list. %v", err)
		}
		for _, i := range calendarList.Items {
			fmt.Printf("Id:%s, Summary:%s\n", i.Id, i.Summary)
		}
	*/
	// Events
	startYear := time.Date(2015, time.September, 1, 1, 0, 0, 0, time.UTC).Format(time.RFC3339)
	endYear := time.Date(2016, time.August, 31, 23, 0, 0, 0, time.UTC).Format(time.RFC3339)
	//startEvent := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List("ug8gqc2m8qr0hdr012lf5grc14@group.calendar.google.com").ShowDeleted(false).
		SingleEvents(true).TimeMin(startYear).TimeMax(endYear).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}
	var eleonoreEugenie = regexp.MustCompile(`Eléonore|Eugénie`)
	var caNounou = regexp.MustCompile(`CA`)
	nombreCA := 0
	var duree time.Duration
	if len(events.Items) > 0 {
		for _, i := range events.Items {
			if eleonoreEugenie.MatchString(i.Summary) {
				// Check des evenements Eléonore et Eugénie
				var when string
				// If the DateTime is an empty string the Event is an all-day Event.
				// So only Date is available.
				if i.Start.DateTime != "" {
					when = i.Start.DateTime
				} else {
					when = i.Start.Date
				}
				// play with time formats
				start, err := time.Parse(time.RFC3339, when)
				if err != nil {
					log.Printf("Unable to parse start date (%s). %v", when, err)
				}
				end, err := time.Parse(time.RFC3339, i.End.DateTime)
				if err != nil {
					log.Printf("Unable to parse start date. %v", err)
				}
				dureeAcceuil := end.Sub(start)
				duree = duree + dureeAcceuil
				//fmt.Printf("%s %s (%v)\n", i.Summary, when, dureeAcceuil)
			}
			if caNounou.MatchString(i.Summary) {
				nombreCA = nombreCA + 1
			}
		}
	} else {
		fmt.Printf("No upcoming events found.\n")
	}
	fmt.Printf("Duree Total d'acceuil: %s\n", duree)
	fmt.Printf("Nombre de CA Total   : %v\n", nombreCA)
}
