package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

type Results []struct {
	Title        string    `json:"Title"`
	Name         string    `json:"Name"`
	Domain       string    `json:"Domain"`
	BreachDate   string    `json:"BreachDate"`
	AddedDate    string    `json:"AddedDate"`
	ModifiedDate time.Time `json:"ModifiedDate"`
	PwnCount     int       `json:"PwnCount"`
	Description  string    `json:"Description"`
	DataClasses  []string  `json:"DataClasses"`
	IsVerified   bool      `json:"IsVerified"`
	IsFabricated bool      `json:"IsFabricated"`
	IsSensitive  bool      `json:"IsSensitive"`
	IsActive     bool      `json:"IsActive"`
	IsRetired    bool      `json:"IsRetired"`
	IsSpamList   bool      `json:"IsSpamList"`
	LogoType     string    `json:"LogoType"`
}

var email string

type Info []struct {
	Source     string      `json:"Source"`
	ID         string      `json:"Id"`
	Title      string      `json:"Title"`
	Date       interface{} `json:"Date"`
	EmailCount int         `json:"EmailCount"`
}

func input() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("[+] Enter Email Address : ")
	scanner.Scan()
	email = scanner.Text()
	return email
}

func fetchpastebin(link string) {
	url := "https://www.pastebin.com/raw/"
	res, err := http.Get(url + link)

	if err != nil {
		panic(err.Error())
	}

	if res.StatusCode != http.StatusNotFound {

		body, _ := ioutil.ReadAll(res.Body)
		results := string(body)
		r, _ := regexp.Compile(email + ":([a-zA-Z0-9]+)")
		if len(r.FindAllString(results, -1)) > 0 {
			fmt.Println(r.FindAllString(results, -1))
		}
	}
}

func fetchpasteaccount() {
	url := "https://haveibeenpwned.com/api/v2/pasteaccount/"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url+email, nil)
	req.Header.Add("User-Agent", "igat")
	res, err := client.Do(req)
	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	var info Info
	json.Unmarshal(body, &info)

	var proceed bool = false
	for _, data := range info {
		if data.Source == "Pastebin" {
			proceed = true
		}
	}
	if proceed {
		fmt.Println("[+] Dumps Found...!")
		fmt.Println()
		fmt.Println("[+] Looking for Passwords...this may take a while...")
		fmt.Println()
	} else {
		fmt.Println("[-] No Dumps Found... :(")
		fmt.Println()
	}
	for _, data := range info {
		if data.Source == "Pastebin" {
			go fetchpastebin(data.ID)
		}
	}
}

func fetchbreachedaccount() interface{} {
	url := "https://haveibeenpwned.com/api/v2/breachedaccount/"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url+email, nil)
	req.Header.Add("User-Agent", "igat")
	res, err := client.Do(req)

	if err != nil {
		panic(err.Error())
	}

	if res.StatusCode == http.StatusNotFound {
		return fmt.Sprintln("[-] Account not pwned... :(")
	}

	fmt.Println("[!] Account pwned...Listing Breaches...")
	fmt.Println()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	var results Results
	json.Unmarshal(body, &results)

	return results
}

func getemailsfromfile(source string) {
	file, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Printf("[+] Checking Breach status for %s\n", scanner.Text())
		fmt.Println()
		time.Sleep(2 * time.Second)
		email = scanner.Text()
		getdata(email)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getdata(account string) {
	start := time.Now()
	data := fetchbreachedaccount()

	if results, ok := data.(Results); ok {
		for _, result := range results {
			fmt.Printf("[+] Breach      : %s\n", result.Title)
			fmt.Printf("[+] Domain      : %s\n", result.Domain)
			fmt.Printf("[+] Date        : %s\n", result.AddedDate)
			fmt.Printf("[+] Fabricated  : %t\n", result.IsFabricated)
			fmt.Printf("[+] Verified    : %t\n", result.IsVerified)
			fmt.Printf("[+] Retired     : %t\n", result.IsRetired)
			fmt.Printf("[+] Spam        : %t\n", result.IsSpamList)
			fmt.Println()
		}
		fetchpasteaccount()
		fmt.Printf("[+] Completed in %.2fs\n", time.Since(start).Seconds())
		fmt.Println()
	} else {
		fmt.Println(data)
	}
}

func main() {
	email := flag.String("email", "", "Email account you want to test")
	file := flag.String("file", "", "Load a file with multiple email accounts")
	flag.Parse()
	if *file != "" {
		getemailsfromfile(*file)
	}
	if *email != "" {
		getdata(*email)
		os.Exit(0)
	}
	if len(flag.Args()) > 0 {
		flag.PrintDefaults()
		os.Exit(1)
	} else {
		input := input()
		getdata(input)
	}

}
