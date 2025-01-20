package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/c4po/terrastate/internal/models"
	"github.com/c4po/terrastate/internal/templates"
	"github.com/c4po/terrastate/internal/utils"
)

type LoginHandler struct{}

var (
	pendingTokens = make(map[string]*models.TokenRequest)
)

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) TerraformLogin(w http.ResponseWriter, r *http.Request) {
	code, err := utils.GenerateCode()
	if err != nil {
		http.Error(w, "Error generating code", http.StatusInternalServerError)
		return
	}

	pendingTokens[code] = &models.TokenRequest{
		Code:      code,
		CreatedAt: time.Now(),
	}

	redirectURL := fmt.Sprintf("https://%s/app/settings/tokens?source=terraform-login&code=%s",
		r.Host, code)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *LoginHandler) Tokens(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	data := struct {
		Code  string
		Token string
	}{
		Code:  code,
		Token: "",
	}

	tmpl, err := template.New("token").Parse(templates.TokenPageTemplate)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func (h *LoginHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	request, exists := pendingTokens[code]
	if !exists {
		http.Error(w, "Invalid code", http.StatusBadRequest)
		return
	}

	if time.Since(request.CreatedAt) > 15*time.Minute {
		delete(pendingTokens, code)
		http.Error(w, "Code has expired", http.StatusBadRequest)
		return
	}

	token, err := utils.GenerateToken()
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	tokens[token] = true
	delete(pendingTokens, code)

	data := struct {
		Code  string
		Token string
	}{
		Code:  code,
		Token: token,
	}

	tmpl, err := template.New("token").Parse(templates.TokenPageTemplate)
	if err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}
