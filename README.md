# gha-token: GitHub App Token Generator

Small tool to generate either GitHub App JWT or installation tokens as described in
[Authenticating with GitHub Apps](https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/).

The goal of this tool is to leverage GitHub App identity and permissions to
interact with GitHub repositories and API. The tool does not require any
webhook endpoint - just a GitHub App created in Settings. In order to generate
installation tokens the App also need to be installed.

For more information about GitHub App check out [the documentation](https://developer.github.com/apps/about-apps/).

## Generating JWT Tokens

JWT tokens allow to interact with GitHub API `/app` endpoint. To generate them
you need the App ID and private key file in PEM format:

```
./gha-token -app 12345 -keyPath path/to/key.pem
```

## Generating Installation Tokens

Installation Tokens can be used to interact with `/installation` endpoint.
Depending on the permissions of the App, these tokens also allow to interact
with Git repositories and GitHub APIs.

To generate an installation token you will either need the Installation ID or
Git repository owner and name.

To generate installation token based on installation ID (e.g. 98765):

```
./gha-token -app 12345 -keyPath path/to/key.pem -inst 98765
```

To generate installation token based on repository owner and name (e.g. me/myrepo):

```
./gha-token -app 12345 -keyPath path/to/key.pem -repo me/myrepo
```

Note that while this method is more convenient than using installation ID, its
implementation will invoke GitHub API multiple time in order to find the
corresponding installation ID and generate token for it.

## GitHub Enterprise

By default the API endpoint used is https://api.github.com. For GitHub Enterprise
you need to pass its URL as parameter, i.e. `-apiUrl https://github.my-company.com/api/v3`.

## Troubleshooting

Use `-v` or `-vv` parameters to get more diagnostic information. Note that `-vv`
will print HTTP requests and responses, including tokens returned by GitHub.
