package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform provider for Appmixer
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("APPMIXER_API_URL", nil),
				Description: "The URL of the Appmixer API",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("APPMIXER_EMAIL", nil),
				Description: "The email used for authentication",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("APPMIXER_PASSWORD", nil),
				Description: "The password used for authentication",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"appmixer_user":    resourceUser(),
			"appmixer_account": resourceAccount(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"appmixer_user":        dataSourceUser(),
			"appmixer_users":       dataSourceUsers(),
			"appmixer_users_count": dataSourceUsersCount(),
			"appmixer_account":     dataSourceAccount(),
			"appmixer_accounts":    dataSourceAccounts(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// Client holds the authenticated client information
type Client struct {
	ApiURL     string
	Email      string
	UserID     string
	AuthToken  string
	Scope      []string // Add user scope to check for admin permissions
	HTTPClient *http.Client
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		IsActive bool     `json:"isActive"`
		Email    string   `json:"email"`
		Plan     string   `json:"plan"`
		Scope    []string `json:"scope"`
		Vendor   []string `json:"vendor"`
		Created  string   `json:"created"`
	} `json:"user"`
	Token string `json:"token"`
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiURL := d.Get("api_url").(string)
	email := d.Get("email").(string)
	password := d.Get("password").(string)

	client := &Client{
		ApiURL: apiURL,
		Email:  email,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Authenticate
	authReq := authRequest{
		Email:    email,
		Password: password,
	}

	authJSON, err := json.Marshal(authReq)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/user/auth", apiURL), bytes.NewBuffer(authJSON))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, diag.Errorf("authentication failed with status %d: %s", res.StatusCode, string(bodyBytes))
	}

	var authRes authResponse
	if err := json.NewDecoder(res.Body).Decode(&authRes); err != nil {
		return nil, diag.FromErr(err)
	}

	client.AuthToken = authRes.Token
	client.UserID = authRes.User.ID
	client.Scope = authRes.User.Scope

	// Check for admin scope for actions that require it
	var isAdmin bool
	for _, scope := range authRes.User.Scope {
		if scope == "admin" {
			isAdmin = true
			break
		}
	}

	// Warn if not admin but trying to use admin-only features
	if !isAdmin {
		tflog.Warn(ctx, "User does not have admin scope. Some operations will fail.", map[string]interface{}{
			"user_id": client.UserID,
		})
	}

	tflog.Info(ctx, "Successfully authenticated", map[string]interface{}{
		"user_id":  client.UserID,
		"is_admin": isAdmin,
	})

	return client, diags
}
