package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// In-memory store for demonstration (not persistent)
var things = make(map[string]string)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return Provider()
		},
	})
}

// Provider defines the Terraform provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"firebase_default_bucket": resourceDefaultBucket(),
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			var client *FirebaseClient
			var err error

			if creds, ok := d.GetOk("credentials"); ok {
				// Use explicit credentials if provided
				client, err = NewFirebaseClient(creds.(string))
			} else {
				// Fall back to ADC
				client, err = NewFirebaseClient("")
			}

			if err != nil {
				return nil, diag.FromErr(err)
			}
			return client, nil
		},
	}
}
