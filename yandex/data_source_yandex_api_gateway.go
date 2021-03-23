package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexApiGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexApiGatewayRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"api_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexApiGatewayRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "api_gateway_id", "name")
	if err != nil {
		return err
	}

	apiGatewayID := d.Get("api_gateway_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		apiGatewayID, err = resolveObjectID(ctx, config, d, sdkresolvers.APIGatewayResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Cloud API Gateway by name: %v", err)
		}
	}

	req := apigateway.GetApiGatewayRequest{
		ApiGatewayId: apiGatewayID,
	}

	apiGateway, err := config.sdk.Serverless().APIGateway().ApiGateway().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud API Gateway %q", d.Id()))
	}

	d.SetId(apiGateway.Id)
	d.Set("api_gateway_id", apiGateway.Id)
	return flattenYandexApiGateway(d, apiGateway)
}
