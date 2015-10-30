package options

var (
	instanceTypes map[string]string
)

func init() {
	instanceTypes = map[string]string{
		"1_0.5":   "ecs.t1.xsmall",
		"1_1.0":   "ecs.t1.small",
		"1_2.0":   "ecs.s1.small",
		"1_4.0":   "ecs.s1.medium",
		"2_2.0":   "ecs.s2.small",
		"2_4.0":   "ecs.s2.large",
		"2_8.0":   "ecs.s2.xlarge",
		"4_4.0":   "ecs.s3.medium",
		"4_8.0":   "ecs.s3.large",
		"4_16.0":  "ecs.m1.medium",
		"8_32.0":  "ecs.m1.xlarge",
		"8_8.0":   "ecs.c1.small",
		"8_16.0":  "ecs.c1.large",
		"16_64.0": "ecs.c2.xlarge",
		"1_8.0":   "ecs.s1.large",
		"2_16.0":  "ecs.s2.2xlarge",
		"4_32.0":  "ecs.m2.medium",
	}
}

// func (od *OptsDriver) GetAliyunOptions(md dcluster.MachineDescription) ([]string, error) {
// 	// TODO
// 	options := []string{}
// 	instanceType, exists := getAliyunInstanceType(md)
// 	if !exists {
// 		return []string{}, errors.New(fmt.Sprintf("No aliyun instance type matched for cpu:%s, mem:%s . Go 'https://gist.github.com/Lax/3a2037a11c49df1aa1e7' for detail", md.Cpu, md.Memory))
// 	}
// 	// instance type
// 	options = append(options, "--aliyun-instance-type-id")
// 	options = append(options, instanceType)

// 	// region
// 	if md.Region == "" {
// 		return []string{}, errors.New(fmt.Sprintf("Aliyun region must be provided."))
// 	}
// 	options = append(options, "--aliyun-region-id")
// 	options = append(options, md.Region)
// 	return options, nil
// }

// func getAliyunInstanceType(md dcluster.MachineDescription) (string, bool) {
// 	cpu := md.GetCpu()
// 	memInGB := float64(md.GetMemInBytes()) / float64(1024) / float64(1024) / float64(1024)
// 	key := fmt.Sprintf("%d_%.1f", cpu, memInGB)
// 	it, exists := instanceTypes[key]
// 	return it, exists
// }
