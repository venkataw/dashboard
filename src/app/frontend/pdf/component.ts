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

import {Component, ViewChild, OnDestroy, OnInit} from '@angular/core';
import {MatSnackBar} from '@angular/material/snack-bar';
import {MatTable} from '@angular/material/table';
import {ReportService} from './client';
import {PdfRequestStatus, PdfTemplate, ReportContents, ReportItem, ReportsZip} from './types';
import {NamespaceList} from '@api/root.api';

const UPDATE_INTERVAL = 5000; // how often to request a new report list from the api server (in ms)

@Component({
  selector: 'kd-report-list',
  templateUrl: './template.html',
  //styleUrls: ['style.scss'],
})
export class ReportComponent implements OnInit, OnDestroy {
  reportListInitialized = false;
  templateListInitialized = false;
  namespaceListInitialized = false;
  isInitialized = false; // final initialization (ready to render)
  pdfList: ReportItem[] = [];
  templateList: PdfTemplate[] = [];
  updateInterval: NodeJS.Timer;
  namespaceList: NamespaceList;

  // table with report list
  colList = ['name'];
  @ViewChild(MatTable) table: MatTable<ReportItem[]>;

  // dropdown for template selection
  selectedTemplate: string;
  selectedNamespace: string;

  constructor(private readonly reportService_: ReportService, private readonly matSnackBar_: MatSnackBar) {}

  ngOnInit(): void {
    // get report list
    const listObservable = this.reportService_.getList();
    const listObserver = {
      next: (x: ReportItem[]) => (this.pdfList = x),
      error: (err: Error) => console.error('Error getting report list: ' + err),
      complete: () => this.finishInit(true, false, false),
    };
    listObservable.subscribe(listObserver);

    // get template list
    const templateObservable = this.reportService_.getTemplates();
    const templateObserver = {
      next: (x: PdfTemplate[]) => (this.templateList = x),
      error: (err: Error) => console.error('Error getting template list: ' + err),
      complete: () => this.finishInit(false, true, false),
    };
    templateObservable.subscribe(templateObserver);

    // get namespace list
    const namespaceObservable = this.reportService_.getNamespaces();
    const namespaceObserver = {
      next: (x: NamespaceList) => (this.namespaceList = x),
      error: (err: Error) => console.error('Error getting namespace list: ' + err),
      complete: () => this.finishInit(false, false, true),
    };
    namespaceObservable.subscribe(namespaceObserver);

    // Update the table once in a while
    this.updateInterval = setInterval(() => {
      this.updateReportListTable();
    }, UPDATE_INTERVAL);
  }
  ngOnDestroy(): void {
    clearInterval(this.updateInterval);
    console.log('I have been destroyed! kd-report-list');
  }

  finishInit(reportListDone = false, templateListDone = false, namespaceListDone = false) {
    if (reportListDone) {
      this.reportListInitialized = true;
    }
    if (templateListDone) {
      this.templateListInitialized = true;
    }
    if (namespaceListDone) {
      this.namespaceListInitialized = true;
    }
    if (this.reportListInitialized && this.templateListInitialized && this.namespaceListInitialized) {
      // done initializing.
      console.log('kd-report-list initialized!');
      this.isInitialized = true;
    }
  }

  downloadPdf(row: ReportItem): void {
    let pdfContents: Uint16Array;
    const listObservable = this.reportService_.getPdf(row.name);
    const listObserver = {
      next: (x: ReportContents) => {
        // Convert base64 encoded binary string from server
        // into Uint16Array to be downloaded via browser
        const raw: string = window.atob(x.contents);
        let u8arr: Uint8Array;
        if (raw.length % 2 === 1) {
          u8arr = new Uint8Array(raw.length + 1);
        } else {
          u8arr = new Uint8Array(raw.length);
        }
        for (let i = 0; i < raw.length; i++) {
          u8arr[i] = raw.charCodeAt(i);
        }
        pdfContents = new Uint16Array(u8arr.buffer);
      },
      error: (err: Error) => console.error('Error getting pdf content: ' + err),
      complete: () => this.savePdf(pdfContents, row.name),
    };
    listObservable.subscribe(listObserver);
  }

  savePdf(contents: Uint16Array, fileName: string) {
    const blob = new Blob([contents], {
      type: 'application/pdf',
    });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    document.body.appendChild(a);
    a.style.display = 'none';
    a.click();
    a.remove();
    setTimeout(() => {
      return window.URL.revokeObjectURL(url);
    }, 1000);
  }

  saveZip(contents: Uint8Array, fileName: string) {
    const blob = new Blob([contents], {
      type: 'octet/stream',
    });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = fileName;
    document.body.appendChild(a);
    a.style.display = 'none';
    a.click();
    a.remove();
    setTimeout(() => {
      return window.URL.revokeObjectURL(url);
    }, 1000);
  }

  // request new report list from server and update reports table
  updateReportListTable(): void {
    const listObservable = this.reportService_.getList();
    const listObserver = {
      next: (x: ReportItem[]) => (this.pdfList = x),
      error: (err: Error) => console.error('Error getting report list: ' + err),
      complete: () => this.table.renderRows(),
    };
    listObservable.subscribe(listObserver);
  }

  genPdf(): void {
    if (this.selectedTemplate) {
      if (this.selectedNamespace) {
        this.matSnackBar_.open('Generating report...');
        console.log('Requesting generation of template ' + this.selectedTemplate);
        // send request and get status back
        const templateObservable = this.reportService_.genPdf(this.selectedTemplate, this.selectedNamespace);
        const templateObserver = {
          next: (x: PdfRequestStatus) => {
            if (x.status === 'ok') {
              this.updateReportListTable();
              this.matSnackBar_.open('Report generated! ' + x.file, 'Dismiss', {duration: 5000});
            } else {
              console.error('Error generating pdf: ' + x.error);
              this.matSnackBar_.open('Error generating report: ' + x.error, 'Dismiss', {duration: 5000});
            }
          },
          error: (err: Error) => {
            console.error('Error sending request to generate pdf: ' + err);
            this.matSnackBar_.open('Error generating report: ' + err, 'Dismiss', {duration: 5000});
          },
        };
        templateObservable.subscribe(templateObserver);
      } else {
        this.matSnackBar_.open('Select a namespace first!', 'Dismiss', {duration: 5000});
      }
    } else {
      this.matSnackBar_.open('Select a template first!', 'Dismiss', {duration: 5000});
    }
  }

  getZip(): void {
    this.matSnackBar_.open('Generating zip archive of all reports...');
    const zipObservable = this.reportService_.getReportsZip();
    const zipObserver = {
      next: (x: ReportsZip) => {
        if (x.status === 'ok') {
          this.matSnackBar_.open('Zip archive generated!', 'Dismiss', {duration: 5000});
          this.saveZip(
            Uint8Array.from(window.atob(x.contents), c => c.charCodeAt(0)),
            'reports-' + new Date().toISOString().replace(/T/, '_').replace(/:/g, '-') + '.zip'
          );
        } else {
          console.error('Error getting zip archive: ' + x.error);
          this.matSnackBar_.open('Error getting zip archive: ' + x.error, 'Dismiss', {duration: 5000});
        }
      },
      error: (err: Error) => {
        console.error('Error sending request to zip archives: ' + err);
        this.matSnackBar_.open('Error sending request to zip archive: ' + err);
      },
    };
    zipObservable.subscribe(zipObserver);
  }
}
