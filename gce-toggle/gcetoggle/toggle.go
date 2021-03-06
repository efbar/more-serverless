package toggle

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func Serve(w http.ResponseWriter, r *http.Request) {

	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		reqBody, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		input = reqBody
	}

	if len(input) == 0 {
		fmt.Println("empty body")
	}

	ctx := context.Background()

	projectId := os.Getenv("PROJECT_ID")
	projectRegion := os.Getenv("REGION")

	var err error
	var secret string
	if _, err := os.Stat("/var/openfaas/secrets/gce-sa-gcp"); err == nil {
		secretFile, _ := ioutil.ReadFile("/var/openfaas/secrets/gce-sa-gcp")
		secret = string(secretFile)
	}

	var computeService *compute.Service
	if len(secret) != 0 {
		computeService, err = compute.NewService(ctx, option.WithCredentialsFile("/var/openfaas/secrets/gce-sa-gcp"))
	} else if len(input) != 0 {
		computeService, err = compute.NewService(ctx, option.WithCredentialsJSON(input))
	} else {
		computeService, err = compute.NewService(ctx)
	}
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	region, err := computeService.Regions.Get(projectId, projectRegion).Do()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var instanceId uint64
	out := []string{}

	for _, val := range region.Zones {
		instances, _ := computeService.Instances.List(projectId, val[strings.LastIndex(val, "/")+1:]).Do()
		for _, v := range instances.Items {
			out = append(out, "Status of "+v.Name+" is "+v.Status)
			instanceId = v.Id
			if v.Status == "TERMINATED" || v.Status == "STOPPED" {
				started, err := computeService.Instances.Start(projectId, v.Zone[strings.LastIndex(v.Zone, "/")+1:], strconv.FormatUint(instanceId, 10)).Do()
				if err != nil {
					fmt.Println(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if started.HTTPStatusCode == 200 {
					out = append(out, ", turning "+v.Name+" ON!\n")
				}
			} else {
				stopped, err := computeService.Instances.Stop(projectId, v.Zone[strings.LastIndex(v.Zone, "/")+1:], strconv.FormatUint(instanceId, 10)).Do()
				if err != nil {
					fmt.Println(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if stopped.HTTPStatusCode == 200 {
					out = append(out, ", turning "+v.Name+" OFF!\n")
				}
			}
		}
	}

	instanceGroupList, err := computeService.RegionInstanceGroupManagers.List(projectId, projectRegion).Do()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, v := range instanceGroupList.Items {
		if v.TargetSize != 0 {
			instanceGroup, err := computeService.RegionInstanceGroupManagers.Resize(projectId, projectRegion, v.Name, 0).Do()
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if instanceGroup.HTTPStatusCode == 200 {
				out = append(out, "Scaling "+v.Name+" DOWN to zero instances!\n")
			}
		} else {
			instanceGroup, err := computeService.RegionInstanceGroupManagers.Resize(projectId, projectRegion, v.Name, 3).Do()
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if instanceGroup.HTTPStatusCode == 200 {
				out = append(out, "Scaling "+v.Name+" UP to three instances!\n")
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "text/plain")
	w.Write([]byte(strings.Join(out, "")))

}
