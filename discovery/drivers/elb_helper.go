package drivers

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

// https://github.com/aws/aws-sdk-go/blob/master/service/elb/examples_test.go

func ELB_Create(region, elbName, protocal string, port int64) (DNSName string, err error) {
	/*
		auth, err := aws.EnvAuth()
		if err != nil {
			c.Fatal(err)
		}
	*/

	//svc := elb.New(auth, &aws.Config{Region: aws.String(region)}) //aws.USEast)

	svc := elb.New(&aws.Config{Region: aws.String(region)}) // need region

	params := &elb.CreateLoadBalancerInput{
		LoadBalancerName: aws.String(elbName),
		AvailabilityZones: []*string{
			aws.String(region + "a"),
			aws.String(region + "b"),
		},
		Listeners: []*elb.Listener{
			{
				InstancePort:     aws.Int64(port),
				InstanceProtocol: aws.String(protocal),
				LoadBalancerPort: aws.Int64(port),
				Protocol:         aws.String(protocal),
			},
		},
		/*
			SecurityGroups: []*string{
				aws.String("SecurityGroupId"), // Required
				// More values...
			},
			Subnets: []*string{
				aws.String("SubnetId"), // Required
				// More values...
			},
		*/
	}
	resp, err := svc.CreateLoadBalancer(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		return "", err
	}

	// Pretty-print the response data.
	return *resp.DNSName, nil
}

func ELB_Delete(region, elbName string) error {
	svc := elb.New(&aws.Config{Region: aws.String(region)})
	params := &elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(elbName), // Required
	}
	// If the load balancer does not exist or has already been
	// deleted, the call to DeleteLoadBalancer still succeeds.
	_, err := svc.DeleteLoadBalancer(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		return err
	}

	return nil
}

func ELB_Exist(region, elbName string) bool {
	svc := elb.New(&aws.Config{Region: aws.String(region)})
	params := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{
			aws.String(elbName), // Required
			// More values...
		},
		/*
			Marker:   aws.String("Marker"),
			PageSize: aws.Int64(1),
		*/
	}
	resp, err := svc.DescribeLoadBalancers(params)

	if err != nil {
		return false
	}

	return len(resp.LoadBalancerDescriptions) > 0
}

func ELB_Register(region, elbName, instanceId string) {
	svc := elb.New(&aws.Config{Region: aws.String(region)}) // need region

	params := &elb.RegisterInstancesWithLoadBalancerInput{
		Instances: []*elb.Instance{ // Required
			{ // Required
				InstanceId: aws.String(instanceId), //"i-e5416820"),
			},
			// More values...
		},
		LoadBalancerName: aws.String(elbName), // Required
	}
	resp, err := svc.RegisterInstancesWithLoadBalancer(params)
	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}
	// Pretty-print the response data.
	fmt.Println(resp)

}

func ELB_Deregister(region, elbName, instanceId string) {
	svc := elb.New(&aws.Config{Region: aws.String(region)}) // need region

	params := &elb.DeregisterInstancesFromLoadBalancerInput{
		Instances: []*elb.Instance{ // Required
			{ // Required
				InstanceId: aws.String(instanceId),
			},
			// More values...
		},
		LoadBalancerName: aws.String(elbName), // Required
	}
	resp, err := svc.DeregisterInstancesFromLoadBalancer(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)

}

/*
func main() {
	//Register("us-west-2", "test", "i-fb41683e")
	DNSName, err := ELB_Create("cn-north-1", "liubin-testlb", "tcp", 80)
	if err != nil {
		fmt.Println("ERR:", err)
	} else {
		fmt.Println("DNSName:", DNSName)
	}

	DNSName, err = ELB_Create("cn-north-1", "liubin-testlb-a", "http", 81)
	if err != nil {
		fmt.Println("ERR:", err)
	} else {
		fmt.Println("DNSName:", DNSName)
	}

	ELB_Delete("cn-north-1", "liubin-testlb")
	ELB_Delete("cn-north-1", "liubin-testlb-a")
	fmt.Println(ELB_Exist("cn-north-1", "liubin-testlb"))
	//Deregister("us-west-2", "test", "i-fb41683e")
	//Deregister("us-west-2", "test", "i-e5416820")
}
*/
