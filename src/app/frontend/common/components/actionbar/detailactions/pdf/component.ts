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
    ['statefulset', 'kd-stateful-set-detail']
  ]);

  constructor(private readonly verber_: VerberService) {}

  onClick(): void {
    const searchElement = ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind);
    if (searchElement != null) {
      console.log('Generating pdf, typeMeta.kind is ' + this.typeMeta.kind + ", mapped is " + searchElement);
      this.generatePdf(ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind));
    } else {
      console.error("K8s object type not supported yet, aborting");
    }
  }

  private generatePdf(componentName: string) {
    const pdf = new jsPDF({
      orientation: 'landscape',
      unit: 'mm',
      //format: [297, 210], A4 paper
      format: [1500, 1100],
      userUnit: 300,
    });
    let targetElement: HTMLElement = document.getElementsByTagName(componentName)[0] as HTMLElement;
    if (targetElement != null) {
      pdf.html(targetElement, {
        callback: function (pdf) {
          pdf.save('Report-' + new Date().toISOString().replace(/T/, '_').replace(/:/g, '-') + '.pdf');
        },
      });
    } else {
      console.error("Error! targetElement was not found! componentName: " + componentName);
    }
  }
}
