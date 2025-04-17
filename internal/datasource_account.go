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

// Note: accountResponse struct is already defined in resource_account.go
// We assume both files are part of the same package `internal`.

func dataSourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountRead,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the account to retrieve.",
			},
			// Computed fields
			"service": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The service identifier for the account.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The user-friendly name for the account.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The primary identifier/name returned by the API.",
			},
			"profile_info": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Additional profile details from the API.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
				Description: "Label for the service type.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Appmixer user ID associated with this account.",
			},
		},
	}
}

func dataSourceAccountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	accountID := d.Get("account_id").(string)
	var diags diag.Diagnostics

	tflog.Debug(ctx, "Reading Appmixer account data source", map[string]interface{}{
		"account_id": accountID,
	})

	respBytes, err := client.DoRequest(ctx, "GET", fmt.Sprintf("/accounts/%s", accountID), nil)
	if err != nil {
		// Handle 404 Not Found gracefully for data sources
		if strings.Contains(err.Error(), "status 404") {
			return diag.Errorf("Account with ID %s not found", accountID)
		}
		return diag.FromErr(fmt.Errorf("failed to read account %s: %w", accountID, err))
	}

	var acc accountResponse
	if err := json.Unmarshal(respBytes, &acc); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse account response for %s: %w", accountID, err))
	}

	// Set the ID for the data source state (required)
	d.SetId(acc.AccountID)

	// Set computed fields
	d.Set("service", acc.Service)
	if acc.DisplayName != nil {
		d.Set("display_name", *acc.DisplayName)
	} else {
		d.Set("display_name", "") // Set empty string if null from API
	}
	d.Set("name", acc.Name)
	d.Set("user_id", acc.UserID)

	// Convert profileInfo map[string]interface{} to map[string]string
	profileInfoMap := make(map[string]string)
	if acc.ProfileInfo != nil {
		for k, v := range acc.ProfileInfo {
			if s, ok := v.(string); ok {
				profileInfoMap[k] = s
			} else {
				tflog.Warn(ctx, "Non-string value found in profile_info for data source", map[string]interface{}{"account_id": accountID, "key": k, "value": v})
				profileInfoMap[k] = fmt.Sprintf("%v", v)
			}
		}
	}
	d.Set("profile_info", profileInfoMap)

	d.Set("icon", acc.Icon)
	d.Set("label", acc.Label)

	return diags
}
