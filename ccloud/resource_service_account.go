package ccloud

import (
	"context"
	"fmt"
	"log"
	"strconv"

	ccloud "github.com/cgroschupp/go-client-confluent-cloud/confluentcloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/go-cty/cty"
)

func serviceAccountResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: serviceAccountCreate,
		ReadContext:   serviceAccountRead,
		DeleteContext: serviceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Service Account Description",
			},
		},
	}
}

func serviceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*ccloud.Client)

	var diags diag.Diagnostics

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	existingAccount, err := lookupServiceAccount(0, name, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if existingAccount != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Username is already taken",
			Detail:   fmt.Sprintf("Duplicate username %s; already existing account has ID %d", name, existingAccount.ID),
			AttributePath: cty.Path{cty.GetAttrStep{Name: "name"}},
		})
		return diags
	}

	// There is no conflict, let's create the account.
	req := ccloud.ServiceAccountCreateRequest{
		Name:        name,
		Description: description,
	}

	serviceAccount, err := c.CreateServiceAccount(&req)
	if err == nil {
		d.SetId(fmt.Sprintf("%d", serviceAccount.ID))

		err = d.Set("name", serviceAccount.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		err = d.Set("description", serviceAccount.Description)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		log.Printf("[ERROR] Could not create Service Account: %s", err)
	}

	return diag.FromErr(err)
}

func serviceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	searchIDString := d.Get("id").(string)
	searchID, err := strconv.Atoi(searchIDString)
	if err != nil {
		return diag.FromErr(err)
	}

	account, err := lookupServiceAccount(searchID, "", meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if account == nil {
		d.SetId("")
		return nil
	}

	d.Set("name", account.Name)
	// To consider: do we want to recreate account on change of description?
	// It is suboptimal, but who knows what the better solution is.
	d.Set("description", account.Description)
	return nil
}

// Look up service account by id / name. Empty or zeroed parameters = match all.
func lookupServiceAccount(id int, name string, meta interface{}) (*ccloud.ServiceAccount, error) {
	c := meta.(*ccloud.Client)
	accounts, err := c.ListServiceAccounts()
	if err != nil {
		return nil, err
	}

	for _, account := range accounts {
		if id != 0 && account.ID != id {
			continue
		}

		if name != "" && account.Name != name {
			continue
		}

		return &account, nil
	}

	return nil, nil
}


func serviceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*ccloud.Client)

	ID, err := strconv.Atoi(d.Id())
	if err != nil {
		log.Printf("[ERROR] Could not parse Service Account ID %s to int", d.Id())
		return diag.FromErr(err)
	}

	err = c.DeleteServiceAccount(ID)
	if err != nil {
		log.Printf("[ERROR] Service Account can not be deleted: %d", ID)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Service Account deleted: %d", ID)

	return nil
}
