package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-japanese-learning/internal/model"
)

type registerRequest struct {
	Email     string          `json:"email"`
	Username  string          `json:"username"`
	Password  string          `json:"password"`
	JLPTLevel model.JLPTLevel `json:"jlpt_level"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *Router) handleRegister(w http.ResponseWriter, req *http.Request) {
	var input registerRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Username = strings.TrimSpace(input.Username)

	if input.Email == "" || input.Username == "" || len(input.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email, username and password are required; password min length is 6"})
		return
	}
	if !model.IsValidJLPT(input.JLPTLevel) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid jlpt level"})
		return
	}

	result, err := r.authService.Register(req.Context(), input.Email, input.Username, input.Password, input.JLPTLevel)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (r *Router) handleLogin(w http.ResponseWriter, req *http.Request) {
	var input loginRequest
	if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	result, err := r.authService.Login(req.Context(), strings.TrimSpace(strings.ToLower(input.Email)), input.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (r *Router) handleLogout(w http.ResponseWriter, req *http.Request) {
	token := extractBearerToken(req.Header.Get("Authorization"))
	if err := r.authService.Logout(req.Context(), token); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (r *Router) handleMe(w http.ResponseWriter, req *http.Request) {
	user, err := currentUser(req.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, user)
}
