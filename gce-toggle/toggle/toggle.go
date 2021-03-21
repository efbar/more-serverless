package toggle

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func Toggle(w http.ResponseWriter, r *http.Request) {

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

	if len(input) != 0 {
		fmt.Printf("request body: %s", string(input))
	} else {
		fmt.Println("empty body")
	}

	ctx := context.Background()

	projectId := os.Getenv("PROJECT_ID")
	projectRegion := os.Getenv("REGION")

	var computeService *compute.Service
	var err error
	if len(input) != 0 {
		computeService, err = compute.NewService(ctx, option.WithCredentialsFile(string(input)))
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
					out = append(out, ", turning "+v.Name+" ON! |")
				}
			} else {
				stopped, err := computeService.Instances.Stop(projectId, v.Zone[strings.LastIndex(v.Zone, "/")+1:], strconv.FormatUint(instanceId, 10)).Do()
				if err != nil {
					fmt.Println(err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if stopped.HTTPStatusCode == 200 {
					out = append(out, ", turning "+v.Name+" OFF! |")
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
				out = append(out, " Scaling "+v.Name+" DOWN to zero instances! |")
			}
		} else {
			instanceGroup, err := computeService.RegionInstanceGroupManagers.Resize(projectId, projectRegion, v.Name, 3).Do()
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if instanceGroup.HTTPStatusCode == 200 {
				out = append(out, " Scaling "+v.Name+" UP to three instances!")
			}
		}
	}

	response := struct {
		Payload     string              `json:"payload"`
		Headers     map[string][]string `json:"headers"`
		Environment []string            `json:"environment"`
	}{
		Payload:     strings.Join(out, ""),
		Headers:     r.Header,
		Environment: os.Environ(),
	}

	resBody, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resBody)

	return
}
