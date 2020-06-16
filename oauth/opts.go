//
// Copyright (c) 2020 SSH Communications Security Inc.
//
// All rights reserved.
//

package oauth

import (
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

// Option is configuration applied to the client
type Option func(*tAuth) *tAuth

// Access setups client access key
func Access(access *string) Option {
	return func(auth *tAuth) *tAuth {
		if access != nil {
			auth.access = *access
		}
		return auth
	}
}

// Secret setups clients secret key
func Secret(secret *string) Option {
	return func(auth *tAuth) *tAuth {
		if secret != nil {
			auth.secret = *secret
		}
		return auth
	}
}

// Digest setups client secret digest
func Digest(oauthAccess, oauthSecret *string) Option {
	return func(auth *tAuth) *tAuth {
		if oauthAccess != nil && oauthSecret != nil {
			auth.digest = base64.StdEncoding.EncodeToString([]byte(*oauthAccess + ":" + *oauthSecret))
		}
		return auth
	}
}

// UseConfigFile setup credential from tol file
func UseConfigFile(path *string) Option {
	return func(auth *tAuth) *tAuth {
		type config struct {
			AuthClientID     string `toml:"oauth_client_id"`
			AuthClientSecret string `toml:"oauth_client_secret"`
			ClientID         string `toml:"api_client_id"`
			ClientSecret     string `toml:"api_client_secret"`
		}
		var file struct {
			Auth config
		}

		if path == nil {
			return auth
		}

		f, err := os.Open(*path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}

		err = toml.Unmarshal(data, &file)
		if err != nil {
			return auth
		}

		if file.Auth.ClientID != "" {
			auth.access = file.Auth.ClientID
		}

		if file.Auth.ClientSecret != "" {
			auth.secret = file.Auth.ClientSecret
		}

		if file.Auth.AuthClientID != "" && file.Auth.AuthClientSecret != "" {
			auth = Digest(&file.Auth.AuthClientID, &file.Auth.AuthClientSecret)(auth)
		}

		return auth
	}
}

// UseEnvironment setup credential from environment variables
func UseEnvironment() Option {
	return func(auth *tAuth) *tAuth {
		if access, ok := os.LookupEnv("PRIVX_API_CLIENT_ID"); ok {
			auth.access = access
		}

		if secret, ok := os.LookupEnv("PRIVX_API_CLIENT_SECRET"); ok {
			auth.secret = secret
		}

		if authAccess, ok := os.LookupEnv("PRIVX_API_OAUTH_CLIENT_ID"); ok {
			if authSecret, ok := os.LookupEnv("PRIVX_API_OAUTH_CLIENT_SECRET"); ok {
				auth = Digest(&authAccess, &authSecret)(auth)
			}
		}

		return auth
	}
}
