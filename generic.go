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
	if err != nil {
		err = fmt.Errorf("Error getting url(%s): %w", url, err)
		return
	}
	return
}

func FromJson[T any](jsonString string) (val T, err error) {
	err = json.Unmarshal([]byte(jsonString), &val)
	if err != nil {
		err = fmt.Errorf("Error getting parsing into %T: %w", val, err)
		return
	}
	return

}

func GetGithubUrls(client *resty.Client) (githubUrls GithubUrls, err error) {
	resp, err := get(client, "https://api.github.com")
	if err != nil {
		return
	}
	githubUrls, err = FromJson[GithubUrls](resp.String())
	return
}

func GetOrganizationInfo(client *resty.Client, githubUrls GithubUrls, organizationName string) (organizationInfo OrganizationInfo, err error) {
	organizationUrl := githubUrls.OrganizationUrl
	organizationUrl = strings.Replace(organizationUrl, "{org}", organizationName, 1)

	resp, err := get(client, organizationUrl)
	if err != nil {
		return
	}

	organizationInfo, err = FromJson[OrganizationInfo](resp.String())
	return
}

func GetOrganizationRepos(client *resty.Client, organizationInfo OrganizationInfo) (repos []Repo, err error) {
	reposUrl := organizationInfo.ReposUrl
	resp, err := get(client, reposUrl)
	if err != nil {
		return
	}

	repos, err = FromJson[[]Repo](resp.String())
	return
}

func GetUserInfo(client *resty.Client, contributor Contributor) (user User, err error) {
	resp, err := get(client, contributor.Url)
	if err != nil {
		return
	}

	user, err = FromJson[User](resp.String())
	return
}

func GetMostPopularRepo(repos []Repo) (mostPopular Repo) {
	for _, repo := range repos {
		if repo.WatchersCount > mostPopular.WatchersCount {
			mostPopular = repo
		}
	}
	return
}

func GetBiggestContributor(contributors []Contributor) (biggestContributor Contributor) {
	for _, contributor := range contributors {
		if contributor.Contributions > biggestContributor.Contributions {
			biggestContributor = contributor
		}
	}
	return
}

func GetContributors(client *resty.Client, mostPopular Repo) (contributors []Contributor, err error) {
	contributorsUrl := mostPopular.ContributorsUrl

	resp, err := get(client, contributorsUrl)
	if err != nil {
		return
	}

	contributors, err = FromJson[[]Contributor](resp.String())
	return
}

func DoItGolangStyle(organizationName string) (user User, err error) {
	client := resty.New()

	githubUrls, err := GetGithubUrls(client)
	if err != nil {
		return
	}

	organizationInfo, err := GetOrganizationInfo(client, githubUrls, organizationName)
	if err != nil {
		return
	}

	repos, err := GetOrganizationRepos(client, organizationInfo)
	if err != nil {
		return
	}

	mostPopular := GetMostPopularRepo(repos)
	contributors, err := GetContributors(client, mostPopular)
	if err != nil {
		return
	}

	biggestContributor := GetBiggestContributor(contributors)
	user, err = GetUserInfo(client, biggestContributor)

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
