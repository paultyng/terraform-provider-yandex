package yandex

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_PersQueue_V1"

	"github.com/ydb-platform/ydb-go-persqueue-sdk/controlplane"
	"github.com/ydb-platform/ydb-go-persqueue-sdk/session"
	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ydsCodecGZIP = "gzip"
	ydsCodecRaw  = "raw"
	ydsCodecZSTD = "zstd"
)

var (
	ydsAllowedCodecs = []string{
		ydsCodecGZIP,
		ydsCodecRaw,
		ydsCodecZSTD,
	}

	ydsCodecNameToCodec = map[string]Ydb_PersQueue_V1.Codec{
		ydsCodecRaw:  Ydb_PersQueue_V1.Codec_CODEC_RAW,
		ydsCodecGZIP: Ydb_PersQueue_V1.Codec_CODEC_GZIP,
		ydsCodecZSTD: Ydb_PersQueue_V1.Codec_CODEC_ZSTD,
	}
)

func createYDSServerlessClient(ctx context.Context, databaseEndpoint string, config *Config) (controlplane.ControlPlane, error) {
	endpoint, databasePath, useTLS, err := parseYandexYDBDatabaseEndpoint(databaseEndpoint)
	if err != nil {
		return nil, err
	}

	opts := session.Options{
		Credentials: credentials.NewAccessTokenCredentials(config.Token),
		Endpoint:    endpoint,
		Database:    databasePath,
	}
	if useTLS {
		opts.TLSConfig = &tls.Config{}
	}

	return controlplane.NewControlPlaneClient(ctx, opts)
}

func resourceYandexYDSServerlessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client, err := createYDSServerlessClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	err = client.CreateTopic(ctx, &Ydb_PersQueue_V1.CreateTopicRequest{
		Path:            d.Get("stream_name").(string),
		OperationParams: &Ydb_Operations.OperationParams{},
		Settings: &Ydb_PersQueue_V1.TopicSettings{
			SupportedCodecs: []Ydb_PersQueue_V1.Codec{
				// TODO(shmel1k@): add mapping.
				Ydb_PersQueue_V1.Codec_CODEC_GZIP,
			},
			PartitionsCount:   int32(d.Get("partitions_count").(int)),
			RetentionPeriodMs: int64(d.Get("retention_period_ms").(int)),
			SupportedFormat:   Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
		},
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}

	d.SetId(d.Get("database_endpoint").(string) + "/" + d.Get("stream_name").(string))

	return nil
}

func flattenYDSDescription(d *schema.ResourceData, desc *Ydb_PersQueue_V1.DescribeTopicResult) error {
	_ = d.Set("stream_name", desc.Self.Name)
	_ = d.Set("partitions_count", desc.Settings.PartitionsCount)
	_ = d.Set("retention_period_ms", desc.Settings.RetentionPeriodMs)

	supportedCodecs := make([]string, 0, len(desc.Settings.SupportedCodecs))
	for _, v := range desc.Settings.SupportedCodecs {
		switch v {
		case Ydb_PersQueue_V1.Codec_CODEC_RAW:
			supportedCodecs = append(supportedCodecs, ydsCodecRaw)
		case Ydb_PersQueue_V1.Codec_CODEC_ZSTD:
			supportedCodecs = append(supportedCodecs, ydsCodecZSTD)
		case Ydb_PersQueue_V1.Codec_CODEC_GZIP:
			supportedCodecs = append(supportedCodecs, ydsCodecGZIP)
		}
	}

	err := d.Set("supported_codecs", supportedCodecs)
	if err != nil {
		return err
	}

	return d.Set("database_endpoint", d.Get("database_endpoint").(string)) // TODO(shmel1k@): remove probably.
}

func resourceYandexYDSServerlessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	client, err := createYDSServerlessClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	description, err := client.DescribeTopic(ctx, d.Get("stream_name").(string))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			d.SetId("") // NOTE(shmel1k@): marking as non-existing resource.
			return nil
		}
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "resource: failed to describe stream",
				Detail:   err.Error(),
			},
		}
	}

	err = flattenYDSDescription(d, description)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to flatten stream description",
				Detail:   err.Error(),
			},
		}
	}

	return nil
}

func mergeYDSReadRulesSettings(
	d *schema.ResourceData,
	readRules []*Ydb_PersQueue_V1.TopicSettings_ReadRule,
) (
	consumersForUpdate []*Ydb_PersQueue_V1.TopicSettings_ReadRule,
) {
	rules := make(map[string]*Ydb_PersQueue_V1.TopicSettings_ReadRule, len(readRules))
	for i := 0; i < len(readRules); i++ {
		rules[readRules[i].ConsumerName] = readRules[i]
	}

	// TODO(shmel1k@): add tests.
	consumers := d.Get("consumers").([]interface{})
	for _, v := range consumers {
		hasDiff := false
		consumer := v.(map[string]interface{})
		consumerName, ok := consumer["name"].(string)
		if !ok {
			// TODO(shmel1k@): think about error.
			continue
		}
		supportedCodecs, ok := consumer["supported_codecs"].([]string)
		if !ok {
			// TODO(shmel1k@): think about error.
			continue
		}
		startingMessageTs, ok := consumer["starting_message_timestamp_ms"].(int)
		if !ok {
			continue
		}
		serviceType, ok := consumer["service_type"].(string)
		if !ok {
			continue
		}

		r := rules[consumerName]
		if r.ServiceType != serviceType {
			hasDiff = true
			r.ServiceType = serviceType
		}
		if r.StartingMessageTimestampMs != int64(startingMessageTs) {
			hasDiff = true
			r.StartingMessageTimestampMs = int64(startingMessageTs)
		}

		newCodecs := make([]Ydb_PersQueue_V1.Codec, 0, len(supportedCodecs))
		for _, codec := range supportedCodecs {
			c := ydsCodecNameToCodec[codec]
			hasCodec := false
			for _, cc := range r.SupportedCodecs {
				if c == cc {
					hasCodec = true
					break
				}
			}
			if !hasCodec {
				hasDiff = true
			}
			newCodecs = append(newCodecs, c)
		}
		r.SupportedCodecs = newCodecs

		if hasDiff {
			consumersForUpdate = append(consumersForUpdate, r)
		}
	}

	return
}

func mergeYDSSettings(
	d *schema.ResourceData,
	settings *Ydb_PersQueue_V1.TopicSettings,
) *Ydb_PersQueue_V1.TopicSettings {
	if d.HasChange("partitions_count") {
		settings.PartitionsCount = int32(d.Get("partitions_count").(int))
	}
	if d.HasChange("supported_codecs") {
		codecs := d.Get("supported_codecs").([]interface{})
		updatedCodecs := make([]Ydb_PersQueue_V1.Codec, 0, len(codecs))

		for _, c := range codecs {
			cc, ok := ydsCodecNameToCodec[strings.ToLower(c.(string))]
			if !ok {
				// TODO(shmel1k@): add validation of unsupported codecs. Use default if unknown is found.
				panic(fmt.Sprintf("Unsupported codec %q found after validation", cc))
			}
			updatedCodecs = append(updatedCodecs, cc)
		}
		settings.SupportedCodecs = updatedCodecs
	}
	if d.HasChange("retention_period_ms") {
		settings.RetentionPeriodMs = int64(d.Get("retention_period_ms").(int))
	}
	if d.HasChange("consumers") {
		settings.ReadRules = mergeYDSReadRulesSettings(d, settings.ReadRules)
	}

	return settings
}

func performYandexYDSUpdate(ctx context.Context, d *schema.ResourceData, config *Config) diag.Diagnostics {
	client, err := createYDSServerlessClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	streamName := d.Get("stream_name").(string)
	desc, err := client.DescribeTopic(ctx, streamName)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("failed to get description for stream %q", streamName),
				Detail:   err.Error(),
			},
		}
	}
	newSettings := mergeYDSSettings(d, desc.GetSettings())

	err = client.AlterTopic(ctx, &Ydb_PersQueue_V1.AlterTopicRequest{
		Path:     streamName,
		Settings: newSettings,
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "got error when tried to alter stream",
				Detail:   err.Error(),
			},
		}
	}

	return nil
}

func resourceYandexYDSServerlessUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	return performYandexYDSUpdate(ctx, d, config)
}

func resourceYandexYDSServerlessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	client, err := createYDSServerlessClient(ctx, d.Get("database_endpoint").(string), config)
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to initialize yds control plane client",
				Detail:   err.Error(),
			},
		}
	}
	defer func() {
		_ = client.Close()
	}()

	streamName := d.Get("stream_name").(string)
	err = client.DropTopic(ctx, &Ydb_PersQueue_V1.DropTopicRequest{
		Path: streamName,
	})
	if err != nil {
		return diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "failed to delete stream",
				Detail:   err.Error(),
			},
		}
	}
	return nil
}

func resourceYandexYDSServerless() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexYDSServerlessCreate,
		ReadContext:   resourceYandexYDSServerlessRead,
		UpdateContext: resourceYandexYDSServerlessUpdate,
		DeleteContext: resourceYandexYDSServerlessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			// TODO(shmel1k@): think about own timeouts.
			Default: schema.DefaultTimeout(yandexYDBServerlessDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"database_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"partitions_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"supported_codecs": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					// TODO(shmel1k@): add validation.
					Type: schema.TypeString,
				},
			},
			"retention_period_ms": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1000 * 60 * 60 * 24, // 1 day
			},
			"consumers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"supported_codecs": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								// TODO(shmel1k@): add validation.
								Type: schema.TypeString,
							},
						},
						"starting_message_timestamp_ms": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"service_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
