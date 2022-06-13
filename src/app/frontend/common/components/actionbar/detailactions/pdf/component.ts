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

import {Component, Input} from '@angular/core';
import {ObjectMeta, TypeMeta} from '@api/root.api';

import {VerberService} from '@common/services/global/verber';
import {jsPDF} from 'jspdf';

@Component({
  selector: 'kd-actionbar-detail-exportpdf',
  templateUrl: './template.html',
})
export class ActionbarDetailExportpdfComponent {
  @Input() objectMeta: ObjectMeta;
  @Input() typeMeta: TypeMeta;
  @Input() displayName: string;

  static k8sObjectMap: Map<string, string> = new Map<string, string>([
    ['pod', 'kd-pod-detail'],
    ['workloads', 'kd-workload-statuses'],
    ['cronjob', 'kd-cron-job-detail'],
    ['daemonset', 'kd-daemon-set-detail'],
    ['deployment', 'kd-deployment-detail'],
    ['job', 'kd-job-detail'],
    ['replicaset', 'kd-replica-set-detail'],
    ['replication-controller', 'kd-replication-controller-detail'],
    ['statefulset', 'kd-stateful-set-detail'],
    ['clusterrole', 'kd-cluster-role-detail'],
    ['clusterrolebinding', 'kd-cluster-role-binding-detail'],
    ['namespace', 'kd-namespace-detail'],
    ['networkpolicy', 'kd-network-policy-detail'],
    ['node', 'kd-node-detail'],
    ['persistentvolume', 'kd-persistent-volume-detail'],
    ['role', 'kd-role-detail'],
    ['rolebinding', 'kd-role-detail'],
    ['serviceaccount', 'kd-service-account-detail'],
    ['ingress', 'kd-ingress-detail'],
    ['ingressclass', 'kd-ingress-class-detail'],
    ['service', 'kd-service-detail'],
    ['configmap', 'kd-config-map-detail'],
    ['persistentvolumeclaim', 'kd-persistent-volume-claim-detail'],
    ['secret', 'kd-secret-detail'],
    ['storageclass', 'kd-storage-class-detail'],
  ]);

  constructor(private readonly verber_: VerberService) {}

  onClick(): void {
    const searchElement = ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind);
    if (searchElement !== null && searchElement !== undefined) {
      console.log('Generating pdf, typeMeta.kind is ' + this.typeMeta.kind + ', mapped is ' + searchElement);
      this.generatePdf(ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind));
    } else {
      console.error('K8s object type not supported yet, aborting');
    }
  }

  private generatePdf(componentName: string) {
    const targetElement: HTMLElement = document.getElementsByTagName(componentName)[0] as HTMLElement;
    const pdf = new jsPDF({
      orientation: 'landscape',
      unit: 'mm',
      //format: [297, 210], A4 paper
      format: [1500, targetElement.offsetWidth],
      userUnit: 300,
    });
    if (targetElement !== null && targetElement !== undefined) {
      pdf.html(targetElement, {
        callback: function (pdf) {
          pdf.save('Report-' + new Date().toISOString().replace(/T/, '_').replace(/:/g, '-') + '.pdf');
        },
      });
    } else {
      console.error('Error! targetElement was not found! componentName: ' + componentName);
    }
  }
}
