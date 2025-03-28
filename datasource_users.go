package main

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter users (e.g., 'scope:admin')",
			},
			"pattern": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter users by pattern in username",
			},
			"sort": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sort users (e.g., 'created:-1')",
				Default:     "created:-1",
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Limit the number of users returned",
				Default:     30,
			},
			"offset": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Offset for pagination",
				Default:     0,
			},
			"users": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
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
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.Errorf("Listing all users requires admin permissions")
	}

	// Build query parameters
	queryParams := ""

	if filter, ok := d.GetOk("filter"); ok {
		queryParams += "filter=" + filter.(string) + "&"
	}

	if pattern, ok := d.GetOk("pattern"); ok {
		queryParams += "pattern=" + pattern.(string) + "&"
	}

	if sort, ok := d.GetOk("sort"); ok {
		queryParams += "sort=" + sort.(string) + "&"
	}

	limit := d.Get("limit").(int)
	offset := d.Get("offset").(int)
	queryParams += "limit=" + strconv.Itoa(limit) + "&offset=" + strconv.Itoa(offset)

	// Make the API request to list users
	resp, err := client.DoRequest(ctx, "GET", "/users?"+queryParams, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	var usersData []userResponse
	if err := json.Unmarshal(resp, &usersData); err != nil {
		return diag.FromErr(err)
	}

	// Transform the data into the terraform schema format
	users := make([]map[string]interface{}, len(usersData))
	for i, user := range usersData {
		userData := map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"email":     user.Email,
			"is_active": user.IsActive,
			"scope":     user.Scope,
			"created":   user.Created,
		}

		// Handle plan as a map
		if user.Plan != nil {
			var planMap map[string]interface{}
			if err := json.Unmarshal(user.Plan, &planMap); err == nil {
				userData["plan"] = planMap
			} else {
				// If it's not an object, try string
				var planStr string
				if err := json.Unmarshal(user.Plan, &planStr); err == nil {
					userData["plan"] = map[string]interface{}{"name": planStr}
				}
			}
		}

		users[i] = userData
	}

	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID for this data source
	d.SetId(strconv.FormatInt(int64(len(users)), 10) + "-users")

	return diags
}
