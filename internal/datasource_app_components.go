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

// Simplified representation of a component manifest for the data source
// We'll flatten parts of it or encode complex nested parts as JSON strings
type componentManifest struct {
	Name               string                 `json:"name"`
	Author             string                 `json:"author,omitempty"`
	Icon               string                 `json:"icon,omitempty"`
	Description        string                 `json:"description,omitempty"`
	Auth               map[string]interface{} `json:"auth,omitempty"`       // Simplified auth representation
	InPorts            json.RawMessage        `json:"inPorts,omitempty"`    // Use RawMessage for complex/variable structure
	OutPorts           json.RawMessage        `json:"outPorts,omitempty"`   // Use RawMessage
	Properties         json.RawMessage        `json:"properties,omitempty"` // Use RawMessage
	Webhook            bool                   `json:"webhook,omitempty"`
	WebhookAsync       bool                   `json:"webhookAsync,omitempty"`
	HttpRequestMethods []string               `json:"httpRequestMethods,omitempty"`
	State              map[string]interface{} `json:"state,omitempty"`
	Private            bool                   `json:"private,omitempty"`
	// Add other relevant fields if needed, potentially using json.RawMessage
}

func dataSourceAppComponents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppComponentsRead,
		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the app (e.g., appmixer.dropbox) whose components are to be retrieved.",
			},
			"components": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of components available for the specified app, including their manifest details.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"author": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"icon": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auth": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString}, // Simplified
						},
						"in_ports_json": { // Store complex parts as JSON strings
							Type:     schema.TypeString,
							Computed: true,
						},
						"out_ports_json": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"properties_json": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"webhook": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"webhook_async": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"http_request_methods": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"state": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString}, // Simplified
						},
						"private": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						// Add other fields as needed
					},
				},
			},
		},
	}
}

func dataSourceAppComponentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	appID := d.Get("app_id").(string)
	var diags diag.Diagnostics

	apiPath := "/apps/components"
	queryParams := url.Values{}
	queryParams.Add("app", appID)
	apiPath = fmt.Sprintf("%s?%s", apiPath, queryParams.Encode())

	tflog.Debug(ctx, "Reading Appmixer app components data source", map[string]interface{}{"app_id": appID, "path": apiPath})

	respBytes, err := client.DoRequest(ctx, "GET", apiPath, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list components for app %s: %w", appID, err))
	}

	var componentsData []componentManifest
	if err := json.Unmarshal(respBytes, &componentsData); err != nil {
		return diag.FromErr(fmt.Errorf("failed to parse components response for app %s: %w", appID, err))
	}

	components := make([]map[string]interface{}, len(componentsData))
	for i, comp := range componentsData {

		authMap := make(map[string]string)
		if comp.Auth != nil {
			for k, v := range comp.Auth {
				authMap[k] = fmt.Sprintf("%v", v) // Convert auth values to strings
			}
		}

		stateMap := make(map[string]string)
		if comp.State != nil {
			for k, v := range comp.State {
				stateMap[k] = fmt.Sprintf("%v", v) // Convert state values to strings
			}
		}

		compMap := map[string]interface{}{
			"name":                 comp.Name,
			"author":               comp.Author,
			"icon":                 comp.Icon,
			"description":          comp.Description,
			"auth":                 authMap,
			"in_ports_json":        string(comp.InPorts),
			"out_ports_json":       string(comp.OutPorts),
			"properties_json":      string(comp.Properties),
			"webhook":              comp.Webhook,
			"webhook_async":        comp.WebhookAsync,
			"http_request_methods": comp.HttpRequestMethods,
			"state":                stateMap,
			"private":              comp.Private,
		}
		components[i] = compMap
	}

	if err := d.Set("components", components); err != nil {
		return diag.FromErr(err)
	}

	// Set an ID based on the app_id
	d.SetId(fmt.Sprintf("app-%s-components", appID))

	return diags
}
