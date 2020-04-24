package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fatih/color"

	"github.com/coreos/go-oidc"
	protocol "github.com/gjbae1212/grpc-vpn/grpc/go"
	"github.com/gjbae1212/grpc-vpn/internal"
	"github.com/pkg/browser"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

// Only Support Redirect URL http://localhost:10000/code
const (
	googleOpenIDProvider = "https://accounts.google.com"
	redirectServer       = "http://localhost:10000"
	redirectPath         = "/code"
	redirectPort         = "10000"
	redirectURL          = "http://localhost:10000/code"
)

// https://developers.google.com/identity/protocols/oauth2/native-app
// https://developers.google.com/identity/protocols/oauth2/openid-connect
type GoogleOpenIDConfig struct {
	ClientId     string // google client id
	ClientSecret string // google secret

	HD          string   // gsuite domain (only vpn-server)
	AllowEmails []string // allow emails (only vpn-server)
}

// unaryServerInterceptor returns new unary server interceptor that checks an authorization with google openID.
func (c *GoogleOpenIDConfig) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		auth, ok := req.(*protocol.AuthRequest)
		if !ok {
			return handler(ctx, req)
		}
		if auth.AuthType != protocol.AuthType_AT_GOOGLE_OPEN_ID {
			return handler(ctx, req)
		}

		if auth.GoogleOpenId.Code == "" {
			return nil, internal.ErrorUnauthorized
		}

		clientID := c.ClientId
		clientSecret := c.ClientSecret
		hd := c.HD
		allowEmails := c.AllowEmails

		provider, err := oidc.NewProvider(context.Background(), googleOpenIDProvider)
		if err != nil {
			return nil, internal.ErrorUnknown
		}

		oauthConf := oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}

		exCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		token, err := oauthConf.Exchange(exCtx, auth.GoogleOpenId.Code)
		if err != nil {
			return nil, internal.ErrorUnauthorized
		}

		expiresIn := token.Extra("expires_in").(float64)
		if expiresIn <= 10 { // Fail when remaining lifetime is less than 10 seconds.
			return nil, internal.ErrorUnauthorized
		}

		rawIdToken, ok := token.Extra("id_token").(string)
		if !ok {
			return nil, internal.ErrorUnauthorized
		}

		idToken, err := provider.Verifier(&oidc.Config{ClientID: clientID}).Verify(ctx, rawIdToken)
		if err != nil {
			return nil, internal.ErrorUnauthorized
		}

		claims := map[string]interface{}{}
		if err := idToken.Claims(&claims); err != nil {
			return nil, internal.ErrorUnauthorized
		}

		// check gsuite domain(matched)
		// if hd is empty, don't check hd
		if hd != "" && hd != claims["hd"].(string) {
			return nil, internal.ErrorUnauthorized
		}

		// check email
		// if allowEmails is empty, don't check email
		if len(allowEmails) != 0 {
			if !internal.IsMatchedStringFromSlice(claims["email"].(string), allowEmails) {
				return nil, internal.ErrorUnauthorized
			}
		}

		// inject user
		newCtx := context.WithValue(ctx, UserCtxName, claims["email"].(string))
		return handler(newCtx, req)
	}
}

// ClientAuthMethod returns auth method for client.
func (c *GoogleOpenIDConfig) clientAuthMethod() ClientAuthMethod {
	return func(conn protocol.VPNClient) (jwt string, err error) {
		if conn == nil {
			return "", errors.Wrapf(internal.ErrorInvalidParams, "Google OpenID ClientAuthMethod")
		}

		// extract information for oauth2
		clientID := c.ClientId
		clientSecret := c.ClientSecret
		if clientID == "" || clientSecret == "" {
			return "", errors.Wrapf(internal.ErrorInvalidParams, "Google OpenID ClientAuthMethod")
		}

		provider, err := oidc.NewProvider(context.Background(), googleOpenIDProvider)
		if err != nil {
			return "", errors.Wrapf(err, "Google OpenID ClientAuthMethod")
		}

		// start http server
		serverCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var code string
		mux := http.NewServeMux()
		server := http.Server{Addr: fmt.Sprintf(":%s", redirectPort), Handler: mux}
		mux.HandleFunc(redirectPath, func(w http.ResponseWriter, req *http.Request) {
			// extract code
			if req.URL.Query()["code"] != nil && len(req.URL.Query()["code"]) > 0 {
				code = req.URL.Query()["code"][0]
				http.Redirect(w, req, redirectServer+"/success", 301)
			} else {
				http.Redirect(w, req, redirectServer+"/fail", 301)
			}
		})
		mux.HandleFunc("/success", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("[REDIRECT] Close this page."))
			cancel()
		})
		mux.HandleFunc("/fail", func(w http.ResponseWriter, req *http.Request) {
			w.Write([]byte("[REDIRECT] Close this page."))
			cancel()
		})

		go func() {
			server.ListenAndServe()
		}()

		state := internal.GenerateRandomString(10)
		conf := oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		}
		if err := browser.OpenURL(conf.AuthCodeURL(state)); err != nil {
			return "", errors.Wrapf(err, "Google OpenID ClientAuthMethod")
		}
		fmt.Println(color.GreenString("[WAIT] YOUR GOOGLE OPENID AUTHENTICATION"))
		select {
		case <-serverCtx.Done():
			server.Shutdown(serverCtx)
		}
		if code == "" {
			return "", errors.Wrapf(internal.ErrorUnauthorized, "Google OpenID ClientAuthMethod")
		}

		// call authentication request to VPN server.
		authCtx, authCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer authCancel()
		response, err := conn.Auth(authCtx, &protocol.AuthRequest{
			AuthType: protocol.AuthType_AT_GOOGLE_OPEN_ID,
			GoogleOpenId: &protocol.AuthRequest_GoogleOpenID{
				Code: code,
			},
		})
		if err != nil {
			return "", errors.Wrapf(internal.ErrorUnauthorized, "Google OpenID ClientAuthMethod")
		}
		if response.ErrorCode != protocol.ErrorCode_EC_SUCCESS || response.Jwt == "" {
			return "", errors.Wrapf(internal.ErrorUnauthorized, "Google OpenID ClientAuthMethod")
		}

		return response.Jwt, nil
	}
}
