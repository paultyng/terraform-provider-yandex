package yandex

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil" //nolint:staticcheck
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/exp/slices"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const (
	yandexFunctionVersionDefaultTimeout = 5 * time.Minute
	versionCreateSourceContentMaxBytes  = 3670016
)

func resourceYandexFunctionVersion() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexFunctionVersionCreate,
		ReadContext:   resourceYandexFunctionVersionRead,
		UpdateContext: resourceYandexFunctionVersionUpdate,
		DeleteContext: resourceYandexFunctionVersionDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexFunctionVersionDefaultTimeout),
			Update: schema.DefaultTimeout(yandexFunctionVersionDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexFunctionVersionDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"runtime": {
				Type:     schema.TypeString,
				Required: true,
			},

			"entrypoint": {
				Type:     schema.TypeString,
				Required: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"execution_timeout": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"environment": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"content": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"package"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zip_filename": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"package": {
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"content"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sha_256": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},

			"connectivity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"loggroup_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"image_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"secrets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"environment_variable": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexFunctionVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req, err := expandFunctionVersion(d)
	if err != nil {
		return diag.Errorf("Error expanding function version while creating Yandex Cloud Function Version: %s", err)
	}

	req.FunctionId = d.Get("function_id").(string)

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Functions().Function().CreateVersion(ctx, req))
	if err != nil {
		return diag.FromErr(err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Function Version: %s", err)
	}

	md, ok := protoMetadata.(*functions.CreateFunctionVersionMetadata)
	if !ok {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Function Version: %s", err)
	}

	d.SetId(md.FunctionVersionId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while requesting API to create Yandex Cloud Function Version: %s", err)
	}
	return resourceYandexFunctionVersionRead(ctx, d, meta)
}

func resourceYandexFunctionVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	versionUpdateAttributes := []string{"runtime", "entrypoint", "memory", "execution_timeout", "service_account_id",
		"environment", "connectivity", "package", "content", "secret"}

	o, n := d.GetChange("tags")
	oldTags, newTags := o.(*schema.Set), n.(*schema.Set)
	newTagsAppended := newTags.Difference(oldTags).Len() != 0

	// Remove deleted tags from the current version.
	if err := removeTags(ctx, config, d.Id(), oldTags.Difference(newTags).List()); err != nil {
		return diag.FromErr(err)
	}

	// If any attributes has changed, create a new version.
	// NOTE: Update of the version leads to the creation of a new version with changed attributes.
	// This process changes resource `id` attribute, leading to warnings in the provider log.
	// The warnings can be ignored until the provider is updated
	// to terraform-plugin-framework instead of the current SDKv2.
	//
	// See: https://stackoverflow.com/questions/75202696/is-it-ok-for-a-provider-to-update-resource-id-on-an-update
	// for more details.
	if idx := slices.IndexFunc(versionUpdateAttributes, d.HasChange); idx != -1 || newTagsAppended {
		return resourceYandexFunctionVersionCreate(ctx, d, meta)
	}

	return nil
}

func resourceYandexFunctionVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &functions.GetFunctionVersionRequest{FunctionVersionId: d.Id()}

	version, err := config.sdk.Serverless().Functions().Function().GetVersion(ctx, req)
	if err != nil {
		return handleNotFoundDiagError(err, d, fmt.Sprintf("Yandex Cloud Function Version %q", d.Id()))
	}

	d.SetId(version.Id)

	if err = d.Set("function_id", version.FunctionId); err != nil {
		return diag.FromErr(err)
	}
	if err = flattenYandexFunctionVersion(d, version); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceYandexFunctionVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	tags := d.Get("tags").(*schema.Set)
	return diag.FromErr(removeTags(ctx, config, d.Id(), tags.List()))
}

func removeTags(ctx context.Context, config *Config, functionVersionID string, tags []interface{}) error {
	for _, tag := range tags {
		req := &functions.RemoveFunctionTagRequest{FunctionVersionId: functionVersionID, Tag: tag.(string)}

		op, err := config.sdk.WrapOperation(config.sdk.Serverless().Functions().Function().RemoveTag(ctx, req))
		if err != nil {
			return fmt.Errorf("Error while requesting API to remove tag Yandex Cloud Function Version: %s", err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while requesting API to remove tag Yandex Cloud Function Version: %s", err)
		}
	}
	return nil
}

func expandFunctionVersion(d *schema.ResourceData) (*functions.CreateFunctionVersionRequest, error) {
	if !isVersionDefined(d) {
		return nil, nil
	}

	versionReq := &functions.CreateFunctionVersionRequest{}
	versionReq.Runtime = d.Get("runtime").(string)
	versionReq.Entrypoint = d.Get("entrypoint").(string)

	versionReq.Resources = &functions.Resources{Memory: int64(int(datasize.MB.Bytes()) * d.Get("memory").(int))}
	if v, ok := d.GetOk("execution_timeout"); ok {
		i, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Cannot define execution_timeout for Yandex Cloud Function: %s", err)
		}
		versionReq.ExecutionTimeout = &duration.Duration{Seconds: i}
	}
	if v, ok := d.GetOk("service_account_id"); ok {
		versionReq.ServiceAccountId = v.(string)
	}
	if v, ok := d.GetOk("environment"); ok {
		env, err := expandLabels(v)
		if err != nil {
			return nil, fmt.Errorf("Cannot define environment variables for Yandex Cloud Function: %s", err)
		}
		if len(env) != 0 {
			versionReq.Environment = env
		}
	}
	if v, ok := d.GetOk("tags"); ok {
		set := v.(*schema.Set)
		for _, t := range set.List() {
			v := t.(string)
			versionReq.Tag = append(versionReq.Tag, v)
		}
	}
	if _, ok := d.GetOk("package"); ok {
		pkg := &functions.Package{
			BucketName: d.Get("package.0.bucket_name").(string),
			ObjectName: d.Get("package.0.object_name").(string),
		}
		if v, ok := d.GetOk("package.0.sha_256"); ok {
			pkg.Sha256 = v.(string)
		}
		versionReq.PackageSource = &functions.CreateFunctionVersionRequest_Package{Package: pkg}
	} else if _, ok := d.GetOk("content"); ok {
		content, err := ZipPathToBytes(d.Get("content.0.zip_filename").(string))
		if err != nil {
			return nil, fmt.Errorf("Cannot define content for Yandex Cloud Function: %s", err)
		}
		if size := len(content); size > versionCreateSourceContentMaxBytes {
			return nil, fmt.Errorf("Zip archive content size %v exceeds the maximum size %v, use object storage to upload the content", size, versionCreateSourceContentMaxBytes)
		}
		versionReq.PackageSource = &functions.CreateFunctionVersionRequest_Content{Content: content}
	} else {
		return nil, fmt.Errorf("Package or content option must be present for Yandex Cloud Function")
	}
	if v, ok := d.GetOk("secrets"); ok {
		secretsList := v.([]interface{})

		versionReq.Secrets = make([]*functions.Secret, len(secretsList))
		for i, s := range secretsList {
			secret := s.(map[string]interface{})

			fs := &functions.Secret{}
			if ID, ok := secret["id"]; ok {
				fs.Id = ID.(string)
			}
			if versionID, ok := secret["version_id"]; ok {
				fs.VersionId = versionID.(string)
			}
			if key, ok := secret["key"]; ok {
				fs.Key = key.(string)
			}
			if environmentVariable, ok := secret["environment_variable"]; ok {
				fs.Reference = &functions.Secret_EnvironmentVariable{EnvironmentVariable: environmentVariable.(string)}
			}

			versionReq.Secrets[i] = fs
		}
	}
	if connectivity := expandFunctionConnectivity(d); connectivity != nil {
		versionReq.Connectivity = connectivity
	}

	return versionReq, nil
}

func expandFunctionConnectivity(d *schema.ResourceData) *functions.Connectivity {
	if id, ok := d.GetOk("connectivity.0.network_id"); ok {
		return &functions.Connectivity{NetworkId: id.(string)}
	}
	return nil
}

func flattenYandexFunctionVersion(d *schema.ResourceData, version *functions.Version) error {
	if err := d.Set("runtime", version.Runtime); err != nil {
		return err
	}
	if err := d.Set("entrypoint", version.Entrypoint); err != nil {
		return err
	}
	if err := d.Set("service_account_id", version.ServiceAccountId); err != nil {
		return err
	}
	if err := d.Set("environment", version.Environment); err != nil {
		return err
	}
	if err := d.Set("loggroup_id", version.LogGroupId); err != nil {
		return err
	}
	if err := d.Set("image_size", version.ImageSize); err != nil {
		return err
	}

	if version.Resources != nil {
		if err := d.Set("memory", int(version.Resources.Memory/int64(datasize.MB.Bytes()))); err != nil {
			return err
		}
	}

	if version.ExecutionTimeout != nil && version.ExecutionTimeout.Seconds != 0 {
		if err := d.Set("execution_timeout", strconv.FormatInt(version.ExecutionTimeout.Seconds, 10)); err != nil {
			return err
		}
	}

	if connectivity := flattenFunctionConnectivity(version.Connectivity); connectivity != nil {
		if err := d.Set("connectivity", connectivity); err != nil {
			return err
		}
	}

	if version.Secrets != nil {
		secrets := flattenFunctionSecrets(version.Secrets)
		if err := d.Set("secret", secrets); err != nil {
			return err
		}
	}

	if version.Connectivity != nil {
		m := make(map[string]interface{})
		if len(version.Connectivity.SubnetId) > 0 && !allSubnetsIdsAreBlank(version.Connectivity.SubnetId) {
			m["subnet_ids"] = version.Connectivity.SubnetId
		}
		m["network_id"] = d.Get("connectivity.0.network_id")
		if err := d.Set("connectivity", []map[string]interface{}{m}); err != nil {
			return err
		}
	}

	tags := &schema.Set{F: schema.HashString}
	for _, v := range version.Tags {
		if v != "$latest" {
			tags.Add(v)
		}
	}
	return d.Set("tags", tags)
}

func flattenFunctionConnectivity(connectivity *functions.Connectivity) []interface{} {
	if connectivity == nil || connectivity.NetworkId == "" {
		return nil
	}
	return []interface{}{map[string]interface{}{"network_id": connectivity.NetworkId}}
}

func flattenFunctionSecrets(secrets []*functions.Secret) []map[string]interface{} {
	s := make([]map[string]interface{}, len(secrets))

	for i, secret := range secrets {
		s[i] = map[string]interface{}{
			"id":                   secret.Id,
			"version_id":           secret.VersionId,
			"key":                  secret.Key,
			"environment_variable": secret.GetEnvironmentVariable(),
		}
	}
	return s
}

func isVersionDefined(d *schema.ResourceData) bool {
	versionAttributes := []string{"runtime", "entrypoint", "memory", "execution_timeout", "service_account_id",
		"environment", "tags", "connectivity", "package", "content", "secret"}

	for _, attribute := range versionAttributes {
		if _, ok := d.GetOk(attribute); ok {
			return true
		}
	}
	return false
}

func zipPathToWriter(root string, buffer io.Writer) error {
	rootDir := filepath.Dir(root)
	zipWriter := zip.NewWriter(buffer)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel := strings.TrimPrefix(path, rootDir)
		entry, err := zipWriter.Create(rel)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		if _, err := io.Copy(entry, file); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}
	return nil
}

func ZipPathToBytes(root string) ([]byte, error) {

	// first, check if the path corresponds to already zipped file
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if info.Mode().IsRegular() {
		bytes, err := ioutil.ReadFile(root)
		if err != nil {
			return nil, err
		}
		if isZipContent(bytes) {
			// file has already zipped, return its content
			return bytes, nil
		}
	}

	// correct path (make directory looks like a directory)
	if info.Mode().IsDir() && !strings.HasSuffix(root, string(os.PathSeparator)) {
		root = root + "/"
	}

	// do real zipping of the given path
	var buffer bytes.Buffer
	err = zipPathToWriter(root, &buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func isZipContent(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x50 && buf[1] == 0x4B &&
		(buf[2] == 0x3 || buf[2] == 0x5 || buf[2] == 0x7) &&
		(buf[3] == 0x4 || buf[3] == 0x6 || buf[3] == 0x8)
}

func allSubnetsIdsAreBlank(subnetIds []string) bool {
	for _, id := range subnetIds {
		if id != "" {
			return false
		}
	}
	return true
}
