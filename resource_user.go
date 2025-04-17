package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Helper function to check if user has admin permissions
func hasAdminPermissions(client *Client) bool {
	for _, scope := range client.Scope {
		if scope == "admin" {
			return true
		}
	}
	return false
}

// Extend the userResponse struct in resource_user.go to include vendor
// If not already defined in datasource_user.go
type userResponseExtended struct {
	userResponse
	Vendor []string `json:"vendor"`
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"plan": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"scope": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vendor": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

type createUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type createUserResponse struct {
	Token string `json:"token"`
}

type updateUserRequest struct {
	Scope  []string `json:"scope,omitempty"`
	Vendor []string `json:"vendor,omitempty"`
}

type deleteStatusResponse struct {
	Status     string `json:"status"`
	StepsDone  int    `json:"stepsDone"`
	StepsTotal int    `json:"stepsTotal"`
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	// Validate email format (should be email format per API docs)
	username := d.Get("username").(string)
	email := d.Get("email").(string)
	password := d.Get("password").(string)

	tflog.Info(ctx, "Creating new Appmixer user", map[string]interface{}{
		"username": username,
		"email":    email,
	})

	// Basic validation - API requires email format
	if len(password) < 5 {
		return diag.Errorf("Password must be at least 5 characters long according to Appmixer requirements")
	}

	// Create user request
	createReq := createUserRequest{
		Email:    email,
		Username: username,
		Password: password,
	}

	// Check if setting admin-only fields and user is not admin
	if d.Get("scope") != nil || d.Get("vendor") != nil {
		if !hasAdminPermissions(client) {
			return diag.Errorf("Setting scope or vendor requires admin permissions")
		}
	}

	// Make the API request to create a user
	resp, err := client.DoRequest(ctx, "POST", "/user", createReq)
	if err != nil {
		return diag.FromErr(err)
	}

	var createRes createUserResponse
	if err := json.Unmarshal(resp, &createRes); err != nil {
		return diag.FromErr(err)
	}

	// The Appmixer API doesn't return the user ID in the create response,
	// So we need to make a separate request to fetch the user by username
	// We'll use admin APIs for this
	usersResp, err := client.DoRequest(ctx, "GET", fmt.Sprintf("/users?pattern=%s", d.Get("username").(string)), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var usersRes []userResponse
	if err := json.Unmarshal(usersResp, &usersRes); err != nil {
		return diag.FromErr(err)
	}

	// Find the user we just created
	var userID string
	for _, user := range usersRes {
		if user.Username == d.Get("username").(string) {
			userID = user.ID
			break
		}
	}

	if userID == "" {
		return diag.Errorf("Failed to find newly created user with username %s", d.Get("username").(string))
	}

	d.SetId(userID)

	// Store the configured password in the state upon creation
	if err := d.Set("password", password); err != nil {
		return diag.FromErr(err)
	}

	// Handle scope and vendor settings if provided
	if d.Get("scope") != nil || d.Get("vendor") != nil {
		updateReq := updateUserRequest{}

		if scopes, ok := d.GetOk("scope"); ok {
			scopesList := scopes.([]interface{})
			scope := make([]string, len(scopesList))
			for i, v := range scopesList {
				scope[i] = v.(string)
			}
			updateReq.Scope = scope
		}

		if vendors, ok := d.GetOk("vendor"); ok {
			vendorsList := vendors.([]interface{})
			vendor := make([]string, len(vendorsList))
			for i, v := range vendorsList {
				vendor[i] = v.(string)
			}
			updateReq.Vendor = vendor
		}

		// Only update if we have something to update
		if len(updateReq.Scope) > 0 || len(updateReq.Vendor) > 0 {
			_, err := client.DoRequest(ctx, "PUT", fmt.Sprintf("/users/%s", userID), updateReq)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Set the computed fields
	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	userID := d.Id()
	tflog.Debug(ctx, "Reading Appmixer user", map[string]interface{}{
		"user_id": userID,
	})

	// Make the API request to get user details
	resp, err := client.DoRequest(ctx, "GET", fmt.Sprintf("/users/%s", userID), nil)
	if err != nil {
		// Check if this was a 404 Not Found error
		if resp != nil && len(resp) > 0 {
			var errorObj map[string]interface{}
			if jsonErr := json.Unmarshal(resp, &errorObj); jsonErr == nil {
				// Check for status code or error message indicating not found
				if status, ok := errorObj["statusCode"].(float64); ok && status == 404 {
					tflog.Warn(ctx, "User not found, removing from state", map[string]interface{}{
						"user_id": userID,
					})
					d.SetId("")
					return diags
				}
			}
		}
		return diag.FromErr(err)
	}

	var userRes userResponseExtended
	if err := json.Unmarshal(resp, &userRes); err != nil {
		return diag.FromErr(err)
	}

	d.Set("username", userRes.Username)
	d.Set("email", userRes.Email)
	d.Set("is_active", userRes.IsActive)

	// Handle plan as a map
	if userRes.Plan != nil {
		var planMap map[string]interface{}
		if err := json.Unmarshal(userRes.Plan, &planMap); err == nil {
			d.Set("plan", planMap)
		} else {
			// If it's not an object, try string
			var planStr string
			if err := json.Unmarshal(userRes.Plan, &planStr); err == nil {
				d.Set("plan", map[string]interface{}{"name": planStr})
			}
		}
	}

	d.Set("scope", userRes.Scope)
	d.Set("created", userRes.Created)

	// Always set vendor to handle empty arrays properly
	if userRes.Vendor != nil {
		d.Set("vendor", userRes.Vendor)
	} else {
		d.Set("vendor", []string{}) // Set empty array if vendor is nil
	}

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	userID := d.Id()

	tflog.Info(ctx, "Updating Appmixer user", map[string]interface{}{
		"user_id":                userID,
		"is_self":                userID == client.UserID,
		"scope_changed":          d.HasChange("scope"),
		"vendor_changed":         d.HasChange("vendor"),
		"password_changed_in_config": d.HasChange("password"),
	})

	// Determine if password update should be triggered based ONLY on HasChange
	shouldUpdatePassword := d.HasChange("password")
	var newPasswordValue string // Still needed if shouldUpdatePassword is true
	if shouldUpdatePassword {
		_, newPassword := d.GetChange("password")
		newPasswordValue = newPassword.(string)
		tflog.Debug(ctx, "Password update triggered by d.HasChange("password") returning true.")
	} else {
		tflog.Debug(ctx, "d.HasChange("password") returned false. No password update triggered.")
	}

	// Prevent modifying your own permissions (scope/vendor)
	if userID == client.UserID && (d.HasChange("scope") || d.HasChange("vendor")) {
		return diag.Errorf("Modifying your own permissions is not allowed for security reasons")
	}

	// When modifying other users, check if current user has admin permissions
	// This check needs to cover scope, vendor, AND intended password changes for other users
	if userID != client.UserID && (d.HasChange("scope") || d.HasChange("vendor") || shouldUpdatePassword) {
		if !hasAdminPermissions(client) {
			return diag.Errorf("Modifying other users (scope, vendor, or password) requires admin permissions")
		}
	}

	// Handle Scope and Vendor Updates
	updateReq := updateUserRequest{}
	needsScopeVendorUpdate := false

	if d.HasChange("scope") {
		scopes := d.Get("scope").([]interface{})
		scope := make([]string, len(scopes))
		for i, v := range scopes {
			scope[i] = v.(string)
		}
		updateReq.Scope = scope
		needsScopeVendorUpdate = true
	}

	if d.HasChange("vendor") {
		vendors := d.Get("vendor").([]interface{})
		vendor := make([]string, len(vendors))
		for i, v := range vendors {
			vendor[i] = v.(string)
		}
		updateReq.Vendor = vendor
		needsScopeVendorUpdate = true
	}

	// Only make Scope/Vendor update request if there are fields to update
	if needsScopeVendorUpdate {
		// Make the API request to update the user scope/vendor
		_, err := client.DoRequest(ctx, "PUT", fmt.Sprintf("/users/%s", userID), updateReq)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Update password if HasChange was true
	if shouldUpdatePassword {
		// Block updating own password
		if userID == client.UserID {
			return diag.Errorf("Cannot update your own password through Terraform. Use the Appmixer UI or API directly")
		} else {
			// Use admin reset-password for other users
			
			// Re-check length constraint on the new password value
			if len(newPasswordValue) < 5 {
				return diag.Errorf("Password must be at least 5 characters long according to Appmixer requirements")
			}
			
			passwordChangeReq := struct {
				Email    string `json:"email"`
				Password string `json:"password"`
			}{
				Email:    d.Get("email").(string), // Need the user's email for reset API
				Password: newPasswordValue,      // Use the stored new password value from config change
			}

			tflog.Debug(ctx, "Attempting to reset password for user via admin API", map[string]interface{}{"user_email": passwordChangeReq.Email})
			_, err := client.DoRequest(ctx, "POST", "/user/reset-password", passwordChangeReq)
			if err != nil {
				return diag.FromErr(err)
			}
			tflog.Info(ctx, "Password successfully reset for user via admin API", map[string]interface{}{"user_email": passwordChangeReq.Email})
			
			// Store the newly set password in the state after successful reset
			if err := d.Set("password", newPasswordValue); err != nil { // Use the stored new password value
				return diag.FromErr(err)
			}
		}
	}

	// Read final state after all potential updates
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	userID := d.Id()
	tflog.Info(ctx, "Deleting Appmixer user", map[string]interface{}{
		"user_id": userID,
	})

	// Check if trying to delete your own account
	if userID == client.UserID {
		return diag.Errorf("Deleting your own account through Terraform is not allowed for security reasons")
	}

	// Check if user has admin permissions
	if !hasAdminPermissions(client) {
		return diag.Errorf("Deleting users requires admin permissions")
	}

	// Make the API request to delete the user
	resp, err := client.DoRequest(ctx, "DELETE", fmt.Sprintf("/users/%s", userID), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse the ticket response
	var ticketRes struct {
		Ticket string `json:"ticket"`
	}
	if err := json.Unmarshal(resp, &ticketRes); err != nil {
		return diag.FromErr(err)
	}

	// Poll the delete status until completed or failed
	ticket := ticketRes.Ticket
	maxRetries := 30
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		statusURL := fmt.Sprintf("/users/%s/delete-status/%s", userID, ticket)
		resp, err := client.DoRequest(ctx, "GET", statusURL, nil)
		if err != nil {
			return diag.FromErr(err)
		}

		var statusRes deleteStatusResponse
		if err := json.Unmarshal(resp, &statusRes); err != nil {
			return diag.FromErr(err)
		}

		if statusRes.Status == "completed" {
			break
		}

		if statusRes.Status == "failed" || statusRes.Status == "cancelled" {
			return diag.Errorf("User deletion failed with status: %s", statusRes.Status)
		}

		// Wait before checking again
		time.Sleep(retryDelay)
	}

	d.SetId("")
	return diags
}
