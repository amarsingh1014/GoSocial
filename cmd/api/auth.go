package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"social/internal/mailer"
	"social/internal/store"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserWihToken struct {
	*store.User `json:"user"`
	Token       string `json:"token"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Username string `json:"username" validate:"required,min=3,max=40"`
}

// RegisterUser godoc
//
//	@Summary		Register a new user
//	@Description	Register a new user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterUserPayload	true	"Register User"
//	@Success		201		{object}	UserWihToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	payload := RegisterUserPayload{}

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	//hash the password
	//store the user
	log.Println("Before creating user")

	user := &store.User{
		Email:    payload.Email,
		Username: payload.Username,
		RoleID: 1,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	ctx := r.Context()

	plainToken := uuid.New().String()

	//store
	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)

	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestError(w, r, "email already exists")
		case store.ErrDuplicateUsername:
			app.badRequestError(w, r, "username already exists")
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	userWithToken := UserWihToken{
		User:  user,
		Token: plainToken,
	}

	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontendURL, plainToken)

	isProduction := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activationURL,
	}

	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProduction)

	if err != nil {
		app.logger.Errorw("Error sending email", "error", err)

		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("Error deleting user", "error", err)
		}

		app.internalServerError(w, r, err.Error())
		return
	}

	app.logger.Infow("Email sent", "status code", status)

	if err := writeJSON(w, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// createTokenHandler godoc
//
//	@Summary		Create a new token
//	@Description	Create a new token
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User Credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	//parse the credentials
	payload := CreateUserTokenPayload{}

	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestError(w, r, err.Error())
		return
	}

	//check if the user exists and match the password
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedError(w, r, "invalid credentials")
		default:
			app.internalServerError(w, r, err.Error())
		}
		return
	}

	err = user.Password.Matches(payload.Password)
	if err != nil {
		app.unauthorizedError(w, r, "invalid credentials")
		return
	}

	//generate a token -> add claims
	claims := jwt.MapClaims{
		"sub": (int)(user.ID),
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"aud": app.config.auth.token.aud,
		"iss": app.config.auth.token.iss,
	}

	token, err := app.authenticator.GenerateToken(claims)

	if err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}

	//return the token
	if err := writeJSON(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err.Error())
		return
	}
}
