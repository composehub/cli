package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/howeyc/gopass"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/yaml.v2"
)

type User struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	Handle    string `json:"handle"`
	Email     string `json:"email"`
	Password  string `json:password`
	Packages  []Package
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Package struct {
	Id             int64  `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Tags           string `json:"tags"`
	Blurb          string `json:"blurb"`
	Description    string `json:"description"`
	RepoUrl        string `json:"repo_url" yaml:"repo_url"`
	Commit         string `json:"commit"`
	User           User
	UserId         int64 `json:"user_id"`
	TotalDownloads int64 `json:"total_downloads"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var CurrentUser = User{}

func init() {
	CurrentUser = getCurrentUser()
	log.Println(CurrentUser)
}

func main() {
	app := cli.NewApp()
	app.Name = "ComposeHub"
	app.Usage = "Install and publish docker compose packages."
	app.Action = func(c *cli.Context) {
		println("boom! I say!")
	}
	app.Commands = []cli.Command{
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "chm install <package> ",
			Action: func(c *cli.Context) {
				println("installed package: ", c.Args().First())
			},
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "chm search <package> ",
			Action:  search,
		},
		{
			Name:    "adduser",
			Aliases: []string{"a"},
			Usage:   "chm adduser",
			Action:  adduserAction,
		},
		{
			Name:    "updateuser",
			Aliases: []string{"uu"},
			Usage:   "chm updateuser",
			Action:  updateuserAction,
		},
		{
			Name:    "publish",
			Aliases: []string{"p"},
			Usage:   "chm publish",
			Action:  publishAction,
		},
		{
			Name:    "init",
			Aliases: []string{"in"},
			Usage:   "chm init",
			Action:  initAction,
		},
		{
			Name:    "configuser",
			Aliases: []string{"cu"},
			Usage:   "chm init",
			Action:  updateuserAction,
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "chm run <package>",
			Action: func(c *cli.Context) {
				println("Running package ", c.Args().First())
			},
		},
	}
	app.Run(os.Args)
}
func search(c *cli.Context) {
	q := c.Args().First()
	println("Searching for", q+"...")
	if resp, err := http.Get("http://plasti.co:3000/search/" + q); err != nil {
		println("Sorry, the query failed with the following message: ", err)
		return
	} else {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			println("Sorry, the query failed with the following message: ", err)
			println(string(body))
			return
		} else {
			packages := []Package{}
			json.Unmarshal(body, &packages)
			for _, p := range packages {
				println(p.Name+":", p.Blurb, "(by "+p.User.Handle+")")

			}
			println("found", len(packages), "packages")
		}
	}
}

func initAction(c *cli.Context) {
	println("creating package.yml (done)")
	yml := `---
name: package-name
blurb: 80 chars line blurb
description: |
  longer description
email: your email
repo_url: your git repo url
tags: [web, framework]
`
	println("please edit package.yml and the run `chm publish`")
	err := ioutil.WriteFile("package.yml", []byte(yml), 0644)
	if err != nil {
		println(err)
	}
}

func publishAction(c *cli.Context) {
	if CurrentUser.Email == "" {
		message := `
        Please create a user account first, run 'chm adduser'
        If you already have an account, please run 'chm update user'
`
		println(message)
		return
	}
	p := Package{}
	data, _ := ioutil.ReadFile("package.yml")
	err := yaml.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	u := "http://plasti.co:3000/publish/" + p.Name
	fmt.Println(p.Name, p.RepoUrl, p)
	request := gorequest.New().SetBasicAuth(CurrentUser.Email, CurrentUser.Password)
	request.Post(u).
		Send(p).
		End(func(resp gorequest.Response, body string, errs []error) {
		if resp.StatusCode == 200 {
			println("Package update successfully!\n")
		} else {
			println(body)

		}
	})

	/*type Config struct {*/
	/*Foo string*/
	/*Bar []string*/
	/*}*/

	/*filename := os.Args[2]*/
	/*var config Config*/
	/*source, err := ioutil.ReadFile(filename)*/
	/*if err != nil {*/
	/*panic(err)*/
	/*}*/
	/*err = yaml.Unmarshal(source, &config)*/
	/*if err != nil {*/
	/*panic(err)*/
	/*}*/
	/*fmt.Printf("Value: %#v\n", config.Bar[0])*/
}

func updateuserAction(c *cli.Context) {
	if handle, email, password, err := promptUserInfo(c, true); err != nil {
		return
	} else {
		user := User{Handle: handle, Password: password, Email: email}
		fmt.Println("user: ", user)
		e, p := email, password
		if CurrentUser.Email != "" {
			e, p = CurrentUser.Email, CurrentUser.Password
		}
		u := "http://plasti.co:3000/users/" + e

		request := gorequest.New().SetBasicAuth(e, p)
		request.Put(u).
			Send(user).
			End(func(resp gorequest.Response, body string, errs []error) {
			fmt.Println(resp.Status)
			err = json.Unmarshal([]byte(body), &user)
			if err != nil {
				log.Fatalf("no config found", err)
				return
			}
			newPassword := CurrentUser.Password
			if password != "" {
				newPassword = password
			}
			CurrentUser = user
			createCHMDir(user.Handle, user.Email, newPassword)
		})
	}
}
func adduserAction(c *cli.Context) {
	if handle, email, password, err := promptUserInfo(c, false); err != nil {
		return
	} else {

		if resp, err := http.PostForm("http://plasti.co:3000/users",
			url.Values{"handle": {handle}, "email": {email}, "password": {password}}); err != nil {
			println("Sorry, the query failed with the following message: ", err)
			return
		} else {
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err != nil {
				fmt.Println(err, string(body))
				return
			} else {
				fmt.Println(err, string(body))
				createCHMDir(handle, email, password)
				CurrentUser = getCurrentUser()
			}
		}
	}
}

func promptUserInfo(c *cli.Context, update bool) (string, string, string, error) {
	reader := bufio.NewReader(os.Stdin)
	updateMsg := ""
	if update {
		updateMsg = "(leave blank if you don't want to update) "
	}
	handle := ""
	if CurrentUser.Handle != "" || !update {
		fmt.Print("Enter handle: " + updateMsg)
		handle, _ = reader.ReadString('\n')
	}
	fmt.Print("Enter email: " + updateMsg)
	email, _ := reader.ReadString('\n')
	if !update {
		fmt.Printf("Password: (leave blank to auto-generate one)")

	} else {
		fmt.Printf("Password: " + updateMsg)

	}
	pass := gopass.GetPasswd() // Silent, for *'s use gopass.GetPasswdMasked()
	// Do something with pass'
	handle = strings.Replace(handle, "\n", "", -1)
	password := strings.Replace(string(pass), "\n", "", -1)
	email = strings.Replace(email, "\n", "", -1)

	if !update && password == "" {
		uuid, err := newUUID()
		if err != nil {
			fmt.Printf("Sorry, something wrong happened\n", err)
			return handle, email, password, errors.New("")
		} else {
			password = uuid
		}
	}
	if !update && handle == "" {
		fmt.Printf("Invalid handle\n")
		return handle, email, password, errors.New("")
	}
	if !validateEmail(email) && !update && email != "" {
		fmt.Printf("invalid email\n")
		return handle, email, password, errors.New("")
	}
	return handle, email, password, nil

}

func createCHMDir(handle, email, password string) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	path := usr.HomeDir + "/.composehub"
	if err := os.Mkdir(path, 0700); err != nil {
		fmt.Println(err)
	}
	config := `---
handle: ` + handle + `
email: ` + email + `
password: ` + password + `
`
	if err := ioutil.WriteFile(path+"/config.yml", []byte(config), 0600); err != nil {
		if os.IsExist(err) {
			fmt.Println("Looks like " + path + "/config.yml already exists, please remove it or edit it manually.")
		} else {
			fmt.Println(err)
		}

	}

}

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func getCurrentUser() User {
	u := User{}
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	path := usr.HomeDir + "/.composehub"
	data, _ := ioutil.ReadFile(path + "/config.yml")
	err = yaml.Unmarshal(data, &u)
	if err != nil {
		log.Fatalf("no config found", err)
	}
	return u
}
