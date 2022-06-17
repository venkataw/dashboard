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
import {ExportPdfComponent} from './pdftracker';

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
    ExportPdfComponent.exportPdfComponent = this;

    if (ExportPdfComponent.curResourceType === 'pod') {
      this.namespace = ExportPdfComponent.curNamespace;
      this.resourceType = ExportPdfComponent.curResourceType;
      this.resourceName = ExportPdfComponent.curResourceName;
      this.containers = ExportPdfComponent.curContainers;
      this.containerName = this.containers[0].name; //TODO: FIX THIS (support for all containers in different pages)
    } else {
      console.log('No logs available for type ' + ExportPdfComponent.curResourceType + ', skipping logs collection');
      this.namespace = undefined;
      this.resourceType = undefined;
      this.resourceName = undefined;
      this.containerName = undefined;
    }

    this.logService = logService;

    console.log('Got details in constructor');
    console.log(this.namespace);
    console.log(this.resourceType);
    console.log(this.resourceName);
    console.log(this.containerName);
  }

  onClick(): void {
    const searchElement = ActionbarDetailExportpdfComponent.k8sObjectMap.get(this.typeMeta.kind);
    const dryRun = false;

    console.log('Got details in onclick()');
    console.log(this.namespace);
    console.log(this.resourceType);
    console.log(this.resourceName);
    console.log(this.containerName);

    if (this.namespace && this.resourceType && this.resourceName && this.containerName) {
      this.logService
        .getResource<LogSources>(`source/${this.namespace}/${this.resourceName}/${this.resourceType}`)
        .pipe(
          switchMap<LogSources, Observable<LogDetails>>(data => {
            this.logSources = data;
            this.pod = data.podNames[0]; // Pick first pod (cannot use resource name as it may not be a pod).
            this.container = this.containerName ? this.containerName : data.containerNames[0]; // Pick from URL or first.

            return this.logService.getResource(`${this.namespace}/${this.pod}/${this.container}`);
          })
        )
        .pipe(take(1))
        .subscribe(data => {
          // finished loading
          console.log('Finished loading. logdetail is:');
          console.log(data);
          this.logsSnapshot = data;
        });
    }

    if (dryRun) return;

    if (searchElement !== null && searchElement !== undefined) {
      console.log('Generating pdf, typeMeta.kind is ' + this.typeMeta.kind + ', mapped is ' + searchElement);
      ActionbarDetailExportpdfComponent.currentSnackbar = this.matSnackBar_.open('Generating pdf...', 'Dismiss');
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
    if (targetElement !== null && targetElement !== undefined) {
      pdf.html(targetElement, {
        callback: function (pdf) {
          ExportPdfComponent.exportPdfComponent.dismissCurrentSnackbar();
          ExportPdfComponent.exportPdfComponent.pdfExportCallback(pdf);
        },
      });
    } else {
      console.error('Error! targetElement was not found! componentName: ' + componentName);
    }
  }

  dismissCurrentSnackbar(): void {
    const curSnackbar: MatSnackBarRef<TextOnlySnackBar> = ActionbarDetailExportpdfComponent.currentSnackbar;
    if (curSnackbar !== null || curSnackbar !== undefined) {
      curSnackbar.dismiss();
    }
  }

  pdfExportCallback(pdf: jsPDF): void {
    if (this.logsSnapshot !== null && this.logsSnapshot !== undefined) {
      // add logs section
      pdf.addPage();
      pdf.setFontSize(56);
      pdf.text('Logs for container "' + this.containerName + '" on "' + this.resourceName + '"', 20, 30);
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
        const splitText: string[] = pdf.splitTextToSize(log.timestamp + ' --- ' + log.content, this.pageWidth);
        pdf.text(splitText, 20, pageY); // TODO: clean this up (different colors)
        i += splitText.length;
      }
    } else {
      console.log('logsSnapshot is null, skipping logs section');
    }
    pdf.save(
      'Report-' + this.objectMeta.name + '-' + new Date().toISOString().replace(/T/, '_').replace(/:/g, '-') + '.pdf'
    );
  }
}
