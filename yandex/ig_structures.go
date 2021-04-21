package yandex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

func flattenInstanceGroupInstanceTemplateResources(resSpec *instancegroup.ResourcesSpec) ([]map[string]interface{}, error) {
	resourceMap := map[string]interface{}{
		"cores":         int(resSpec.Cores),
		"core_fraction": int(resSpec.CoreFraction),
		"memory":        toGigabytesInFloat(resSpec.Memory),
		"gpus":          int(resSpec.Gpus),
	}

	return []map[string]interface{}{resourceMap}, nil
}

func flattenInstanceGroupManagedInstanceNetworkInterfaces(instance *instancegroup.ManagedInstance) ([]map[string]interface{}, string, string, error) {
	nics := make([]map[string]interface{}, len(instance.NetworkInterfaces))
	var externalIP, internalIP string

	for i, iface := range instance.NetworkInterfaces {
		index, err := strconv.Atoi(iface.Index)
		if err != nil {
			return nil, "", "", fmt.Errorf("Error while convert index of Network Interface: %s", err)
		}

		nics[i] = map[string]interface{}{
			"index":       index,
			"mac_address": iface.MacAddress,
			"subnet_id":   iface.SubnetId,
			"ipv4":        false,
			"ipv6":        false,
		}

		if iface.PrimaryV4Address != nil {
			nics[i]["ipv4"] = true
			nics[i]["ip_address"] = iface.PrimaryV4Address.Address
			if internalIP == "" {
				internalIP = iface.PrimaryV4Address.Address
			}

			if iface.PrimaryV4Address.OneToOneNat != nil {
				nics[i]["nat"] = true
				nics[i]["nat_ip_address"] = iface.PrimaryV4Address.OneToOneNat.Address
				nics[i]["nat_ip_version"] = iface.PrimaryV4Address.OneToOneNat.IpVersion.String()
				if externalIP == "" {
					externalIP = iface.PrimaryV4Address.OneToOneNat.Address
				}
			} else {
				nics[i]["nat"] = false
			}
		}

		if iface.PrimaryV6Address != nil {
			nics[i]["ipv6"] = true
			nics[i]["ipv6_address"] = iface.PrimaryV6Address.Address
			if externalIP == "" {
				externalIP = iface.PrimaryV6Address.Address
			}
		}
	}

	return nics, externalIP, internalIP, nil
}

func flattenInstanceGroupInstanceTemplate(template *instancegroup.InstanceTemplate) ([]map[string]interface{}, error) {
	templateMap := make(map[string]interface{})

	templateMap["description"] = template.GetDescription()
	templateMap["labels"] = template.GetLabels()
	templateMap["platform_id"] = template.GetPlatformId()
	templateMap["metadata"] = template.GetMetadata()
	templateMap["service_account_id"] = template.GetServiceAccountId()
	templateMap["name"] = template.GetName()
	templateMap["hostname"] = template.GetHostname()

	resourceSpec, err := flattenInstanceGroupInstanceTemplateResources(template.GetResourcesSpec())
	if err != nil {
		return nil, err
	}
	templateMap["resources"] = resourceSpec

	bootDiskSpec, err := flattenInstanceGroupAttachedDisk(template.GetBootDiskSpec())
	if err != nil {
		return []map[string]interface{}{templateMap}, err
	}
	templateMap["boot_disk"] = []map[string]interface{}{bootDiskSpec}

	secondarySpecs := template.GetSecondaryDiskSpecs()
	secondarySpecsList := make([]map[string]interface{}, len(secondarySpecs))
	for i, spec := range secondarySpecs {
		flattened, err := flattenInstanceGroupAttachedDisk(spec)
		if err != nil {
			return nil, err
		}
		secondarySpecsList[i] = flattened
	}
	templateMap["secondary_disk"] = secondarySpecsList

	networkSpecs := template.GetNetworkInterfaceSpecs()
	networkSpecsList := make([]map[string]interface{}, len(networkSpecs))
	for i, spec := range networkSpecs {
		networkSpecsList[i] = flattenInstanceGroupNetworkInterfaceSpec(spec)
	}
	templateMap["network_interface"] = networkSpecsList

	if template.SchedulingPolicy != nil {
		templateMap["scheduling_policy"] = []map[string]interface{}{{"preemptible": template.SchedulingPolicy.Preemptible}}
	}

	if template.PlacementPolicy != nil {
		placementPolicy, err := flattenInstanceGroupPlacementPolicy(template.PlacementPolicy)
		if err != nil {
			return []map[string]interface{}{templateMap}, err
		}
		templateMap["placement_policy"] = placementPolicy
	}

	if template.NetworkSettings != nil {
		templateMap["network_settings"] = flattenInstanceGroupNetworkSettings(template.GetNetworkSettings())
	}

	return []map[string]interface{}{templateMap}, nil
}

func flattenInstanceGroupVariable(v []*instancegroup.Variable) map[string]string {
	variables := make(map[string]string)
	for _, raw := range v {
		variables[raw.GetKey()] = raw.GetValue()
	}
	return variables
}

func flattenInstanceGroupNetworkSettings(ns *instancegroup.NetworkSettings) []map[string]interface{} {
	return []map[string]interface{}{{"type": ns.GetType().String()}}
}

func flattenInstanceGroupNetworkInterfaceSpec(nicSpec *instancegroup.NetworkInterfaceSpec) map[string]interface{} {
	nat := (nicSpec.GetPrimaryV4AddressSpec().GetOneToOneNatSpec() != nil) ||
		(nicSpec.GetPrimaryV6AddressSpec().GetOneToOneNatSpec() != nil)

	subnets := &schema.Set{F: schema.HashString}

	if nicSpec.SubnetIds != nil {
		for _, s := range nicSpec.SubnetIds {
			subnets.Add(s)
		}
	}

	networkInterface := map[string]interface{}{
		"network_id": nicSpec.NetworkId,
		"subnet_ids": subnets,
		"nat":        nat,
		"ipv4":       nicSpec.PrimaryV4AddressSpec != nil,
		"ipv6":       nicSpec.PrimaryV6AddressSpec != nil,
	}

	if nicSpec.GetSecurityGroupIds() != nil {
		networkInterface["security_group_ids"] = convertStringArrToInterface(nicSpec.SecurityGroupIds)
	}

	if sp := nicSpec.GetPrimaryV4AddressSpec().GetDnsRecordSpecs(); sp != nil {
		networkInterface["dns_record"] = flattenInstanceGroupDnsRecords(sp)
	}

	if sp := nicSpec.GetPrimaryV6AddressSpec().GetDnsRecordSpecs(); sp != nil {
		networkInterface["ipv6_dns_record"] = flattenInstanceGroupDnsRecords(sp)
	}

	if sp := nicSpec.GetPrimaryV4AddressSpec().GetOneToOneNatSpec().GetDnsRecordSpecs(); sp != nil {
		networkInterface["nat_dns_record"] = flattenInstanceGroupDnsRecords(sp)
	}

	return networkInterface
}

func flattenInstanceGroupDnsRecords(specs []*instancegroup.DnsRecordSpec) []map[string]interface{} {
	res := make([]map[string]interface{}, len(specs))

	for i, spec := range specs {
		res[i] = map[string]interface{}{
			"fqdn": spec.Fqdn,
			"ptr":  spec.Ptr,
		}
		if spec.DnsZoneId != "" {
			res[i]["dns_zone_id"] = spec.DnsZoneId
		}
		if spec.Ttl != 0 {
			res[i]["ttl"] = spec.Ttl
		}
	}

	return res
}

func flattenInstanceGroupDeployPolicy(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}
	if ig.DeployPolicy != nil {
		res["max_expansion"] = ig.DeployPolicy.MaxExpansion
		res["max_creating"] = ig.DeployPolicy.MaxCreating
		res["max_deleting"] = ig.DeployPolicy.MaxDeleting
		res["max_unavailable"] = ig.DeployPolicy.MaxUnavailable
		if ig.DeployPolicy.StartupDuration != nil {
			res["startup_duration"] = ig.DeployPolicy.StartupDuration.Seconds
		}
		res["strategy"] = strings.ToLower(ig.DeployPolicy.Strategy.String())
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupScalePolicy(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	if sp := ig.GetScalePolicy().GetFixedScale(); sp != nil {
		res["fixed_scale"] = []map[string]interface{}{{"size": int(sp.Size)}}
	}

	if sp := ig.GetScalePolicy().GetAutoScale(); sp != nil {
		res["auto_scale"], _ = flattenInstanceGroupAutoScale(sp)
	}

	if sp := ig.GetScalePolicy().GetTestAutoScale(); sp != nil {
		res["test_auto_scale"], _ = flattenInstanceGroupAutoScale(sp)
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupAutoScale(sp *instancegroup.ScalePolicy_AutoScale) ([]map[string]interface{}, error) {
	subres := map[string]interface{}{}
	subres["min_zone_size"] = int(sp.MinZoneSize)
	subres["max_size"] = int(sp.MaxSize)
	subres["initial_size"] = int(sp.InitialSize)

	if sp.MeasurementDuration != nil {
		subres["measurement_duration"] = int(sp.MeasurementDuration.Seconds)
	}

	if sp.WarmupDuration != nil {
		subres["warmup_duration"] = int(sp.WarmupDuration.Seconds)
	}

	if sp.StabilizationDuration != nil {
		subres["stabilization_duration"] = int(sp.StabilizationDuration.Seconds)
	}

	if sp.CpuUtilizationRule != nil {
		subres["cpu_utilization_target"] = sp.CpuUtilizationRule.UtilizationTarget
	}

	if len(sp.CustomRules) > 0 {
		rules := make([]map[string]interface{}, len(sp.CustomRules))
		subres["custom_rule"] = rules

		for i, rule := range sp.CustomRules {
			rules[i] = map[string]interface{}{
				"rule_type":   instancegroup.ScalePolicy_CustomRule_RuleType_name[int32(rule.RuleType)],
				"metric_type": instancegroup.ScalePolicy_CustomRule_MetricType_name[int32(rule.MetricType)],
				"metric_name": rule.MetricName,
				"target":      rule.Target,
				"labels":      rule.GetLabels(),
			}
		}
	}

	return []map[string]interface{}{subres}, nil
}

func flattenInstanceGroupAllocationPolicy(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	zones := &schema.Set{F: schema.HashString}

	for _, zone := range ig.AllocationPolicy.Zones {
		zones.Add(zone.ZoneId)
	}

	res["zones"] = zones
	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupHealthChecks(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	if ig.HealthChecksSpec == nil {
		return nil, nil
	}

	res := make([]map[string]interface{}, len(ig.HealthChecksSpec.HealthCheckSpecs))

	for i, spec := range ig.HealthChecksSpec.HealthCheckSpecs {
		specDict := map[string]interface{}{}
		specDict["interval"] = int(spec.Interval.Seconds)
		specDict["timeout"] = int(spec.Timeout.Seconds)
		specDict["healthy_threshold"] = int(spec.HealthyThreshold)
		specDict["unhealthy_threshold"] = int(spec.UnhealthyThreshold)

		if spec.GetHttpOptions() != nil {
			specDict["http_options"] = []map[string]interface{}{
				{
					"port": int(spec.GetHttpOptions().Port),
					"path": spec.GetHttpOptions().Path,
				},
			}
		}

		if spec.GetTcpOptions() != nil {
			specDict["tcp_options"] = []map[string]interface{}{
				{
					"port": int(spec.GetTcpOptions().Port),
				},
			}
		}

		res[i] = specDict
	}
	return res, nil
}

func flattenInstanceGroupLoadBalancerState(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	if loadBalancerState := ig.GetLoadBalancerState(); loadBalancerState != nil {
		res["target_group_id"] = loadBalancerState.TargetGroupId
		res["status_message"] = loadBalancerState.StatusMessage
	}

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupLoadBalancerSpec(ig *instancegroup.InstanceGroup) ([]map[string]interface{}, error) {
	if ig.LoadBalancerSpec == nil || ig.LoadBalancerSpec.TargetGroupSpec == nil {
		return nil, nil
	}

	res := map[string]interface{}{}
	res["target_group_name"] = ig.LoadBalancerSpec.TargetGroupSpec.GetName()
	res["target_group_description"] = ig.LoadBalancerSpec.TargetGroupSpec.GetDescription()
	res["target_group_labels"] = ig.LoadBalancerSpec.TargetGroupSpec.GetLabels()
	res["target_group_id"] = ig.LoadBalancerState.GetTargetGroupId()
	res["status_message"] = ig.LoadBalancerState.GetStatusMessage()

	return []map[string]interface{}{res}, nil
}

func flattenInstanceGroupPlacementPolicy(policy *instancegroup.PlacementPolicy) ([]map[string]interface{}, error) {
	if policy != nil {
		placementMap := map[string]interface{}{
			"placement_group_id": policy.PlacementGroupId,
		}
		return []map[string]interface{}{placementMap}, nil
	}
	return nil, nil
}

func expandInstanceGroupResourcesSpec(d *schema.ResourceData, prefix string) (*instancegroup.ResourcesSpec, error) {
	rs := &instancegroup.ResourcesSpec{}

	if v, ok := d.GetOk(prefix + ".0.cores"); ok {
		rs.Cores = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + ".0.gpus"); ok {
		rs.Gpus = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + ".0.core_fraction"); ok {
		rs.CoreFraction = int64(v.(int))
	}

	if v, ok := d.GetOk(prefix + ".0.memory"); ok {
		rs.Memory = toBytesFromFloat(v.(float64))
	}

	return rs, nil
}

func expandInstanceGroupTemplateAttachedDiskSpec(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.AttachedDiskSpec, error) {
	ads := &instancegroup.AttachedDiskSpec{}

	if v, ok := d.GetOk(prefix + ".device_name"); ok {
		ads.DeviceName = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".mode"); ok {
		diskMode, err := parseInstanceGroupDiskMode(v.(string))
		if err != nil {
			return nil, err
		}

		ads.Mode = diskMode
	}

	// create new one disk
	if _, ok := d.GetOk(prefix + ".initialize_params"); ok {
		bootDiskSpec, err := expandInstanceGroupAttachenDiskSpecSpec(d, prefix+".initialize_params.0", config)
		if err != nil {
			return nil, err
		}

		ads.DiskSpec = bootDiskSpec
	}

	return ads, nil
}

func expandInstanceGroupAttachenDiskSpecSpec(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.AttachedDiskSpec_DiskSpec, error) {
	diskSpec := &instancegroup.AttachedDiskSpec_DiskSpec{}

	if v, ok := d.GetOk(prefix + ".description"); ok {
		diskSpec.Description = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".type"); ok {
		diskSpec.TypeId = v.(string)
	}

	if _, ok := d.GetOk(prefix + ".image_id"); ok {
		if _, ok := d.GetOk(prefix + ".snapshot_id"); ok {
			return diskSpec, fmt.Errorf("Use one of  'image_id', 'snapshot_id', not both.")
		}
	}

	var minStorageSizeBytes int64
	if v, ok := d.GetOk(prefix + ".image_id"); ok {
		imageID := v.(string)
		diskSpec.SourceOneof = &instancegroup.AttachedDiskSpec_DiskSpec_ImageId{
			ImageId: imageID,
		}

		size, err := getImageMinStorageSize(imageID, config)
		if err != nil {
			return nil, err
		}
		minStorageSizeBytes = size
	}

	if v, ok := d.GetOk(prefix + ".snapshot_id"); ok {
		snapshotID := v.(string)
		diskSpec.SourceOneof = &instancegroup.AttachedDiskSpec_DiskSpec_SnapshotId{
			SnapshotId: snapshotID,
		}

		size, err := getSnapshotMinStorageSize(snapshotID, config)
		if err != nil {
			return nil, err
		}
		minStorageSizeBytes = size
	}

	if v, ok := d.GetOk(prefix + ".size"); ok {
		diskSpec.Size = toBytes(v.(int))
	}

	if diskSpec.Size == 0 {
		diskSpec.Size = minStorageSizeBytes
	}

	return diskSpec, nil
}

func expandInstanceGroupSecondaryDiskSpecs(d *schema.ResourceData, prefix string, config *Config) ([]*instancegroup.AttachedDiskSpec, error) {
	secondaryDisksCount := d.Get(prefix + ".#").(int)
	ads := make([]*instancegroup.AttachedDiskSpec, secondaryDisksCount)

	for i := 0; i < secondaryDisksCount; i++ {
		disk, err := expandInstanceGroupTemplateAttachedDiskSpec(d, fmt.Sprintf(prefix+".%d", i), config)
		if err != nil {
			return nil, err
		}
		ads[i] = disk
	}
	return ads, nil
}

func expandInstanceGroupNetworkInterfaceSpecs(d *schema.ResourceData, prefix string) ([]*instancegroup.NetworkInterfaceSpec, error) {
	nicsConfig := d.Get(prefix).([]interface{})
	nics := make([]*instancegroup.NetworkInterfaceSpec, len(nicsConfig))

	for i, raw := range nicsConfig {
		nics[i] = expandInstanceGroupNetworkInterfaceSpec(raw.(map[string]interface{}))
	}

	return nics, nil
}

func expandInstanceGroupNetworkInterfaceSpec(data map[string]interface{}) *instancegroup.NetworkInterfaceSpec {
	res := &instancegroup.NetworkInterfaceSpec{
		NetworkId: data["network_id"].(string),
	}

	if subnets, ok := data["subnet_ids"]; ok {
		sub := subnets.(*schema.Set).List()

		res.SubnetIds = make([]string, 0)

		for _, s := range sub {
			res.SubnetIds = append(res.SubnetIds, s.(string))
		}
	}

	if enableIPV4, ok := data["ipv4"].(bool); ok && enableIPV4 {
		res.PrimaryV4AddressSpec = &instancegroup.PrimaryAddressSpec{}
	}

	if enableIPV6, ok := data["ipv6"].(bool); ok && enableIPV6 {
		res.PrimaryV6AddressSpec = &instancegroup.PrimaryAddressSpec{}
	}

	if nat, ok := data["nat"].(bool); ok && nat {
		natSpec := &instancegroup.OneToOneNatSpec{
			IpVersion: instancegroup.IpVersion_IPV4,
		}
		if res.PrimaryV4AddressSpec == nil {
			res.PrimaryV4AddressSpec = &instancegroup.PrimaryAddressSpec{
				OneToOneNatSpec: natSpec,
			}
		} else {
			res.PrimaryV4AddressSpec.OneToOneNatSpec = natSpec
		}
	}

	if sgids, ok := data["security_group_ids"]; ok {
		res.SecurityGroupIds = expandSecurityGroupIds(sgids)
	}

	if d, ok := data["dns_record"]; ok {
		if res.PrimaryV4AddressSpec != nil {
			res.PrimaryV4AddressSpec.DnsRecordSpecs = expandInstanceGroupDnsRecords(d.([]interface{}))
		}
	}

	if d, ok := data["ipv6_dns_record"]; ok {
		if res.PrimaryV6AddressSpec != nil {
			res.PrimaryV6AddressSpec.DnsRecordSpecs = expandInstanceGroupDnsRecords(d.([]interface{}))
		}
	}

	if d, ok := data["nat_dns_record"]; ok {
		if res.PrimaryV4AddressSpec != nil && res.PrimaryV4AddressSpec.OneToOneNatSpec != nil {
			res.PrimaryV4AddressSpec.OneToOneNatSpec.DnsRecordSpecs = expandInstanceGroupDnsRecords(d.([]interface{}))
		}
	}

	return res
}

func expandInstanceGroupDnsRecords(data []interface{}) []*instancegroup.DnsRecordSpec {
	recs := make([]*instancegroup.DnsRecordSpec, len(data))

	for i, raw := range data {
		d := raw.(map[string]interface{})
		r := &instancegroup.DnsRecordSpec{Fqdn: d["fqdn"].(string)}
		if s, ok := d["dns_zone_id"]; ok {
			r.DnsZoneId = s.(string)
		}
		if s, ok := d["ttl"]; ok {
			r.Ttl = int64(s.(int))
		}
		if s, ok := d["ptr"]; ok {
			r.Ptr = s.(bool)
		}
		recs[i] = r
	}

	return recs
}

// revive:disable:var-naming
func expandInstanceGroupInstanceTemplate(d *schema.ResourceData, prefix string, config *Config) (*instancegroup.InstanceTemplate, error) {
	var platformId, description, serviceAccount, name, hostname string

	if v, ok := d.GetOk(prefix + ".platform_id"); ok {
		platformId = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".description"); ok {
		description = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".service_account_id"); ok {
		serviceAccount = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".name"); ok {
		name = v.(string)
	}

	if v, ok := d.GetOk(prefix + ".hostname"); ok {
		hostname = v.(string)
	}

	resourceSpec, err := expandInstanceGroupResourcesSpec(d, prefix+".resources")
	if err != nil {
		return nil, fmt.Errorf("Error create 'resources' object of api request: %s", err)
	}

	bootDiskSpec, err := expandInstanceGroupTemplateAttachedDiskSpec(d, prefix+".boot_disk.0", config)
	if err != nil {
		return nil, fmt.Errorf("Error create 'boot_disk' object of api request: %s", err)
	}

	secondaryDiskSpecs, err := expandInstanceGroupSecondaryDiskSpecs(d, prefix+".secondary_disk", config)
	if err != nil {
		return nil, fmt.Errorf("Error create 'secondary_disk' object of api request: %s", err)
	}

	nicSpecs, err := expandInstanceGroupNetworkInterfaceSpecs(d, prefix+".network_interface")
	if err != nil {
		return nil, fmt.Errorf("Error create 'network' object of api request: %s", err)
	}

	schedulingPolicy := expandInstanceGroupSchedulingPolicy(d, prefix+".scheduling_policy")
	placementPolicy := expandInstanceGroupPlacementPolicy(d, prefix+".placement_policy")

	labels, err := expandLabels(d.Get(prefix + ".labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance group: %s", err)
	}

	metadata, err := expandLabels(d.Get(prefix + ".metadata"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding metadata while creating instance group: %s", err)
	}

	networkSettings, err := expandInstanceGroupNetworkSettings(d.Get(prefix + ".network_settings.0.type"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding network settings while creating instance group: %s", err)
	}

	template := &instancegroup.InstanceTemplate{
		BootDiskSpec:          bootDiskSpec,
		Description:           description,
		Labels:                labels,
		Metadata:              metadata,
		NetworkInterfaceSpecs: nicSpecs,
		PlatformId:            platformId,
		ResourcesSpec:         resourceSpec,
		SchedulingPolicy:      schedulingPolicy,
		SecondaryDiskSpecs:    secondaryDiskSpecs,
		ServiceAccountId:      serviceAccount,
		NetworkSettings:       networkSettings,
		Name:                  name,
		Hostname:              hostname,
		PlacementPolicy:       placementPolicy,
	}

	return template, nil
}

func expandInstanceGroupVariables(v interface{}) ([]*instancegroup.Variable, error) {
	variables := make([]*instancegroup.Variable, 0)
	if v == nil {
		return variables, nil
	}

	for key, val := range v.(map[string]interface{}) {
		variables = append(variables, &instancegroup.Variable{
			Key:   key,
			Value: val.(string),
		})
	}
	return variables, nil
}

func parseInstanceGroupNetworkSettingsType(str string) (instancegroup.NetworkSettings_Type, error) {
	val, ok := instancegroup.NetworkSettings_Type_value[str]
	if !ok {
		return instancegroup.NetworkSettings_TYPE_UNSPECIFIED, fmt.Errorf("value for 'type' should be 'STANDARD' or 'SOFTWARE_ACCELERATED' or 'HARDWARE_ACCELERATED', not '%s'", str)
	}
	return instancegroup.NetworkSettings_Type(val), nil
}

func expandInstanceGroupAutoScale(d *schema.ResourceData, prefix string) (*instancegroup.ScalePolicy_AutoScale, error) {
	autoScale := &instancegroup.ScalePolicy_AutoScale{
		MinZoneSize: int64(d.Get(prefix + ".min_zone_size").(int)),
		MaxSize:     int64(d.Get(prefix + ".max_size").(int)),
		InitialSize: int64(d.Get(prefix + ".initial_size").(int)),
	}

	if v, ok := d.GetOk(prefix + ".measurement_duration"); ok {
		autoScale.MeasurementDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := d.GetOk(prefix + ".warmup_duration"); ok {
		autoScale.WarmupDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := d.GetOk(prefix + ".cpu_utilization_target"); ok {
		autoScale.CpuUtilizationRule = &instancegroup.ScalePolicy_CpuUtilizationRule{UtilizationTarget: v.(float64)}
	}

	if v, ok := d.GetOk(prefix + ".stabilization_duration"); ok {
		autoScale.StabilizationDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if customRulesCount := d.Get(prefix + ".custom_rule.#").(int); customRulesCount > 0 {
		rules := make([]*instancegroup.ScalePolicy_CustomRule, customRulesCount)
		for i := 0; i < customRulesCount; i++ {
			key := fmt.Sprintf(prefix+".custom_rule.%d", i)
			if rule, err := expandInstanceGroupCustomRule(d, key); err == nil {
				rules[i] = rule
			} else {
				return nil, err
			}
		}
		autoScale.CustomRules = rules
	}

	return autoScale, nil
}

func expandInstanceGroupDeployPolicy(d *schema.ResourceData) (*instancegroup.DeployPolicy, error) {
	policy := &instancegroup.DeployPolicy{
		MaxUnavailable: int64(d.Get("deploy_policy.0.max_unavailable").(int)),
		MaxDeleting:    int64(d.Get("deploy_policy.0.max_deleting").(int)),
		MaxCreating:    int64(d.Get("deploy_policy.0.max_creating").(int)),
		MaxExpansion:   int64(d.Get("deploy_policy.0.max_expansion").(int)),
	}

	if v, ok := d.GetOk("deploy_policy.0.startup_duration"); ok {
		policy.StartupDuration = &duration.Duration{Seconds: int64(v.(int))}
	}

	if v, ok := d.GetOk("deploy_policy.0.strategy"); ok {
		typeVal, ok := instancegroup.DeployPolicy_Strategy_value[strings.ToUpper(v.(string))]
		if !ok {
			return nil, fmt.Errorf("value for 'strategy' should be 'proactive' or 'opportunistic', not '%s'", v)
		}
		policy.Strategy = instancegroup.DeployPolicy_Strategy(typeVal)
	}

	return policy, nil
}

func expandInstanceGroupAllocationPolicy(d *schema.ResourceData) (*instancegroup.AllocationPolicy, error) {
	if v, ok := d.GetOk("allocation_policy.0.zones"); ok {
		zones := make([]*instancegroup.AllocationPolicy_Zone, 0)

		for _, s := range v.(*schema.Set).List() {
			zones = append(zones, &instancegroup.AllocationPolicy_Zone{ZoneId: s.(string)})
		}

		policy := &instancegroup.AllocationPolicy{Zones: zones}
		return policy, nil
	}

	return nil, fmt.Errorf("Zones must be defined")
}

func expandInstanceGroupHealthCheckSpec(d *schema.ResourceData) (*instancegroup.HealthChecksSpec, error) {
	checksCount := d.Get("health_check.#").(int)

	if checksCount == 0 {
		return nil, nil
	}

	checks := make([]*instancegroup.HealthCheckSpec, checksCount)

	for i := 0; i < checksCount; i++ {
		key := fmt.Sprintf("health_check.%d", i)
		hc := &instancegroup.HealthCheckSpec{
			HealthyThreshold:   int64(d.Get(key + ".healthy_threshold").(int)),
			UnhealthyThreshold: int64(d.Get(key + ".unhealthy_threshold").(int)),
		}
		if v, ok := d.GetOk(key + ".interval"); ok {
			hc.Interval = &duration.Duration{Seconds: int64(v.(int))}
		}
		if v, ok := d.GetOk(key + ".timeout"); ok {
			hc.Timeout = &duration.Duration{Seconds: int64(v.(int))}
		}
		checks[i] = hc

		if _, ok := d.GetOk(key + ".tcp_options"); ok {
			hc.HealthCheckOptions = &instancegroup.HealthCheckSpec_TcpOptions_{TcpOptions: &instancegroup.HealthCheckSpec_TcpOptions{Port: int64(d.Get(key + ".tcp_options.0.port").(int))}}
			continue
		}

		if _, ok := d.GetOk(key + ".http_options"); ok {
			hc.HealthCheckOptions = &instancegroup.HealthCheckSpec_HttpOptions_{
				HttpOptions: &instancegroup.HealthCheckSpec_HttpOptions{Port: int64(d.Get(key + ".http_options.0.port").(int)), Path: d.Get(key + ".http_options.0.path").(string)},
			}
			continue
		}

		return nil, fmt.Errorf("need tcp_options or http_options")
	}

	return &instancegroup.HealthChecksSpec{HealthCheckSpecs: checks}, nil
}

func expandInstanceGroupLoadBalancerSpec(d *schema.ResourceData) (*instancegroup.LoadBalancerSpec, error) {
	if _, ok := d.GetOk("load_balancer"); !ok {
		return nil, nil
	}

	spec := &instancegroup.TargetGroupSpec{
		Name:        d.Get("load_balancer.0.target_group_name").(string),
		Description: d.Get("load_balancer.0.target_group_description").(string),
	}

	if v, ok := d.GetOk("load_balancer.0.target_group_labels"); ok {
		labels, err := expandLabels(v)
		if err != nil {
			return nil, fmt.Errorf("Error expanding labels: %s", err)
		}

		spec.Labels = labels
	}

	return &instancegroup.LoadBalancerSpec{TargetGroupSpec: spec}, nil
}

func expandInstanceGroupPlacementPolicy(d *schema.ResourceData, prefix string) *instancegroup.PlacementPolicy {
	if v, ok := d.GetOk(prefix + ".0.placement_group_id"); ok {
		return &instancegroup.PlacementPolicy{PlacementGroupId: v.(string)}
	}
	return nil
}

func flattenInstanceGroupAttachedDisk(diskSpec *instancegroup.AttachedDiskSpec) (map[string]interface{}, error) {
	bootDisk := map[string]interface{}{
		"device_name": diskSpec.GetDeviceName(),
		"mode":        diskSpec.GetMode().String(),
	}

	diskSpecSpec := diskSpec.GetDiskSpec()

	if diskSpec == nil {
		return bootDisk, fmt.Errorf("no disk spec")
	}

	bootDisk["initialize_params"] = []map[string]interface{}{{
		"description": diskSpecSpec.Description,
		"size":        toGigabytes(diskSpecSpec.Size),
		"type":        diskSpecSpec.TypeId,
		"image_id":    diskSpecSpec.GetImageId(),
		"snapshot_id": diskSpecSpec.GetSnapshotId(),
	}}

	return bootDisk, nil
}

func parseInstanceGroupDiskMode(mode string) (instancegroup.AttachedDiskSpec_Mode, error) {
	val, ok := instancegroup.AttachedDiskSpec_Mode_value[mode]
	if !ok {
		return instancegroup.AttachedDiskSpec_MODE_UNSPECIFIED, fmt.Errorf("value for 'mode' should be 'READ_WRITE' or 'READ_ONLY', not '%s'", mode)
	}
	return instancegroup.AttachedDiskSpec_Mode(val), nil
}

func expandInstanceGroupNetworkSettings(v interface{}) (*instancegroup.NetworkSettings, error) {
	ns := &instancegroup.NetworkSettings{}
	if v == nil || v.(string) == "" {
		return nil, nil
	}
	t, err := parseInstanceGroupNetworkSettingsType(v.(string))
	if err != nil {
		return nil, err
	}
	ns.Type = t
	return ns, nil
}

func expandInstanceGroupScalePolicy(d *schema.ResourceData) (*instancegroup.ScalePolicy, error) {
	var policy = &instancegroup.ScalePolicy{}

	if _, ok := d.GetOk("scale_policy.0.fixed_scale"); ok {
		v := d.Get("scale_policy.0.fixed_scale.0.size").(int)
		policy.ScaleType = &instancegroup.ScalePolicy_FixedScale_{FixedScale: &instancegroup.ScalePolicy_FixedScale{Size: int64(v)}}
	}

	if _, ok := d.GetOk("scale_policy.0.auto_scale"); ok {
		autoScale, err := expandInstanceGroupAutoScale(d, "scale_policy.0.auto_scale.0")
		if err != nil {
			return nil, err
		}
		policy.ScaleType = &instancegroup.ScalePolicy_AutoScale_{AutoScale: autoScale}
		return policy, nil
	}

	if _, ok := d.GetOk("scale_policy.0.test_auto_scale"); ok {
		testAutoScale, err := expandInstanceGroupAutoScale(d, "scale_policy.0.test_auto_scale.0")
		if err != nil {
			return nil, err
		}
		policy.TestAutoScale = testAutoScale
		return policy, nil
	}

	if policy.ScaleType == nil {
		return nil, fmt.Errorf("Only fixed_scale and auto_scale policy are supported")
	}

	return policy, nil
}

func expandInstanceGroupSchedulingPolicy(d *schema.ResourceData, prefix string) *instancegroup.SchedulingPolicy {
	p := d.Get(prefix + ".0.preemptible").(bool)
	return &instancegroup.SchedulingPolicy{Preemptible: p}
}

func expandInstanceGroupCustomRule(d *schema.ResourceData, prefix string) (*instancegroup.ScalePolicy_CustomRule, error) {
	ruleType, ok := instancegroup.ScalePolicy_CustomRule_RuleType_value[d.Get(prefix+".rule_type").(string)]
	if !ok {
		return nil, fmt.Errorf("invalid value for rule_type")
	}

	metricType, ok := instancegroup.ScalePolicy_CustomRule_MetricType_value[d.Get(prefix+".metric_type").(string)]
	if !ok {
		return nil, fmt.Errorf("invalid value for metric_type")
	}

	labels, err := expandLabels(d.Get(prefix + ".labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating custom rule: %s", err)
	}

	return &instancegroup.ScalePolicy_CustomRule{
		RuleType:   instancegroup.ScalePolicy_CustomRule_RuleType(ruleType),
		MetricType: instancegroup.ScalePolicy_CustomRule_MetricType(metricType),
		MetricName: d.Get(prefix + ".metric_name").(string),
		Target:     d.Get(prefix + ".target").(float64),
		Labels:     labels,
	}, nil
}

func flattenInstanceGroupManagedInstances(instances []*instancegroup.ManagedInstance) ([]map[string]interface{}, error) {
	if instances == nil {
		return []map[string]interface{}{}, nil
	}

	res := make([]map[string]interface{}, len(instances))

	for i, instance := range instances {
		instDict := make(map[string]interface{})
		instDict["status"] = instance.GetStatus().String()
		instDict["instance_id"] = instance.GetInstanceId()
		instDict["fqdn"] = instance.GetFqdn()
		instDict["name"] = instance.GetName()
		instDict["status_message"] = instance.GetStatusMessage()
		instDict["zone_id"] = instance.GetZoneId()

		changedAt, err := getTimestamp(instance.GetStatusChangedAt())
		if err != nil {
			return res, err
		}
		instDict["status_changed_at"] = changedAt

		networkInterfaces, _, _, err := flattenInstanceGroupManagedInstanceNetworkInterfaces(instance)
		if err != nil {
			return res, err
		}

		instDict["network_interface"] = networkInterfaces
		res[i] = instDict
	}

	return res, nil
}
