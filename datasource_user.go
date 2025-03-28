package main

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// userResponse is shared between datasource_user.go and resource_user.go
type userResponse struct {
	ID       string          `json:"id"`
	Username string          `json:"username"`
	IsActive bool            `json:"isActive"`
	Email    string          `json:"email"`
	Plan     json.RawMessage `json:"plan"`
	Scope    []string        `json:"scope"`
	Created  string          `json:"created"`
}

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"username": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	// Use the client to fetch current user information
	resp, err := client.DoRequest(ctx, "GET", "/user", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var userRes userResponse
	if err := json.Unmarshal(resp, &userRes); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(userRes.ID)
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

	return diags
}
