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
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/persistentvolumeclaim"
	v1 "k8s.io/api/core/v1"
)

func formatLabelString(labels map[string]string) string {
	var formatted string = ""
	if len(labels) == 0 {
		return "No labels"
	}
	for key, value := range labels {
		formatted += key + ":" + value + ", "
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
func formatTaintString(taints []v1.Taint) string {
	var formatted string = ""
	if len(taints) == 0 {
		return "No taints"
	}
	for _, taint := range taints {
		formatted += taint.Key + ":" + taint.Value + ", "
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
func formatInternalIpString(addresses []v1.NodeAddress) string {
	var formatted string = ""
	if len(addresses) == 0 {
		return "No addresses"
	}
	for _, address := range addresses {
		if address.Type == "InternalIP" {
			formatted += address.Address + ", "
		}
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
func formatEventListArray(events common.EventList) []string {
	return formatEventArray(events.Events)
}
func formatEventArray(events []common.Event) []string {
	eventArr := make([]string, len(events))
	for i, event := range events {
		eventArr[i] = event.Message + ", Reason: " + event.Reason // TODO: ADD MORE DETAILS TO EVENT LIST
	}
	if len(eventArr) == 0 {
		return []string{"No events"}
	}
	return eventArr
}
func formatSimplePvcList(pvcList []persistentvolumeclaim.PersistentVolumeClaim) string {
	var formatted string = ""
	if len(pvcList) == 0 {
		return "No PVCs associated with this pod"
	}
	for _, pvc := range pvcList {
		formatted += pvc.ObjectMeta.Name + ", "
	}
	formatted = formatted[0 : len(formatted)-2]
	return formatted
}
