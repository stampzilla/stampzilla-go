package oauth

import (
	"encoding/json"
	"net/http"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"github.com/sirupsen/logrus"
)

type Authorization struct {
	UserID   string `json:"u"`
	Instance string `json:"i"`
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
	manager.MapClientStorage(clientStore)

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
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		return
	}

	instance, ok := store.Get("instance")
	uid, ok2 := store.Get("user_id")
	if !ok || !ok2 {
		// Store authorization parameters in the session
		if r.Form == nil {
			r.ParseForm()
		}

		store.Set("ReturnUri", r.Form)
		store.Save()

		// Redirect to the login page
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}

	/* LOGOUT the user each time
	store.Delete("user_id")
	store.Save()
	*/

	authorization := &Authorization{
		UserID:   uid.(string),
		Instance: instance.(string),
	}

	encoded, err := json.Marshal(authorization)

	return string(encoded), err
}
