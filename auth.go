package nelly

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"

	"github.com/pharmatics/rest-util"
)

var errorHandler = func(w http.ResponseWriter, r *http.Request, err string) {
	status := restutil.NewFailureStatus(err, restutil.StatusReasonUnauthorized, nil)
	restutil.ResponseJSON(status, w, status.Code)
}

// Jwks is a set of keys which contains the public keys used to verify JWT issued
// by the authorization server and signed using the RS256 signing algorithm.
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

// JSONWebKeys is a JSON Web Key
type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(token *jwt.Token, jwksURL string) (string, error) {
	cert := ""
	resp, err := http.Get(jwksURL)

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Invlid token: can't find appropriate kid header claim")
		return cert, err
	}

	return cert, nil
}

// WithAuthSigningMethodHS256 handler authinticates requests with HS256 algorithm
func WithAuthSigningMethodHS256(secret string, audience string, issuer string) Handler {

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {

			// Verify 'aud' claim
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(audience, false)
			if !checkAud {
				return token, errors.New("Invalid audience")
			}
			// Verify 'iss' claim
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(issuer, false)
			if !checkIss {
				return token, errors.New("Invalid issuer")
			}

			return []byte(secret), nil
		},
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
		ErrorHandler:  errorHandler,
	})

	return withAuth(jwtMiddleware)
}

// WithAuthSigningMethodRS256 handler authinticates requests with RS256 algorithm
func WithAuthSigningMethodRS256(jwksEndpoint string, audience string, issuer string) Handler {

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {

			// Verify 'aud' claim
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(audience, false)
			if !checkAud {
				return token, errors.New("Invalid audience")
			}
			// Verify 'iss' claim
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(issuer, false)
			if !checkIss {
				return token, errors.New("Invalid issuer")
			}

			cert, err := getPemCert(token, jwksEndpoint)
			if err != nil {
				return nil, err
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
		ErrorHandler:  errorHandler,
	})

	return withAuth(jwtMiddleware)
}

func withAuth(jwtMiddleware *jwtmiddleware.JWTMiddleware) Handler {

	fn := func(h httprouter.Handle) httprouter.Handle {

		return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
			err := jwtMiddleware.CheckJWT(w, req)

			// If there was an error, do not continue.
			if err != nil {
				return
			}
			// Dispatch to the internal handler
			h(w, req, p)
		}
	}

	return fn

}
