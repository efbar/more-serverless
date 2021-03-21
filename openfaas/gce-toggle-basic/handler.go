package function

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	compute "google.golang.org/api/compute/v1"
	option "google.golang.org/api/option"
)

// Handle a serverless request
func Handle(req []byte) string {

	ctx := context.Background()

	projectId := os.Getenv("PROJECT_ID")
	projectRegion := os.Getenv("REGION")

	computeService, err := compute.NewService(ctx, option.WithCredentialsFile("/var/openfaas/secrets/sa-gcp"))
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Sprintln(err.Error())
	}

	region, err := computeService.Regions.Get(projectId, projectRegion).Do()
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Sprintln(err.Error())
	}

	var instanceId uint64
	out := []string{}

	for _, val := range region.Zones {
		instances, _ := computeService.Instances.List(projectId, val[strings.LastIndex(val, "/")+1:]).Do()
		for _, v := range instances.Items {
			out = append(out, "Status of "+v.Name+" is "+v.Status+"\n")
			instanceId = v.Id
			if v.Status == "TERMINATED" || v.Status == "STOPPED" {
				started, err := computeService.Instances.Start(projectId, v.Zone[strings.LastIndex(v.Zone, "/")+1:], strconv.FormatUint(instanceId, 10)).Do()
				if err != nil {
					fmt.Println(err.Error())
					return fmt.Sprintln(err.Error())
				}
				if started.HTTPStatusCode == 200 {
					out = append(out, "turning "+v.Name+" ON!\n")
				}
			} else {
				stopped, err := computeService.Instances.Stop(projectId, v.Zone[strings.LastIndex(v.Zone, "/")+1:], strconv.FormatUint(instanceId, 10)).Do()
				if err != nil {
					fmt.Println(err.Error())
					return fmt.Sprintln(err.Error())
				}
				if stopped.HTTPStatusCode == 200 {
					out = append(out, "turning "+v.Name+" OFF!\n")
				}
			}
		}
	}

	instanceGroupList, err := computeService.RegionInstanceGroupManagers.List(projectId, projectRegion).Do()
	if err != nil {
		fmt.Println(err.Error())
		return fmt.Sprintln(err.Error())
	}

	for _, v := range instanceGroupList.Items {
		if v.TargetSize != 0 {
			instanceGroup, err := computeService.RegionInstanceGroupManagers.Resize(projectId, projectRegion, v.Name, 0).Do()
			if err != nil {
				fmt.Println(err.Error())
				return fmt.Sprintln(err.Error())
			}
			if instanceGroup.HTTPStatusCode == 200 {
				out = append(out, "Scaling "+v.Name+" down to zero!\n")
			}
		} else {
			instanceGroup, err := computeService.RegionInstanceGroupManagers.Resize(projectId, projectRegion, v.Name, 3).Do()
			if err != nil {
				fmt.Println(err.Error())
				return fmt.Sprintln(err.Error())
			}
			if instanceGroup.HTTPStatusCode == 200 {
				out = append(out, "Scaling "+v.Name+" up to one!\n")
			}
		}
	}

	return strings.Join(out, "")
}
