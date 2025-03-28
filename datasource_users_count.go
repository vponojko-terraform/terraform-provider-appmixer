package main

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsersCount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersCountRead,
		Schema: map[string]*schema.Schema{
			"total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

type countResponse struct {
	Count int `json:"count"`
}

func dataSourceUsersCountRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	var diags diag.Diagnostics

	// Check if user has admin permissions
	hasAdminScope := false
	for _, scope := range client.Scope {
		if scope == "admin" {
			hasAdminScope = true
			break
		}
	}

	if !hasAdminScope {
		return diag.Errorf("Getting user count requires admin permissions")
	}

	// Make the API request to get the user count
	resp, err := client.DoRequest(ctx, "GET", "/users/count", nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var countData countResponse
	if err := json.Unmarshal(resp, &countData); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(int64(countData.Count), 10))
	d.Set("total", countData.Count)

	return diags
}
