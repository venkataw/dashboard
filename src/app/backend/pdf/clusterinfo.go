// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pdf

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/logs"
	"github.com/kubernetes/dashboard/src/app/backend/resource/node"
	"github.com/kubernetes/dashboard/src/app/backend/resource/persistentvolumeclaim"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
)

func getPodDetail(namespace string) (response pod.PodList, err error) {
	resp, err := getHttp("pod/" + namespace)
	if err != nil {
		log.Printf("Error getting pod detail in namespace %s, error: %v", namespace, err)
		return pod.PodList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod detail in namespace %s, error: %v", namespace, err)
		return pod.PodList{}, err
	}

	var detail pod.PodList = pod.PodList{}
	err = json.Unmarshal(bodyBytes, &detail)
	if err != nil {
		log.Printf("Error parsing json of pod detail in namespace %s, error: %v", namespace, err)
		return pod.PodList{}, err
	}

	return detail, nil
}
func getPodLogs(namespace string, pod string) (response logs.LogDetails, err error) {
	resp, err := getHttp("log/" + namespace + "/" + pod)
	if err != nil {
		log.Printf("Error getting pod logs in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod logs in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}

	var logDetails logs.LogDetails = logs.LogDetails{}
	err = json.Unmarshal(bodyBytes, &logDetails)
	if err != nil {
		log.Printf("Error parsing json of pod logs in namespace %s, error: %v", namespace, err)
		return logs.LogDetails{}, err
	}

	return logDetails, nil
}
func getPodEvents(namespace string, pod string) (response common.EventList, err error) {
	resp, err := getHttp("pod/" + namespace + "/" + pod + "/event")
	if err != nil {
		log.Printf("Error getting pod events in namespace %s for pod %s, error: %v", namespace, pod, err)
		return common.EventList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod events in namespace %s for pod %s, error: %v", namespace, pod, err)
		return common.EventList{}, err
	}

	var events common.EventList = common.EventList{}
	err = json.Unmarshal(bodyBytes, &events)
	if err != nil {
		log.Printf("Error parsing json of pod logs in namespace %s for pod %s, error: %v", namespace, pod, err)
		return common.EventList{}, err
	}

	return events, nil
}
func getPodPvc(namespace string, pod string) (response persistentvolumeclaim.PersistentVolumeClaimList, err error) {
	resp, err := getHttp("pod/" + namespace + "/" + pod + "/persistentvolumeclaim")
	if err != nil {
		log.Printf("Error getting pod pvc in namespace %s for pod %s, error: %v", namespace, pod, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pod pvc in namespace %s for pod %s, error: %v", namespace, pod, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}

	var pvcList persistentvolumeclaim.PersistentVolumeClaimList = persistentvolumeclaim.PersistentVolumeClaimList{}
	err = json.Unmarshal(bodyBytes, &pvcList)
	if err != nil {
		log.Printf("Error parsing json of pod pvc in namespace %s for pod %s, error: %v", namespace, pod, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}

	return pvcList, nil
}
func getNodeList() (nodes node.NodeList, err error) {
	resp, err := getHttp("node")
	if err != nil {
		log.Printf("Error getting node list, error: %v", err)
		return node.NodeList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of node list, error: %v", err)
		return node.NodeList{}, err
	}

	var nodeDetails node.NodeList = node.NodeList{}
	err = json.Unmarshal(bodyBytes, &nodeDetails)
	if err != nil {
		log.Printf("Error parsing json of node list, error: %v", err)
		return node.NodeList{}, err
	}

	return nodeDetails, nil
}
func getNodeDetail(nodeName string) (nodeDetail node.NodeDetail, err error) {
	resp, err := getHttp("node/" + nodeName)
	if err != nil {
		log.Printf("Error getting node detail for node %s, error: %v", nodeName, err)
		return node.NodeDetail{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of node detail %s, error: %v", nodeName, err)
		return node.NodeDetail{}, err
	}

	var nodeDetails node.NodeDetail = node.NodeDetail{}
	err = json.Unmarshal(bodyBytes, &nodeDetails)
	if err != nil {
		log.Printf("Error parsing json of node detail %s, error: %v", nodeName, err)
		return node.NodeDetail{}, err
	}

	return nodeDetails, nil
}
func getPvcDetail(namespace string) (claimList persistentvolumeclaim.PersistentVolumeClaimList, err error) {
	resp, err := getHttp("persistentvolumeclaim/" + namespace)
	if err != nil {
		log.Printf("Error getting pvc list for namespace %s, error: %v", namespace, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of pvc list in namespace %s, error: %v", namespace, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}

	var pvcList persistentvolumeclaim.PersistentVolumeClaimList = persistentvolumeclaim.PersistentVolumeClaimList{}
	err = json.Unmarshal(bodyBytes, &pvcList)
	if err != nil {
		log.Printf("Error parsing json of pvc list in namespace %s, error: %v", namespace, err)
		return persistentvolumeclaim.PersistentVolumeClaimList{}, err
	}

	return pvcList, nil
}

// Helper http/s functions
func getProtocol() string {
	if Secure {
		return "https://"
	} else {
		return "http://"
	}
}
func getHttp(path string) (resp *http.Response, err error) {
	return http.Get(getProtocol() + "localhost:" + fmt.Sprint(ApiPort) + "/api/v1/" + path)
}
func parseHtmlToBytes(response *http.Response) (result []byte, err error) {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error reading http body, error: %v", err)
		return nil, err
	}
	bodyBytes := []byte(body)
	return bodyBytes, nil
}
