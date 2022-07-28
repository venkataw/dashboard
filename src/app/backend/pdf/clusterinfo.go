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
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/logs"
	ns "github.com/kubernetes/dashboard/src/app/backend/resource/namespace"
	"github.com/kubernetes/dashboard/src/app/backend/resource/node"
	"github.com/kubernetes/dashboard/src/app/backend/resource/persistentvolumeclaim"
	"github.com/kubernetes/dashboard/src/app/backend/resource/pod"
)

var client http.Client
var bearerToken string

func initializeHttpClient() {
	client = http.Client{}
}
func setBearerToken(token string) {
	bearerToken = token
}

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
func namespaceExists(namespace string) (nsExists bool, err error) {
	resp, err := getHttp("namespace/")
	if err != nil {
		log.Printf("Error getting namespace list, error: %v", err)
		return false, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of namespace list, error: %v", err)
		return false, err
	}
	var namespaces ns.NamespaceList = ns.NamespaceList{}
	err = json.Unmarshal(bodyBytes, &namespaces)
	if err != nil {
		log.Printf("Error parsing json of namespace list, error: %v", err)
		return false, err
	}
	for _, ns := range namespaces.Namespaces {
		if ns.ObjectMeta.Name == namespace {
			return true, nil
		}
	}
	return false, errors.New("Could not find namespace " + namespace)
}
func getEvents(namespace string) (events common.EventList, err error) {
	resp, err := getHttp("event/" + namespace)
	if err != nil {
		log.Printf("Error getting event list for namespace %s, error: %v", namespace, err)
		return common.EventList{}, err
	}
	bodyBytes, err := parseHtmlToBytes(resp)
	if err != nil {
		log.Printf("Error parsing html of event list in namespace %s, error: %v", namespace, err)
		return common.EventList{}, err
	}

	var eventList common.EventList = common.EventList{}
	err = json.Unmarshal(bodyBytes, &eventList)
	if err != nil {
		log.Printf("Error parsing json of event list in namespace %s, error: %v", namespace, err)
		return common.EventList{}, err
	}

	return eventList, nil
}
func getPvcEvents(namespace, pvcName string) (events []common.Event, err error) {
	eventList, err := getEvents(namespace)
	if err != nil {
		log.Printf("Error getting events, cannot get pvc events. Error: %v", err)
		return []common.Event{}, err
	}
	events = make([]common.Event, 0)
	for _, event := range eventList.Events {
		if event.SubObjectKind == "PersistentVolumeClaim" { // TODO: test this
			if event.SubObjectName == pvcName {
				events = append(events, event)
			}
		}
	}
	if len(events) == 0 {
		return []common.Event{}, nil // no events
	}
	return events, nil
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
	request, err := http.NewRequest("GET", getProtocol()+"localhost:"+fmt.Sprint(ApiPort)+"/api/v1/"+path, nil)
	if err != nil {
		log.Printf("Error creating request for %s, error: %v", path, err)
		return nil, err
	}
	if bearerToken != "" {
		// add bearer token if exists
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error doing request for %s, error: %v", path, err)
		return nil, err
	}
	return response, nil
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
