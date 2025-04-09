package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDefaultBucket() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDefaultBucketCreate,
		ReadContext:   resourceDefaultBucketRead,
		DeleteContext: resourceDefaultBucketDelete,
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Firebase project ID.",
			},
			"bucket_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the default bucket.",
			},
			"location": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "us",
				ForceNew:    true,
				Description: "The location of the default bucket (e.g., 'us').",
			},
		},
	}
}

// resourceThingCreate handles resource creation
func resourceDefaultBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*FirebaseClient)
	project := d.Get("project").(string)
	bucketName := d.Get("bucket_name").(string)
	location := d.Get("location").(string)

	// Check if the default bucket already exists
	exists, bucketName, err := checkDefaultBucketExists(ctx, client, project)
	if err != nil {
		return diag.FromErr(err)
	}
	if exists {
		d.SetId(fmt.Sprintf("projects/%s/defaultBucket", project))
		d.Set("bucket_name", bucketName)
		return nil
	}

	// Set the ID to the name (in a real provider, this might be a unique ID from an API)
	d.SetId(fmt.Sprintf("projects/%s/defaultBucket", project))
	d.Set("bucket_name", bucketName)
	d.Set("location", location)

	return nil
}

// resourceThingRead handles reading the resource state
func resourceDefaultBucketRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*FirebaseClient)
	project := d.Get("project").(string)

	exists, bucketName, err := checkDefaultBucketExists(ctx, client, project)
	if err != nil {
		return diag.FromErr(err)
	}
	if !exists {
		d.SetId("")
		return nil
	}

	// Set all attributes
	d.SetId(fmt.Sprintf("projects/%s/defaultBucket", project))
	if err := d.Set("bucket_name", bucketName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("location", d.Get("location").(string)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// resourceThingUpdate handles updating the resource
func resourceDefaultBucketUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	id := d.Id()
	newName := d.Get("name").(string)

	// Update the in-memory store
	if _, exists := things[id]; exists {
		delete(things, id)
		things[newName] = newName
		d.SetId(newName)
		return nil
	}

	return diag.Errorf("thing with ID %s not found", id)
}

// resourceThingDelete handles deleting the resource
func resourceDefaultBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Note: Firebase doesn't allow deleting the default bucket without deleting the project.
	// We'll just clear the state.
	d.SetId("")
	return nil
}

// curl -X GET\
// >   -H "Authorization: Bearer $ACCESS_TOKEN" \
// >   -H "Content-Type: application/json" \
// >   "https://firebasestorage.googleapis.com/v1alpha/projects/{project}/defaultBucket"

// {
//   "name": "projects/{project}/defaultBucket",
//   "location": "US",
//   "bucket": {
//     "name": "projects/{project}/buckets/{project}.firebasestorage.app"
//   },
//   "storageClass": "STANDARD"
// }

func checkDefaultBucketExists(ctx context.Context, client *FirebaseClient, project string) (bool, string, error) {
	url := fmt.Sprintf("https://firebasestorage.googleapis.com/v1alpha/projects/%s/defaultBucket", project)
	resp, err := client.DoRequest(ctx, "GET", url, nil)
	if err != nil {
		return false, "", fmt.Errorf("failed to check default bucket: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, "", fmt.Errorf("unexpected status when checking default bucket: %d", resp.StatusCode)
	}

	var result struct {
		Bucket string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, "", fmt.Errorf("failed to decode default bucket response: %v", err)
	}

	return true, result.Bucket, nil
}
