package oauth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/sirupsen/logrus"
)

type Authorization struct {
	UserID   string `json:"u"`
	ClientID string `json:"c"`
}

func New() *server.Server {
	manager := manage.NewDefaultManager()

	// token memory store
	manager.MustTokenStorage(store.NewFileTokenStore("./certs/tokens.db"))

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("test", &models.Client{
		ID:     "test",
		Secret: "test",
		Domain: "https://oauth-redirect.googleusercontent.com/r/stampzilla-f92ac",
	})
	clientStore.Set("postman", &models.Client{
		ID:     "postman",
		Secret: "postman",
		Domain: "https://oauth.pstmn.io/v1/callback",
	})
	clientStore.Set("debug", &models.Client{
		ID:     "debug",
		Secret: "debug",
		Domain: "https://oauthdebugger.com/debug",
	})
	clientStore.Set("app", &models.Client{
		ID:     "app",
		Secret: "app",
		Domain: "app",
	})
	manager.MapClientStorage(clientStore)
	manager.SetValidateURIHandler(func(baseURI, redirectURI string) error {
		base, err := url.Parse(baseURI)
		if err != nil {
			return err
		}

		redirect, err := url.Parse(redirectURI)
		if err != nil {
			return err
		}

		if baseURI == "app" {
			if redirectURI == "stampzilla://redirect" {
				return nil
			}
			if strings.HasPrefix(redirectURI, "exp://") && strings.HasSuffix(redirect.Host, ":19000") {
				return nil
			}
			if redirect.Hostname() == "localhost" {
				return nil
			}
			if redirectURI == "https://auth.expo.io/@stamp/stampzilla-app" {
				return nil
			}
			if redirect.Host == "exp://exp.host/@stamp/stampzilla-app" {
				return nil
			}
			spew.Dump(redirect)
			spew.Dump(redirect.Hostname())
			spew.Dump(strings.HasPrefix(redirectURI, "exp://"), strings.HasSuffix(redirect.Host, ":19000"))
			return errors.ErrInvalidRedirectURI
		}

		if !strings.HasSuffix(redirect.Host, base.Host) {
			return errors.ErrInvalidRedirectURI
		}
		return nil
	})

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		logrus.Info("OAuth 2.0 Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		logrus.Info("Response Error:", re.Error)
	})

	return srv
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	auth := r.Context().Value("authorization")
	if auth == nil {
		return "", fmt.Errorf("authorization is missing")
	}

	return string(auth.([]uint8)), nil
}
