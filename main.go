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
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"
	"github.com/parnurzeal/gorequest"
	"gopkg.in/yaml.v2"
)

type User struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Handle      string `json:"handle"`
	Email       string `json:"email"`
	Password    string `json:password`
	LatestCheck time.Time
	Packages    []Package
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	Cmd            string `json:"cmd"`
	Private        bool   `json:"private"`
	User           User
	UserId         int64 `json:"user_id"`
	TotalDownloads int64 `json:"total_downloads"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

var CurrentUser = User{}
var CurrentPackage = Package{}
var EndPoint = "https://composehub.com"
var Dev, Version string

func init() {
	Version = "0.0.0"
	if os.Getenv("ENDPOINT") != "" {
		EndPoint = os.Getenv("ENDPOINT")
	}
	createConfigDir()
	CurrentUser = getCurrentUser()
	checkUpdateCheckFile()
	updateCheckFile()
	CurrentPackage = getCurrentPackage("")
	Dev = os.Getenv("DEV")
	devlog(CurrentUser)
}

func main() {
	app := cli.NewApp()
	app.Name = "ComposeHub"
	app.Usage = "Install and publish docker compose packages."
	app.Version = Version
	app.EnableBashCompletion = true
	app.Action = cli.ShowAppHelp
	app.Commands = []cli.Command{
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "ch install <package>",
			Action:  installAction,
		},
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "ch search <package> ",
			Action:  search,
		},
		{
			Name:    "adduser",
			Aliases: []string{"a"},
			Usage:   "ch adduser",
			Action:  adduserAction,
		},
		{
			Name:    "updateuser",
			Aliases: []string{"uu"},
			Usage:   "ch updateuser",
			Action:  updateuserAction,
		},
		{
			Name:    "publish",
			Aliases: []string{"p"},
			Usage:   "ch publish",
			Action:  publishAction,
		},
		{
			Name:    "init",
			Aliases: []string{"in"},
			Usage:   "ch init",
			Action:  initAction,
		},
		{
			Name:    "configuser",
			Aliases: []string{"cu"},
			Usage:   "ch init",
			Action:  updateuserAction,
		},
		{
			Name:    "resetpassord",
			Aliases: []string{"rp"},
			Usage:   "ch resetpassword",
			Action:  resetpassordAction,
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "ch run <package>",
			Action:  runAction,
		},
	}
	app.Run(os.Args)
}
func runAction(c *cli.Context) {
	install(c, false)
	/*wg := new(sync.WaitGroup)*/

	if CurrentPackage.Cmd != "" {
		cmd := exec.Command("sh", "-c", CurrentPackage.Cmd, ".")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		println(CurrentPackage.Description)
	}
}

func search(c *cli.Context) {
	q := c.Args().First()
	println("Searching for", q+" on https://composehub.com...") /*"+EndPoint+"...")*/
	if resp, err := http.Get(EndPoint + "/search/" + q + timestamp()); err != nil {
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
	println("creating composehub.yml (done)")
	yml := `---
name: package-name
blurb: 80 chars line blurb
description: |
  longer description
email: ` + CurrentUser.Email + `
repo_url: http://github.com/foo/bar
tags: tag1,tag2
private: false
cmd: 
`
	println("please edit composehub.yml and the run `ch publish`")
	err := ioutil.WriteFile("composehub.yml", []byte(yml), 0644)
	if err != nil {
		println(err)
	}
}

func installAction(c *cli.Context) {
	install(c, true)
}

func publishAction(c *cli.Context) {
	if CurrentUser.Email == "" {
		message := `
        Please create a user account first, run 'ch adduser'
        If you already have an account, please run 'ch updateuser'
`
		println(message)
		return
	}
	p := getCurrentPackage("")
	u := EndPoint + "/publish/" + p.Name
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
		u := EndPoint + "/users/" + e

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
			createUserConfig(user.Handle, user.Email, newPassword)
		})
	}
}
func adduserAction(c *cli.Context) {
	if handle, email, password, err := promptUserInfo(c, false); err != nil {
		return
	} else {

		if resp, err := http.PostForm(EndPoint+"/users",
			url.Values{"handle": {handle}, "email": {email}, "password": {password}}); err != nil {
			println("Sorry, the query failed with the following message: ", err)
			return
		} else {
			defer resp.Body.Close()
			if body, err := ioutil.ReadAll(resp.Body); err != nil {
				fmt.Println(err, string(body))
				return
			} else {
				devlog(err, string(body))
				createUserConfig(handle, email, password)
				CurrentUser = getCurrentUser()
			}
		}
	}
}
func resetpassordAction(c *cli.Context) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter email: ")
	email, _ := reader.ReadString('\n')
	email = strings.Replace(email, "\n", "", -1)

	u := EndPoint + "/users/" + email + "/reset-password"
	request := gorequest.New()
	request.Post(u).
		End(func(resp gorequest.Response, body string, errs []error) {
		if resp.StatusCode == 200 {
			fmt.Print("Please check your email, copy the token and paste it here: ")
			token, _ := reader.ReadString('\n')
			token = strings.Replace(token, "\n", "", -1)
			fmt.Printf("Password: ")
			pass := gopass.GetPasswd() // Silent, for *'s use gopass.GetPasswdMasked()
			password := strings.Replace(string(pass), "\n", "", -1)

			resetPassword(email, token, password)
		} else {
			println(body)
		}
	})
}

func resetPassword(email, token, password string) {
	devlog(email, token, password)
	u := EndPoint + "/users/" + email + "/reset-password/" + token
	request := gorequest.New()
	request.Put(u).
		Send(User{Password: password}).
		End(func(resp gorequest.Response, body string, errs []error) {
		if resp.StatusCode == 200 {
			fmt.Print("Your password has been updated successfully!")
			user := User{}
			err := json.Unmarshal([]byte(body), &user)
			if err != nil {
				log.Fatalf("no config found", err)
				return
			}
			user.Password = password
			CurrentUser = user
			createUserConfig(user.Handle, user.Email, user.Password)
		} else {
			println(body)
		}
	})
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

func createUserConfig(handle, email, password string) {
	if path, err := composeHubConfigPath(); err != nil {
		fmt.Println(err)
		return
	} else {

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
	if path, err := composeHubConfigPath(); err != nil {
		fmt.Println(err)
		return u
	} else {
		data, _ := ioutil.ReadFile(path + "/config.yml")
		err = yaml.Unmarshal(data, &u)
		if err != nil {
			log.Fatalf("no config found", err)
		}
		return u
	}
}

func getCurrentPackage(pkg string) Package {
	if pkg == "" {
		pkg = "composehub.yml"
	}
	p := Package{}
	data, _ := ioutil.ReadFile(pkg)
	err := yaml.Unmarshal(data, &p)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	devlog(p)
	return p
}

func devlog(v ...interface{}) {
	if Dev != "" {
		devlog(v)
	}
}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func composeHubConfigPath() (string, error) {
	if home, err := homedir.Dir(); err != nil {
		fmt.Println(err)
		return "", err
	} else {
		path := home + "/.composehub"
		return path, err
	}
}

func createConfigDir() {
	if path, err := composeHubConfigPath(); err != nil {
		fmt.Println(err)
		return
	} else {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(path, 0700); err != nil {
					fmt.Println(err)
				}
				// file does not exist
			} else {
				fmt.Println(err)
				// other error
			}
		}
	}
}

func updateCheckFile() {
	if path, err := composeHubConfigPath(); err != nil {
		fmt.Println(err)
	} else {
		err := ioutil.WriteFile(path+"/versioncheck", []byte(string(time.Now().Format(time.RFC3339))), 0644)
		if err != nil {
			println(err)
		}
	}
}

func checkUpdateCheckFile() {
	if path, err := composeHubConfigPath(); err != nil {
		fmt.Println(err)
	} else {
		data, _ := ioutil.ReadFile(path + "/versioncheck")
		t, err := time.Parse(time.RFC3339, string(data))
		devlog("since:", time.Since(t), err)

		if time.Since(t).Hours() > float64(48) {
			/*if true {*/
			checkForUpdate()
		}
	}
}

func checkForUpdate() {
	if resp, err := http.Get(EndPoint + "/checkupdate/" + Version + timestamp()); err != nil {
		println("Sorry, the query failed with the following message: ", err)
		return
	} else {
		defer resp.Body.Close()
		if body, err := ioutil.ReadAll(resp.Body); err != nil {
			println("Sorry, the query failed with the following message: ", err)
			println(string(body))
			return
		} else {
			b := string(body)
			if b != "\"ok\"" {
				println("There's a new version " + b + " available")
				/*println("curl -L https://composehub.org/install/" + runtime.GOOS + "&GOARCH=" + runtime.GOARCH + " > /usr/local/bin/docker-compose")*/
				println("curl -L https://composehub.com/install/" + runtime.GOOS)
			}
		}

	}
}

func install(c *cli.Context, showDescription bool) Package {
	p := Package{}
	q := c.Args().First()
	if res, err := isEmpty("."); !res || err != nil {
		println("This dir is not empty. You can only install packages in empty directories.")
		println("Try `mkdir " + q + " && cd " + q + "`")
		return p
	}
	u := EndPoint + "/packages/" + q + timestamp()
	request := gorequest.New().SetBasicAuth(CurrentUser.Email, CurrentUser.Password)
	request.Get(u).
		End(func(resp gorequest.Response, body string, errs []error) {
		if resp.StatusCode == 200 {
			if err := json.Unmarshal([]byte(body), &p); err != nil {
				log.Fatalf("no config found", err)
				return
			} else {
				CurrentPackage = p
				devlog(p)
			}
			/*println("Cloning repo...\n")*/
			cmd := exec.Command("git", "clone", p.RepoUrl, ".")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			println("Package installed successfully!\n")
			if showDescription {
				println(p.Description)
			}
		} else {
			println(body)
		}
	})
	return p
}

func timestamp() string {
	return "?t=" + time.Now().UTC().Format("20060102150405")
}

/*mkdir my-gitlab*/
/*cd my-gitlab*/
/*ch search gitlab*/
/*ch install gitlab*/
/*docker-compose up*/

/*cd ..*/
/*mkdir my-wordpress*/
/*cd my-wordpress*/
/*ch search run wordpress*/

/*mkdir my-rails-app*/
/*cd my-rails-app*/
/*ch install rails*/
/*docker-compose run web bundle install*/
