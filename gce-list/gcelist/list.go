package list

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ryanuber/columnize"
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

	var computeService *compute.Service
	var err error

	var secret string

	if _, err := os.Stat("/var/openfaas/secrets/gce-sa-gcp"); err == nil {
		secretFile, _ := ioutil.ReadFile("/var/openfaas/secrets/gce-sa-gcp")
		secret = string(secretFile)
	}

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

	out := []string{"NAME\tZONE\tMACHINE_TYPE\tPREEMPTIBLE\tINTERNAL_IP\tEXTERNAL_IP\tSTATUS"}
	for _, val := range region.Zones {
		instances, _ := computeService.Instances.List(projectId, val[strings.LastIndex(val, "/")+1:]).Do()
		for _, v := range instances.Items {
			zones := strings.Split(v.Zone, "/")
			mType := strings.Split(v.MachineType, "/")
			out = append(out, v.Name+"\t"+zones[len(zones)-1]+"\t"+mType[len(mType)-1]+"\t"+strconv.FormatBool(v.Scheduling.Preemptible)+"\t"+v.NetworkInterfaces[0].NetworkIP+"\t"+v.NetworkInterfaces[0].AccessConfigs[0].NatIP+"\t"+v.Status)
		}
	}
	columnConf := columnize.DefaultConfig()
	columnConf.Delim = "\t"
	columnConf.Glue = "  "
	columnConf.NoTrim = false
	resBody := columnize.Format(out, columnConf)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "text/plain")
	w.Write([]byte(resBody))

}
