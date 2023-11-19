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

import E "github.com/IBM/fp-go/either"
import F "github.com/IBM/fp-go/function"

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

func get(client *resty.Client, url string) E.Either[error, *resty.Response] {
	resp, err := client.R().
		EnableTrace().
		Get(url)
	if err != nil {
		return E.Left[*resty.Response](fmt.Errorf("Error getting url(%s): %w", url, err))
	}
	return E.Right[error](resp)
}

func FromJson[T any](jsonString string) E.Either[error, T] {
	var val T
	err := json.Unmarshal([]byte(jsonString), &val)
	if err != nil {
		return E.Left[T](fmt.Errorf("Error getting parsing into %T: %w", val, err))
	}
	return E.Right[error](val)
}

func GetToJson[T any](client *resty.Client, url string) E.Either[error, T] {
	return E.MonadChain(
		E.MonadMap(
			get(client, url),
			GetResponseString,
		),
		FromJson[T],
	)
}

func GetResponseString(resp *resty.Response) string {
	return resp.String()
}

func GetGithubUrls(client *resty.Client) E.Either[error, GithubUrls] {
	return GetToJson[GithubUrls](client, "https://api.github.com")
}

func GetOrganizationInfo(client *resty.Client, organizationName string, githubUrls GithubUrls) E.Either[error, OrganizationInfo] {
	organizationUrl := githubUrls.OrganizationUrl
	organizationUrl = strings.Replace(organizationUrl, "{org}", organizationName, 1)
	return GetToJson[OrganizationInfo](client, organizationUrl)
}

func GetOrganizationRepos(client *resty.Client, organizationInfo OrganizationInfo) E.Either[error, []Repo] {
	return GetToJson[[]Repo](client, organizationInfo.ReposUrl)
}

func GetUserInfo(client *resty.Client, contributor Contributor) E.Either[error, User] {
	return GetToJson[User](client, contributor.Url)
}

func GetContributors(client *resty.Client, mostPopular Repo) E.Either[error, []Contributor] {
	return GetToJson[[]Contributor](client, mostPopular.ContributorsUrl)
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

func DoItFpStyle(organizationName string) E.Either[error, User] {
	client := resty.New()

	return F.Pipe6(
		GetGithubUrls(client),
		E.Chain(F.Curry3(GetOrganizationInfo)(client)(organizationName)),
		E.Chain(F.Curry2(GetOrganizationRepos)(client)),
		E.Map[error](GetMostPopularRepo),
		E.Chain(F.Curry2(GetContributors)(client)),
		E.Map[error](GetBiggestContributor),
		E.Chain(F.Curry2(GetUserInfo)(client)),
	)

}

func main() {
	fmt.Println(
		F.Pipe1(
			DoItFpStyle("golang"),
			E.Fold[error, User, string](
				func(err error) string {
					return fmt.Sprint("Error golang style:", err)
				},
				func(user User) string {
					return fmt.Sprint("The largest contributor to go is:", user.Name, "(", user.Login, ")")
				}),
		),
	)
}
