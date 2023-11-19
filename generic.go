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
	"fmt"
	"strings"
)

// Import resty into your code and refer it as `resty`.
import "github.com/go-resty/resty/v2"

import E "github.com/IBM/fp-go/either"
import F "github.com/IBM/fp-go/function"
import J "github.com/IBM/fp-go/json"
import A "github.com/IBM/fp-go/array"

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

func Get(client *resty.Client, url string) E.Either[error, *resty.Response] {
	return E.TryCatchError(
		client.R().EnableTrace().Get(url),
	)
}

func GetToJson[T any](client *resty.Client, url string) E.Either[error, T] {
	return F.Pipe2(
		Get(client, url),
		E.Map[error](GetResponseBody),
		E.Chain(J.Unmarshal[T]),
	)
}

func GetResponseBody(resp *resty.Response) []byte {
	return resp.Body()
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
	return F.Pipe1(repos, A.Reduce(func(acc, repo Repo) Repo {
		if acc.WatchersCount > repo.WatchersCount {
			return acc
		}
		return repo
	}, repos[0]))
}

func GetBiggestContributor(contributors []Contributor) (biggestContributor Contributor) {
	return F.Pipe1(contributors, A.Reduce(func(acc, contributor Contributor) Contributor {
		if acc.Contributions > contributor.Contributions {
			return acc
		}
		return contributor
	}, contributors[0]))
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
			E.Fold(
				func(err error) string {
					return fmt.Sprint("Error golang style:", err)
				},
				func(user User) string {
					return fmt.Sprint("The largest contributor to go is:", user.Name, "(", user.Login, ")")
				}),
		),
	)
}
