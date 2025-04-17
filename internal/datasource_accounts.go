package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Note: accountResponse struct is already defined in resource_account.go
// We reuse it here as the structure from GET /accounts is a list of these.

func dataSourceAccounts() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAccountsRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter accounts based on Appmixer API filter syntax (e.g., 'service:!appmixer:aws').",
			},
			"accounts": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of accounts matching the filter.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"profile_info": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"icon": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAccountsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	apiPath := "/accounts"
	queryParams := url.Values{}

	if filterVal, ok := d.GetOk("filter"); ok {
		queryParams.Add("filter", filterVal.(string))
	}

	if len(queryParams) > 0 {
		apiPath = fmt.Sprintf("%s?%s", apiPath, queryParams.Encode())
	}

	tflog.Debug(ctx, "Listing Appmixer accounts data source", map[string]interface{}{
		"filter": d.Get("filter").(string),
		"path":   apiPath,
	})

	respBytes, err := client.DoRequest(ctx, "GET", apiPath, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list accounts: %w", err))
	}

	var accountsData []accountResponse
	if err := json.Unmarshal(respBytes, &accountsData); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse accounts list response: %w", err))
	}

	accounts := make([]map[string]interface{}, len(accountsData))
	for i, acc := range accountsData {
		profileInfoMap := make(map[string]string)
		if acc.ProfileInfo != nil {
			for k, v := range acc.ProfileInfo {
				if s, ok := v.(string); ok {
					profileInfoMap[k] = s
				} else {
					profileInfoMap[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		accMap := map[string]interface{}{
			"account_id":   acc.AccountID,
			"service":      acc.Service,
			"name":         acc.Name,
			"profile_info": profileInfoMap,
			"icon":         acc.Icon,
			"label":        acc.Label,
			"user_id":      acc.UserID,
		}
		if acc.DisplayName != nil {
			accMap["display_name"] = *acc.DisplayName
		} else {
			accMap["display_name"] = ""
		}
		accounts[i] = accMap
	}

	if err := d.Set("accounts", accounts); err != nil {
		return diag.FromErr(err)
	}

	// Generate a stable ID based on the filter used (or lack thereof)
	filterStr := d.Get("filter").(string)
	if filterStr == "" {
		filterStr = "all"
	}
	d.SetId(fmt.Sprintf("accounts-%s-%d", filterStr, len(accounts)))

	return diags
}
