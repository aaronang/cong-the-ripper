// functions (not methods) related to AWS go here

package master

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/aaronang/cong-the-ripper/lib"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// createSlaves creates a new slave instance.
func createSlaves(svc *ec2.EC2, count int) ([]*ec2.Instance, error) {
	params := &ec2.RunInstancesInput{
		ImageId:      aws.String(lib.SlaveImage),
		InstanceType: aws.String(lib.SlaveType),
		MinCount:     aws.Int64(int64(count)),
		MaxCount:     aws.Int64(int64(count)),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(lib.SlaveARN),
		},
		SecurityGroupIds: []*string{aws.String("sg-646fbb02")},
	}
	resp, err := svc.RunInstances(params)
	return resp.Instances, err
}

// terminateSlaves terminates a slave instance.
func terminateSlaves(svc *ec2.EC2, instances []*ec2.Instance) (*ec2.TerminateInstancesOutput, error) {
	params := &ec2.TerminateInstancesInput{
		InstanceIds: instanceIds(instances),
	}
	return svc.TerminateInstances(params)
}

// sendTask sends a task to a slave instance.
func sendTask(t *lib.Task, addr string) (*http.Response, error) {
	url := lib.Protocol + addr + lib.TasksCreatePath
	body, err := t.ToJSON()
	if err != nil {
		panic(err)
	}
	return http.Post(url, lib.BodyType, bytes.NewBuffer(body))
}

func newEC2() *ec2.EC2 {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(lib.AWSRegion)},
	)
	if err != nil {
		panic(err)
	}
	return ec2.New(sess)
}

func instanceIds(instances []*ec2.Instance) []*string {
	instanceIds := make([]*string, len(instances))
	for i, instance := range instances {
		instanceIds[i] = instance.InstanceId
	}
	return instanceIds
}

func getPublicIP(svc *ec2.EC2, instance *ec2.Instance) *string {
	params := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("instance-state-name"),
				Values: []*string{
					aws.String("pending"),
					aws.String("running"),
				},
			},
		},
		InstanceIds: []*string{
			instance.InstanceId,
		},
	}

	var i int
	for {
		res, err := svc.DescribeInstances(&params)

		// ignore the error because we may try again
		if err == nil &&
			len(res.Reservations) == 1 &&
			len(res.Reservations[0].Instances) == 1 {

			if res.Reservations[0].Instances[0].PublicIpAddress != nil {
				return res.Reservations[0].Instances[0].PublicIpAddress
			}

		}
		time.Sleep(10 * time.Second)
		i++
		if i > 12 {
			log.Println("Unable to find public IP")
			return nil
		}
	}
}

func slavesToInstances(slaves []slave) []*ec2.Instance {
	res := make([]*ec2.Instance, len(slaves))
	for i := range slaves {
		res[i] = slaves[i].instance
	}
	return res
}

func slavesMapToInstances(slaves map[string]slave) []*ec2.Instance {
	res := make([]*ec2.Instance, len(slaves))
	i := 0
	for _, v := range slaves {
		res[i] = v.instance
	}
	return res
}
