package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type Config struct {
	Contrat struct {
		DateDebutContrat string  `json:"dateDebutContrat"`
		DateFinContrat   string  `json:"dateFinContrat"`
		NombreHeureTotal string  `json:"nombreHeureTotal"`
		SalaireDeBase    float64 `json:"salaireDeBase"`
		NombreDeCa       float64 `json:"nombreDeCa"`
	} `json:"contrat"`
	Tarifs struct {
		TauxHoraire float64 `json:"tauxHoraire"`
		Entretien   float64 `json:"entretien"`
		Gouter      float64 `json:"gouter"`
		Repas       float64 `json:"repas"`
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

func lastDate(date string) int {

	mydate := strings.Split(date, "-")
	var year, month int
	year, _ = strconv.Atoi(mydate[0])
	month, _ = strconv.Atoi(mydate[1])
	// Given a month and a year, return the last
	// date of the year.
	//
	// Just bump it up a month, subtract an hour, and grab
	// that date.

	if month == 12 {
		year += 1
	}
	month += 1

	t := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	prev := t.Add(-time.Hour)

	return prev.Day()

}

func main() {
	ctx := context.Background()
	var startPeriod = flag.String("start", "2015-09-01", "Date de début de période (format YYYY-MM-DD)")
	var endPeriod = flag.String("end", "2016-08-31", "Date de début de période (Format YYYY-MM-DD)")
	var month = flag.String("month", "", "Mois du calcul (format YYYY-MM)")
	flag.Parse()
	if *month != "" {
		*startPeriod = fmt.Sprintf("%v-01", *month)
		endDay := lastDate(*month)
		*endPeriod = fmt.Sprintf("%v-%v", *month, endDay)
	}
	// Read the global config
	conf, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Unable to read client config file: %v", err)
	}
	var myconfig Config
	err = json.Unmarshal(conf, &myconfig)
	if err != nil {
		log.Fatalf("Unable to read parse config file: %v", err)
	}
	startCalcul, err := time.Parse("2006-01-02", *startPeriod)
	if err != nil {
		log.Fatalf("Unable to read parse startDate : %v", err)
	}
	endCalcul, err := time.Parse("2006-01-02", *endPeriod)
	if err != nil {
		log.Fatalf("Unable to read parse endDate : %v", err)
	}
	//debutContrat, _ := time.Parse("2006-Jan-02", myconfig.Contrat.DateDebutContrat)
	//finContrat, _ := time.Parse("2006-Jan-02", myconfig.Contrat.DateFinContrat)
	//fmt.Printf("Contrat:\n\tDebut: %v\n\tFin: %v\n", debutContrat, finContrat)

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
	//startEvent := time.Now().Format(time.RFC3339)
	var eleonore = regexp.MustCompile(`Eléonore`)
	var eugenie = regexp.MustCompile(`Eugénie`)
	var caNounou = regexp.MustCompile(`CA`)
	nombreCA := 0
	nombreCADepuisLeDebut := 0
	nombreDeGouter := 0.0
	nombreDeRepas := 0.0
	nombreDeJourDepuisLeDebut := 0
	nombreDeJour := 0
	var duree time.Duration
	var dureeDepuisLeDebut time.Duration
	// Calcul des jours de la période en cours
	events, err := srv.Events.List("ug8gqc2m8qr0hdr012lf5grc14@group.calendar.google.com").ShowDeleted(false).
		SingleEvents(true).TimeMin(startCalcul.Format(time.RFC3339)).TimeMax(endCalcul.Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}
	if len(events.Items) > 0 {
		for _, i := range events.Items {
			if eleonore.MatchString(i.Summary) || eugenie.MatchString(i.Summary) {
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
				//jour := fmt.Sprintf("%v-%v-%v", start.Day(), start.Month(), start.Year())
				nombreDeJour = nombreDeJour + 1
				if eleonore.MatchString(i.Summary) {
					nombreDeRepas = nombreDeRepas + 1
					if end.Hour() >= 16 {
						nombreDeGouter = nombreDeGouter + 1
					}
				}
			}
			if caNounou.MatchString(i.Summary) {
				nombreCA = nombreCA + 1
			}
		}
	} else {
		fmt.Printf("No upcoming events found.\n")
	}
	// Calcul des jours depuis le début du contrat
	dateDebutContrat, err := time.Parse("2006-Jan-02", myconfig.Contrat.DateDebutContrat)
	if err != nil {
		log.Fatalf("Cannnt parse date debut contrat")
	}
	events, err = srv.Events.List("ug8gqc2m8qr0hdr012lf5grc14@group.calendar.google.com").ShowDeleted(false).
		SingleEvents(true).TimeMin(dateDebutContrat.Format(time.RFC3339)).TimeMax(endCalcul.Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}
	if len(events.Items) > 0 {
		for _, i := range events.Items {
			if eleonore.MatchString(i.Summary) || eugenie.MatchString(i.Summary) {
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
				dureeDepuisLeDebut = dureeDepuisLeDebut + dureeAcceuil
				//fmt.Printf("%s %s (%v)\n", i.Summary, when, dureeAcceuil)
				//jour := fmt.Sprintf("%v-%v-%v", start.Day(), start.Month(), start.Year())
				nombreDeJourDepuisLeDebut = nombreDeJourDepuisLeDebut + 1

			}
			if caNounou.MatchString(i.Summary) {
				nombreCADepuisLeDebut = nombreCADepuisLeDebut + 1
			}
		}
	} else {
		fmt.Printf("No upcoming events found.\n")
	}
	//salaireNet := duree.Hours() * myconfig.Tarifs.TauxHoraire
	fmt.Printf("Calcul pour la période de %v à %v\n", *startPeriod, *endPeriod)
	fmt.Printf("\tNombre de jours d'accueil: %v\n", nombreDeJour)
	fmt.Printf("\tDurée d'accueil: %v heures (%v depuis %v / %v)\n", duree.Hours(), dureeDepuisLeDebut.Hours(), myconfig.Contrat.DateDebutContrat, myconfig.Contrat.NombreHeureTotal)

	fmt.Printf("\tSalaire de base: %v€\n", myconfig.Contrat.SalaireDeBase)
	//fmt.Printf("\tSalaire net hypothetique: %v\n", salaireNet)
	fmt.Printf("\tNombre de CA   : %v (%v/%v)\n", nombreCA, nombreCADepuisLeDebut, myconfig.Contrat.NombreDeCa)
	fmt.Printf("Gouter:\n\tNombre: %v\n\tA payer: %v€\n", nombreDeGouter, nombreDeGouter*myconfig.Tarifs.Gouter)
	fmt.Printf("Repas:\n\tNombre: %v\n\tA payer: %v€\n", nombreDeRepas, nombreDeRepas*myconfig.Tarifs.Repas)
	fmt.Printf("Entretien:\n\tNombre: %v\n\tA payer: %v€\n", nombreDeJour, float64(nombreDeJour)*myconfig.Tarifs.Entretien)
	fmt.Printf("\n\nNet a payer: %v€\n", myconfig.Contrat.SalaireDeBase+nombreDeGouter*myconfig.Tarifs.Gouter+nombreDeRepas+nombreDeRepas*myconfig.Tarifs.Repas+float64(nombreDeJour)*myconfig.Tarifs.Entretien)
}
