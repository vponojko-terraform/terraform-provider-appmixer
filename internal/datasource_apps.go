package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Represents the structure of a single app from the GET /apps response
// The API returns a map, where the key is the app name (e.g., "appmixer.asana")
// and the value is the app details.
type appResponse struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func dataSourceApps() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppsRead,
		Schema: map[string]*schema.Schema{
			// No input arguments needed for GET /apps
			"apps": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of available applications (services/modules).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique name/ID of the application (e.g., appmixer.asana).",
						},
						"label": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The user-friendly label for the application.",
						},
						"category": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The category of the application.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the application.",
						},
						"icon": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Base64 encoded icon for the application.",
						},
					},
				},
			},
		},
	}
}

func dataSourceAppsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	tflog.Debug(ctx, "Reading Appmixer apps data source")

	respBytes, err := client.DoRequest(ctx, "GET", "/apps", nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list apps: %w", err))
	}

	// The API returns a map[string]appResponse, not a list.
	var appsData map[string]appResponse
	if err := json.Unmarshal(respBytes, &appsData); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse apps response: %w", err))
	}

	// Convert the map to a list for Terraform
	apps := make([]map[string]interface{}, 0, len(appsData))
	for appName, appDetails := range appsData {
		appMap := map[string]interface{}{
			"name":        appName, // Use the map key as the name
			"label":       appDetails.Label,
			"category":    appDetails.Category,
			"description": appDetails.Description,
			"icon":        appDetails.Icon,
		}
		apps = append(apps, appMap)
	}

	if err := d.Set("apps", apps); err != nil {
		return diag.FromErr(err)
	}

	// Set a static ID for this data source, as it represents a collection
	d.SetId("appmixer-available-apps")

	return diags
}
