package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Import resty into your code and refer it as `resty`.
import "github.com/go-resty/resty/v2"

// "organization_url": "https://api.github.com/orgs/{org}",

// import "github.com/jessevdk/go-flags"
// import "github.com/repeale/fp-go"

type GithubUrls struct {
	OrganizationUrl string `json:"organization_url"`
}

type OrganizationInfo struct {
	ReposUrl string `json:"repos_url"`
}

type Repo struct {
	Name            string `json:"name"`
	ContributorsUrl string `json:"contributors_url"`
	WatchersCount   int    `json:"watchers_count"`
}

type Contributor struct {
	Contributions int    `json:"contributions"`
	Url           string `json:"url"`
}

type User struct {
	Name  string `json:"name"`
	Login string `json:"login"`
}

func get(client *resty.Client, url string) (resp *resty.Response, err error) {
	resp, err = client.R().
		EnableTrace().
		Get(url)
	return
}

func DoItGolangStyle(organizationName string) {
	client := resty.New()

	resp, err := get(client, "https://api.github.com")
	if err != nil {
		return
	}
	respString := resp.String()
	var githubUrls GithubUrls
	err = json.Unmarshal([]byte(respString), &githubUrls)
	if err != nil {
		return
	}
	organizationUrl := githubUrls.OrganizationUrl

	organizationUrl = strings.Replace(organizationUrl, "{org}", organizationName, 1)
	resp, err = get(client, organizationUrl)
	if err != nil {
		return
	}

	respString = resp.String()
	var orgUrls OrganizationInfo
	err = json.Unmarshal([]byte(respString), &orgUrls)
	if err != nil {
		return
	}

	reposUrl := orgUrls.ReposUrl
	resp, err = get(client, reposUrl)
	if err != nil {
		return
	}

	respString = resp.String()
	var repos []Repo
	err = json.Unmarshal([]byte(respString), &repos)
	if err != nil {
		return
	}
	var mostPopular Repo
	for _, repo := range repos {
		if repo.WatchersCount > mostPopular.WatchersCount {
			mostPopular = repo
		}
	}
	contributorsUrl := mostPopular.ContributorsUrl

	resp, err = get(client, contributorsUrl)
	if err != nil {
		return
	}

	respString = resp.String()
	var contributors []Contributor
	err = json.Unmarshal([]byte(respString), &contributors)
	if err != nil {
		return
	}

	var biggestContributor Contributor
	for _, contributor := range contributors {
		if contributor.Contributions > biggestContributor.Contributions {
			biggestContributor = contributor
		}
	}

	resp, err = get(client, biggestContributor.Url)
	if err != nil {
		return
	}
	respString = resp.String()
	var user User
	err = json.Unmarshal([]byte(respString), &user)
	if err != nil {
		return
	}

	fmt.Println("The largest contributor to go is:", user.Name, "(", user.Login, ")")
}

func main() {
	DoItGolangStyle("golang")
}
