package vaultstatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	vault "github.com/hashicorp/vault/api"
	"github.com/ryanuber/columnize"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type Response struct {
	Payload EnrichedStatus      `json:"payload"`
	Headers map[string][]string `json:"headers"`
}

type EnrichedStatus struct {
	Type                     string `json:"seal_type"`
	Initialized              string `json:"initialized"`
	Sealed                   string `json:"sealed"`
	T                        string `json:"total_recovery_shares"`
	N                        string `json:"threshold"`
	Progress                 string `json:"unseal_progress,omitempty"`
	Nonce                    string `json:"unseal_nonce,omitempty"`
	Version                  string `json:"version"`
	Migration                string `json:"migration,omitempty"`
	ClusterName              string `json:"cluster_name,omitempty"`
	ClusterID                string `json:"cluster_id,omitempty"`
	RecoverySeal             string `json:"recovery_seal,omitempty"`
	StorageType              string `json:"storage_type,omitempty"`
	HAEnabled                string `json:"ha_enabled"`
	HAMode                   string `json:"ha_mode,omitempty"`
	HACluster                string `json:"ha_cluster,omitempty"`
	IsSelf                   string `json:"is_self,omitempty"`
	ActiveTime               string `json:"active_time,omitempty"`
	LeaderAddress            string `json:"leader_address,omitempty"`
	LeaderClusterAddress     string `json:"leader_cluster_address,omitempty"`
	PerfStandby              string `json:"performance_standby,omitempty"`
	PerfStandbyLastRemoteWAL string `json:"performance_standby_last_remote_wal,omitempty"`
	LastWAL                  string `json:"last_wal,omitempty"`
	RaftCommittedIndex       string `json:"raft_committed_index,omitempty"`
	RaftAppliedIndex         string `json:"raft_applied_index,omitempty"`
}

func Serve(w http.ResponseWriter, r *http.Request) {

	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	rb := RequestBody{}
	json.Unmarshal(input, &rb)

	vaultEndpoint := "http://localhost:8200"
	if rb.Endpoint != "" {
		vaultEndpoint = rb.Endpoint
	}

	conf := vault.DefaultConfig()

	conf.Address = vaultEndpoint

	client, err := vault.NewClient(conf)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rb.Token != "" {
		client.SetToken(rb.Token)
	}

	status, err := client.Sys().SealStatus()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	leaderStatus, err := client.Sys().Leader()
	if err != nil && strings.Contains(err.Error(), "Vault is sealed") {
		leaderStatus = &api.LeaderResponse{HAEnabled: true}
		err = nil
	}
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"Key\tValue", "---\t-----"}

		v := reflect.ValueOf(*status)

		for i := 0; i < v.NumField(); i++ {

			if v.Type().Field(i).Name == "Type" {
				var sealPrefix string
				if status.RecoverySeal {
					sealPrefix = "Recovery "
				}
				out = append(out, fmt.Sprintf("%sSeal Type\t%v\n", sealPrefix, v.Field(i).Interface()))
			} else if v.Type().Field(i).Name == "T" {
				out = append(out, fmt.Sprintf("%s\t%v\n", "Total Recovery Shares", v.Field(i).Interface()))
			} else if v.Type().Field(i).Name == "N" {
				out = append(out, fmt.Sprintf("%s\t%v\n", "Threshold", v.Field(i).Interface()))
			} else {
				if v.Type().Field(i).Name != "Progress" &&
					v.Type().Field(i).Name != "Nonce" &&
					v.Type().Field(i).Name != "RecoverySeal" &&
					v.Type().Field(i).Name != "StorageType" &&
					v.Type().Field(i).Name != "ClusterName" &&
					v.Type().Field(i).Name != "ClusterID" {
					out = append(out, fmt.Sprintf("%s\t%v\n", v.Type().Field(i).Name, v.Field(i).Interface()))
				}
			}
		}

		if status.Sealed {
			out = append(out, fmt.Sprintf("Unseal Progress\t%d/%d", status.Progress, status.T))
			out = append(out, fmt.Sprintf("Unseal Nonce\t%s", status.Nonce))
		}

		if status.Migration {
			out = append(out, fmt.Sprintf("Seal Migration in Progress\t%t", status.Migration))
		}
		out = append(out, fmt.Sprintf("Storage Type\t%s", status.StorageType))

		if status.ClusterName != "" && status.ClusterID != "" {
			out = append(out, fmt.Sprintf("Cluster Name\t%s", status.ClusterName))
			out = append(out, fmt.Sprintf("Cluster ID\t%s", status.ClusterID))
		}

		out = append(out, fmt.Sprintf("HA Enabled\t%t", leaderStatus.HAEnabled))

		if leaderStatus.HAEnabled {
			mode := "sealed"
			if !status.Sealed {
				out = append(out, fmt.Sprintf("HA Cluster\t%s", leaderStatus.LeaderClusterAddress))
				mode = "standby"
				showLeaderAddr := false
				if leaderStatus.IsSelf {
					mode = "active"
				} else {
					if leaderStatus.LeaderAddress == "" {
						leaderStatus.LeaderAddress = "<none>"
					}
					showLeaderAddr = true
				}
				out = append(out, fmt.Sprintf("HA Mode\t%s", mode))

				if leaderStatus.IsSelf && !leaderStatus.ActiveTime.IsZero() {
					out = append(out, fmt.Sprintf("Active Since\t%s", leaderStatus.ActiveTime.Format(time.RFC3339Nano)))
				}

				if showLeaderAddr {
					out = append(out, fmt.Sprintf("Active Node Address\t%s", leaderStatus.LeaderAddress))
				}

				if leaderStatus.PerfStandby {
					out = append(out, fmt.Sprintf("Performance Standby Node\t%t", leaderStatus.PerfStandby))
					out = append(out, fmt.Sprintf("Performance Standby Last Remote WAL\t%d", leaderStatus.PerfStandbyLastRemoteWAL))
				}
			}
		}

		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "    "
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var out EnrichedStatus

		if status.RecoverySeal {
			out.RecoverySeal = "recovery"
		}

		if leaderStatus.HAEnabled {
			mode := "sealed"
			if !status.Sealed {
				out.LeaderClusterAddress = leaderStatus.LeaderClusterAddress
				mode = "standby"
				showLeaderAddr := false
				if leaderStatus.IsSelf {
					mode = "active"
				} else {
					if leaderStatus.LeaderAddress == "" {
						leaderStatus.LeaderAddress = "<none>"
					}
					showLeaderAddr = true
				}
				out.HAMode = mode

				if leaderStatus.IsSelf && !leaderStatus.ActiveTime.IsZero() {
					out.ActiveTime = leaderStatus.ActiveTime.Format(time.RFC3339Nano)
				}

				if showLeaderAddr {
					out.LeaderAddress = leaderStatus.LeaderAddress
				}

				if leaderStatus.PerfStandby {
					out.PerfStandby = fmt.Sprintf("%t", leaderStatus.PerfStandby)
					out.PerfStandbyLastRemoteWAL = fmt.Sprintf("%d", leaderStatus.PerfStandbyLastRemoteWAL)
				}
			}
		}

		out.Type = status.Type
		out.Initialized = fmt.Sprintf("%v", status.Initialized)
		out.Sealed = fmt.Sprintf("%v", status.Sealed)
		out.Version = fmt.Sprintf("%v", status.Version)
		out.T = fmt.Sprintf("%v", status.T)
		out.N = fmt.Sprintf("%v", status.N)
		out.HAEnabled = strconv.FormatBool(leaderStatus.HAEnabled)
		out.Migration = fmt.Sprintf("%t", status.Migration)
		out.IsSelf = strconv.FormatBool(leaderStatus.IsSelf)
		out.ActiveTime = leaderStatus.ActiveTime.Format(time.RFC3339Nano)
		out.LeaderAddress = leaderStatus.LeaderAddress
		out.LeaderClusterAddress = leaderStatus.LeaderClusterAddress
		out.PerfStandby = strconv.FormatBool(leaderStatus.PerfStandby)
		out.PerfStandbyLastRemoteWAL = strconv.FormatUint(leaderStatus.PerfStandbyLastRemoteWAL, 10)
		out.LastWAL = strconv.FormatUint(leaderStatus.LastWAL, 10)
		out.RaftCommittedIndex = strconv.FormatUint(leaderStatus.RaftCommittedIndex, 10)
		out.RaftAppliedIndex = strconv.FormatUint(leaderStatus.RaftAppliedIndex, 10)

		jsonResponse := Response{
			Payload: out,
			Headers: r.Header,
		}
		resBody, err := json.Marshal(jsonResponse)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(resBody)
	}

}
