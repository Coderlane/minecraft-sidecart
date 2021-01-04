package auth

const defaultOAuthJSON = `{
  "installed": {
    "client_id": "test_client_id",
    "project_id": "test_project_id",
    "auth_uri": "http://localhost:9000/auth",
    "token_uri": "http://localhost:9000/token",
    "client_secret": "test_client_secret",
    "redirect_uris": [
      "urn:ietf:wg:oauth:2.0:oob"
    ]
  }
}
`

const defaultIDPJSON = `{
  "api_key": "test_secret"
}
`
