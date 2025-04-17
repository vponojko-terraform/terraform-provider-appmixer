package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Represents the structure of an account from the GET /accounts/:accountId API
type accountResponse struct {
	AccountID   string                 `json:"accountId"`
	Name        string                 `json:"name"`
	DisplayName *string                `json:"displayName"` // Use pointer for nullability
	Service     string                 `json:"service"`
	UserID      string                 `json:"userId"`
	ProfileInfo map[string]interface{} `json:"profileInfo"`
	Icon        string                 `json:"icon"`
	Label       string                 `json:"label"`
}

// Represents the structure for creating an account via POST /accounts
type createAccountRequest struct {
	Service     string            `json:"service"`
	Token       map[string]string `json:"token"`                 // API expects string values in the token map
	DisplayName *string           `json:"displayName,omitempty"` // Add optional display name
}

// Represents the response from POST /accounts
type createAccountResponse struct {
	AccountID string `json:"accountId"`
	TokenID   string `json:"tokenId"`
}

// Represents the structure for updating an account via PUT /accounts/:accountId
type updateAccountRequest struct {
	DisplayName string `json:"displayName"`
}

func resourceAccount() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAccountCreate,
		ReadContext:   resourceAccountRead,
		UpdateContext: resourceAccountUpdate,
		DeleteContext: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext, // Import using accountId
		},
		Schema: map[string]*schema.Schema{
			"service": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The service identifier for the account (e.g., 'appmixer:aws', 'appmixer:acme').",
			},
			"token": {
				Type:        schema.TypeMap,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
				Description: "A map containing the authentication credentials. Keys depend on the 'service' type (e.g., 'accessKeyId', 'secretKey' for AWS; 'username', 'password' for PWD). Values must be strings.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // Computed because the API might return it even if not set
				Description: "An optional user-friendly name for the account.",
			},
			// Computed fields read from the API
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The primary identifier/name returned by the API (e.g., username, email).",
			},
			"profile_info": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Additional profile details from the API.",
				Elem: &schema.Schema{
					Type: schema.TypeString, // Assuming profile info values are strings, adjust if needed
				},
			},
			"icon": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Base64 encoded icon for the service.",
			},
			"label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Label for the service type (e.g., 'Slack', 'Pipedrive').",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Appmixer user ID associated with this account.",
			},
		},
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	service := d.Get("service").(string)
	tokenInput := d.Get("token").(map[string]interface{})

	tflog.Info(ctx, "Creating new Appmixer account", map[string]interface{}{
		"service": service,
	})

	// Convert token map[string]interface{} to map[string]string for the API
	tokenMap := make(map[string]string)
	for k, v := range tokenInput {
		if s, ok := v.(string); ok {
			tokenMap[k] = s
		} else {
			return diag.Errorf("value for key '%s' in 'token' map is not a string", k)
		}
	}

	createReq := createAccountRequest{
		Service: service,
		Token:   tokenMap,
	}
	if displayName, ok := d.GetOk("display_name"); ok {
		displayNameStr := displayName.(string)
		createReq.DisplayName = &displayNameStr
	}

	respBytes, err := client.DoRequest(ctx, "POST", "/accounts", createReq)
	if err != nil {
		// Provide more context for common auth errors
		if strings.Contains(err.Error(), "Credentials validation failed") || strings.Contains(err.Error(), "Invalid credentials") {
			return diag.Errorf("Failed to create account for service '%s': Invalid credentials provided in the 'token' attribute. Please check the required keys and values for this service type. Original error: %v", service, err)
		}
		if strings.Contains(err.Error(), "missing") && strings.Contains(err.Error(), "required key") {
			return diag.Errorf("Failed to create account for service '%s': Missing required key in the 'token' attribute. Please check the required keys for this service type. Original error: %v", service, err)
		}
		return diag.FromErr(fmt.Errorf("failed to create account for service '%s': %w", service, err))
	}

	var createRes createAccountResponse
	if err := json.Unmarshal(respBytes, &createRes); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse create account response: %w", err))
	}

	if createRes.AccountID == "" {
		return diag.Errorf("API did not return an accountId after creating the account for service %s", service)
	}

	d.SetId(createRes.AccountID)
	tflog.Info(ctx, "Successfully received account ID from creation POST", map[string]interface{}{
		"account_id": createRes.AccountID,
		"service":    service,
	})

	// Set configured values in state immediately (except sensitive token)
	d.Set("service", service)
	if displayName, ok := d.GetOk("display_name"); ok {
		d.Set("display_name", displayName.(string))
	}

	// NOTE: We are *not* calling resourceAccountRead here.
	// Computed values (name, profile_info, etc.) will be populated on the next read.
	return nil // Return nil diagnostics for successful creation based on POST response
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	accountID := d.Id()
	var diags diag.Diagnostics

	tflog.Debug(ctx, "Reading Appmixer account", map[string]interface{}{
		"account_id": accountID,
	})

	respBytes, err := client.DoRequest(ctx, "GET", fmt.Sprintf("/accounts/%s", accountID), nil)
	if err != nil {
		// Handle 404 Not Found
		if strings.Contains(err.Error(), "status 404") {
			// Return a specific error for 404, allowing the caller (Create function) to handle retries.
			// For normal reads (plan/refresh), Terraform core will handle removing the resource if it gets this error.
			// No need to d.SetId("") here anymore.
			return diag.FromErr(fmt.Errorf("failed to read account %s: API returned status 404", accountID))
			// tflog.Warn(ctx, "Account not found, removing from state", map[string]interface{}{"account_id": accountID})
			// d.SetId("") // Mark resource for removal - Let Terraform core handle this based on the error.
			// return diags
		}
		return diag.FromErr(fmt.Errorf("failed to read account %s: %w", accountID, err))
	}

	var acc accountResponse
	if err := json.Unmarshal(respBytes, &acc); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse account response for %s: %w", accountID, err))
	}

	// Set computed fields
	d.Set("name", acc.Name)
	if acc.DisplayName != nil {
		d.Set("display_name", *acc.DisplayName)
	} else {
		if _, ok := d.GetOk("display_name"); !ok {
			d.Set("display_name", "")
		}
	}
	d.Set("service", acc.Service)
	d.Set("user_id", acc.UserID)

	// Convert profileInfo map[string]interface{} to map[string]string for Terraform state
	profileInfoMap := make(map[string]string)
	if acc.ProfileInfo != nil {
		for k, v := range acc.ProfileInfo {
			if s, ok := v.(string); ok {
				profileInfoMap[k] = s
			} else {
				// Handle cases where profile info might not be a simple string
				tflog.Warn(ctx, "Non-string value found in profile_info", map[string]interface{}{"key": k, "value": v})
				profileInfoMap[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	d.Set("profile_info", profileInfoMap)

	d.Set("icon", acc.Icon)
	d.Set("label", acc.Label)

	// Note: We don't set 'token' here as it's sensitive and write-only (ForceNew)

	return diags
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	accountID := d.Id()

	tflog.Info(ctx, "Updating Appmixer account", map[string]interface{}{
		"account_id": accountID,
	})

	// API only supports updating display_name
	if d.HasChange("display_name") {
		displayName := d.Get("display_name").(string)
		updateReq := updateAccountRequest{DisplayName: displayName}

		tflog.Debug(ctx, "Updating display name for account", map[string]interface{}{
			"account_id":   accountID,
			"display_name": displayName,
		})

		_, err := client.DoRequest(ctx, "PUT", fmt.Sprintf("/accounts/%s", accountID), updateReq)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to update display_name for account %s: %w", accountID, err))
		}
	} else {
		tflog.Debug(ctx, "No detectable changes requiring API update for account", map[string]interface{}{
			"account_id": accountID,
		})
		// Return early if no relevant attribute changed
		return resourceAccountRead(ctx, d, m)
	}

	// Read the updated resource
	return resourceAccountRead(ctx, d, m)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	accountID := d.Id()
	var diags diag.Diagnostics

	tflog.Info(ctx, "Deleting Appmixer account", map[string]interface{}{
		"account_id": accountID,
	})

	_, err := client.DoRequest(ctx, "DELETE", fmt.Sprintf("/accounts/%s", accountID), nil)
	if err != nil {
		// Check if already deleted (404) - Allow delete to succeed if already gone
		if strings.Contains(err.Error(), "status 404") {
			tflog.Warn(ctx, "Account already deleted", map[string]interface{}{"account_id": accountID})
			d.SetId("") // Ensure resource is removed from state
			return diags
		}
		return diag.FromErr(fmt.Errorf("failed to delete account %s: %w", accountID, err))
	}

	// Success
	d.SetId("")
	tflog.Info(ctx, "Successfully deleted account", map[string]interface{}{"account_id": accountID})
	return diags
}
