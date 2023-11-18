/*
llg is Lakin's place to learn go
Copyright (C) 2023  Lakin Wecker

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Import resty into your code and refer it as `resty`.
import "github.com/go-resty/resty/v2"

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

func DoItGolangStyle(organizationName string) (user User, err error) {
	client := resty.New()

	resp, err := get(client, "https://api.github.com")
	if err != nil {
		err = fmt.Errorf("Error getting github urls: %w", err)
		return
	}
	fmt.Println(resp.String())
	respString := resp.String()
	var githubUrls GithubUrls
	err = json.Unmarshal([]byte(respString), &githubUrls)
	if err != nil {
		err = fmt.Errorf("Error getting parsing urls: %w", err)
		return
	}
	organizationUrl := githubUrls.OrganizationUrl
	organizationUrl = strings.Replace(organizationUrl, "{org}", organizationName, 1)

	resp, err = get(client, organizationUrl)
	if err != nil {
		err = fmt.Errorf("Error getting organization info from %s: %w", organizationUrl, err)
		return
	}

	respString = resp.String()
	var orgUrls OrganizationInfo
	err = json.Unmarshal([]byte(respString), &orgUrls)
	if err != nil {
		err = fmt.Errorf("Error parsing organization info: %w", err)
		return
	}

	reposUrl := orgUrls.ReposUrl
	resp, err = get(client, reposUrl)
	if err != nil {
		err = fmt.Errorf("Error getting repos url: %w", err)
		return
	}

	respString = resp.String()
	var repos []Repo
	err = json.Unmarshal([]byte(respString), &repos)
	if err != nil {
		err = fmt.Errorf("Error parsing repos: %w", err)
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
		err = fmt.Errorf("Error getting contributorUrl: %w", err)
		return
	}

	respString = resp.String()
	var contributors []Contributor
	err = json.Unmarshal([]byte(respString), &contributors)
	if err != nil {
		err = fmt.Errorf("Error parsing contributors: %w", err)
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
		err = fmt.Errorf("Error getting biggest contributor: %w", err)
		return
	}
	respString = resp.String()
	err = json.Unmarshal([]byte(respString), &user)
	if err != nil {
		err = fmt.Errorf("Error parsing biggest contributor: %w", err)
		return
	}
	return

}

func main() {
	user, err := DoItGolangStyle("golang")
	if err != nil {
		fmt.Println("Error golang style:", err)
	} else {
		fmt.Println("The largest contributor to go is:", user.Name, "(", user.Login, ")")
	}
}
