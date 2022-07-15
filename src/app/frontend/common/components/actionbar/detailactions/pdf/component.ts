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
import {MatSnackBar, MatSnackBarRef, TextOnlySnackBar} from '@angular/material/snack-bar';
import {LogDetails, LogSources, ObjectMeta, TypeMeta, Container} from '@api/root.api';
import {LogService} from '@common/services/global/logs';
import {jsPDF} from 'jspdf';
import {Observable} from 'rxjs';
import {switchMap, take} from 'rxjs/operators';

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

  static currentSnackbar: MatSnackBarRef<TextOnlySnackBar>;
  static exportPdfComponent: ActionbarDetailExportpdfComponent;

  k8sObjectsWithLogs: string[] = ['pod', 'replicaset', 'statefulset', 'daemonset', 'job'];

  logSources: LogSources;
  pod: string;
  container: string;

  namespace: any;
  resourceType: any;
  resourceName: any;
  containerName: any;
  containers: Container[];
  logsSnapshot: LogDetails;

  private pageWidth: number;

  constructor(private readonly matSnackBar_: MatSnackBar, readonly logService: LogService) {
    this.logService = logService;
    ActionbarDetailExportpdfComponent.exportPdfComponent = this;
  }

  onClick(): void {
    const searchElement = ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind);

    if (this.k8sObjectsWithLogs.includes(this.typeMeta.kind)) {
      this.logService
        .getResource<LogSources>(`source/${this.objectMeta.namespace}/${this.objectMeta.name}/${this.typeMeta.kind}`)
        .pipe(
          switchMap<LogSources, Observable<LogDetails>>(data => {
            this.logSources = data;
            const pod = data.podNames[0]; // Pick first pod (cannot use resource name as it may not be a pod).
            this.containerName = data.containerNames[0]; // Pick from URL or first.

            return this.logService.getResource(`${this.objectMeta.namespace}/${pod}/${this.containerName}`);
          })
        )
        .pipe(take(1))
        .subscribe(data => {
          // finished loading
          this.logsSnapshot = data;
        });
    }

    if (searchElement) {
      ActionbarDetailExportpdfComponent.currentSnackbar = this.matSnackBar_.open('Generating pdf...', 'Dismiss', {
        duration: 10000,
      });
      this.generatePdf(ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind));
    } else {
      console.error('K8s object type not supported yet, aborting');
    }
  }

  private generatePdf(componentName: string) {
    const targetElement: HTMLElement = document.getElementsByTagName(componentName)[0] as HTMLElement;
    this.pageWidth = targetElement.offsetWidth;
    const pdf = new jsPDF({
      orientation: 'landscape',
      unit: 'mm',
      //format: [297, 210], A4 paper
      format: [1500, this.pageWidth],
      userUnit: 300,
    });
    if (targetElement) {
      pdf.html(targetElement, {
        callback: function (pdf) {
          ActionbarDetailExportpdfComponent.exportPdfComponent.pdfExportCallback(pdf);
          ActionbarDetailExportpdfComponent.currentSnackbar.dismiss();
        },
      });
    } else {
      console.error('Error! targetElement was not found! componentName: ' + componentName);
    }
  }

  pdfExportCallback(pdf: jsPDF): void {
    if (this.logsSnapshot) {
      // add logs section
      pdf.addPage();
      pdf.setFontSize(56);
      pdf.text('Logs for container "' + this.containerName + '" on "' + this.objectMeta.name + '"', 20, 30);
      pdf.setFontSize(36);
      let i = 0;
      for (const log of this.logsSnapshot.logs) {
        let pageY = 50 + 20 * i;
        if (pageY > 1500) {
          // need to oveflow to new page
          pdf.addPage();
          i = 0;
          pageY = 50 + 20 * i;
        }
        pdf.setTextColor('#333333');
        pdf.text(log.timestamp + ' --- ', 20, pageY);
        const splitText: string[] = pdf.splitTextToSize(log.content, this.pageWidth);
        pdf.setTextColor('#000000');
        pdf.text(splitText, 245, pageY);
        i += splitText.length;
      }
    }
    pdf.save(
      'Report-' + this.objectMeta.name + '-' + new Date().toISOString().replace(/T/, '_').replace(/:/g, '-') + '.pdf'
    );
  }
}
