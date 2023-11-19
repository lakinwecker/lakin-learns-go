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

// import "github.com/jessevdk/go-flags"
// import "github.com/repeale/fp-go"
import either "github.com/IBM/fp-go/either"
import function "github.com/IBM/fp-go/function"

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

func get(client *resty.Client, url string) either.Either[error, *resty.Response] {
	resp, err := client.R().
		EnableTrace().
		Get(url)
	if err != nil {
		return either.Left[*resty.Response](fmt.Errorf("Error getting url(%s): %w", url, err))
	}
	return either.Right[error](resp)
}

func FromJson[T any](jsonString string) either.Either[error, T] {
	var val T
	err := json.Unmarshal([]byte(jsonString), &val)
	if err != nil {
		return either.Left[T](fmt.Errorf("Error getting parsing into %T: %w", val, err))
	}
	return either.Right[error](val)
}

func GetToJson[T any](client *resty.Client, url string) either.Either[error, T] {
	return either.MonadChain(
		either.MonadMap(
			get(client, url),
			GetResponseString,
		),
		FromJson[T],
	)
}

func GetResponseString(resp *resty.Response) string {
	return resp.String()
}

func GetGithubUrls(client *resty.Client) either.Either[error, GithubUrls] {
	return GetToJson[GithubUrls](client, "https://api.github.com")
}

func GetOrganizationInfo(client *resty.Client, organizationName string) func(GithubUrls) either.Either[error, OrganizationInfo] {
	return func(githubUrls GithubUrls) either.Either[error, OrganizationInfo] {
		organizationUrl := githubUrls.OrganizationUrl
		organizationUrl = strings.Replace(organizationUrl, "{org}", organizationName, 1)
		return GetToJson[OrganizationInfo](client, organizationUrl)
	}
}

func GetOrganizationRepos(client *resty.Client) func(OrganizationInfo) either.Either[error, []Repo] {
	return func(organizationInfo OrganizationInfo) either.Either[error, []Repo] {
		return GetToJson[[]Repo](client, organizationInfo.ReposUrl)
	}
}

func GetUserInfo(client *resty.Client) func(Contributor) either.Either[error, User] {
	return func(contributor Contributor) either.Either[error, User] {
		return GetToJson[User](client, contributor.Url)
	}
}

func GetContributors(client *resty.Client) func(Repo) either.Either[error, []Contributor] {
	return func(mostPopular Repo) either.Either[error, []Contributor] {
		return GetToJson[[]Contributor](client, mostPopular.ContributorsUrl)
	}
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

func DoItFpStyle(organizationName string) either.Either[error, User] {
	client := resty.New()

	return function.Pipe6(
		GetGithubUrls(resty.New()),
		either.Chain(GetOrganizationInfo(client, organizationName)),
		either.Chain(GetOrganizationRepos(client)),
		either.Map[error](GetMostPopularRepo),
		either.Chain(GetContributors(client)),
		either.Map[error](GetBiggestContributor),
		either.Chain(GetUserInfo(client)),
	)

}

func main() {
	fmt.Println(
		either.Fold[error, User, string](
			func(err error) string {
				return fmt.Sprint("Error golang style:", err)
			},
			func(user User) string {
				return fmt.Sprint("The largest contributor to go is:", user.Name, "(", user.Login, ")")
			})(DoItFpStyle("golang")),
	)
}
