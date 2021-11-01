## gha-token: GitHub App Token Generator

Small tool to generate either GitHub App JWT or installation tokens as described in
[Authenticating with GitHub Apps](https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/).

The goal of this tool is to leverage GitHub App identity and permissions to
interact with GitHub repositories and API. The tool does not require any
webhook endpoint - just a GitHub App created in Settings. In order to generate
installation tokens the App also needs to be installed in one or more repositories.

For more information about GitHub App check out [the documentation](https://developer.github.com/apps/about-apps/).

## TL;DR - In Action

Assuming the GitHub App has ID `12345` and you saved the generated key in `key.pem`,
and you installed this App in your repository `me/myrepo` (where `me` is the name
of the user or organization), you can now do the following:

To clone the repo using GitHub App identity:

```
TOKEN=$(gha-token -a 12345 -k key.pem -r me/myrepo)
git clone https://x-access-token:${TOKEN}@github.com/me/myrepo.git
```

To get the list of issues for your repository using GitHub API:

```
TOKEN=$(gha-token -a 12345 -k key.pem -r me/myrepo)
curl -i -H "Authorization: token ${TOKEN}" https://api.github.com/repos/me/myrepo/issues
```

## Releases

Looking for pre-built binaries? You can find them on the [Releases](https://github.com/slawekzachcial/gha-token/releases) page. Currently 64-bit Linux and MacOS are available.

## Generating JWT Tokens

JWT tokens allow to interact with GitHub API `/app` endpoint. To generate them
you need the App ID and private key file in PEM format:

```
./gha-token --appId 12345 --keyPath path/to/key.pem
```

IMPORTANT: Generated JWT token expires after 10 minutes.

## Generating Installation Tokens

Installation Tokens can be used to interact with `/installation` endpoint.
Depending on the permissions of the App, these tokens also allow to interact
with Git repositories and GitHub APIs.

To generate an installation token you will either need the Installation ID or
Git repository owner and name.

To generate installation token based on installation ID (e.g. 98765):

```
./gha-token --appId 12345 --keyPath path/to/key.pem --installId 98765
```

To generate installation token based on repository owner and name (e.g. me/myrepo):

```
./gha-token --appId 12345 --keyPath path/to/key.pem --repo me/myrepo
```

Note that while this method is more convenient than using installation ID, its
implementation will invoke GitHub API multiple times in order to find the
corresponding installation ID and generate token for it.

IMPORTANT: Installation tokens expire after 1 hour.

## Available Command Line Flags

To get the list of flags simply run:

```
$> ./gha-token

Usage: gha-token [flags]

Flags:
  -g, --apiUrl string      GitHub API URL (default "https://api.github.com")
  -a, --appId string       Application ID as defined in app settings (Required)
  -i, --installId string   Installation ID of the application
  -k, --keyPath string     Path to key PEM file generated in app settings (Required)
  -r, --repo string        {owner/repo} of the GitHub repository
  -v, --verbose            Verbose stderr
```

## GitHub App Available Endpoints

The list of endpoints is available [here](https://developer.github.com/v3/apps/available-endpoints/).

## GitHub Enterprise

By default the API endpoint used is https://api.github.com. For GitHub Enterprise
you need to pass its URL as parameter, i.e. `--apiUrl https://github.my-company.com/api/v3`.

## Troubleshooting

Use `--verbose` to get more diagnostic information. Note that the output will contain
details about HTTP requests and responses, including tokens returned by GitHub.

## Building

Make sure your `GOPATH` is [properly set](https://github.com/golang/go/wiki/GOPATH).

Get the dependencies:

```
go get github.com/dgrijalva/jwt-go
go get github.com/spf13/pflag
go get -u golang.org/x/lint/golint
```

Build:

```
golint && go build
```

Build for multiple platforms:

```
mkdir -p $GOPATH/bin/{linux,darwin}/amd64
GOOS=linux GOARCH=amd64 go build -o $GOPATH/bin/linux/amd64/gha-token
GOOS=darwin GOARCH=amd64 go build -o $GOPATH/bin/darwin/amd64/gha-token
```
