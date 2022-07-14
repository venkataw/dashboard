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
import {MatTable} from '@angular/material/table';
import {ReportService} from './client';
import {PdfTemplate, ReportContents, ReportItem} from './reporttypes';

const UPDATE_INTERVAL = 5000; // how often to request a new report list from the api server (in ms)

@Component({
  selector: 'kd-report-list',
  templateUrl: './template.html',
  //styleUrls: ['style.scss'],
})
export class ReportComponent implements OnInit, OnDestroy {
  reportListInitialized = false;
  templateListInitialized = false;
  isInitialized = false; // final initialization (ready to render)
  pdfList: ReportItem[] = [];
  templateList: PdfTemplate[] = [];
  updateInterval: NodeJS.Timer;

  // table with report list
  colList = ['name'];
  @ViewChild(MatTable) table: MatTable<ReportItem[]>;

  constructor(private readonly reportService_: ReportService) {}

  ngOnInit(): void {
    // get report list
    const listObservable = this.reportService_.getList();
    const listObserver = {
      next: (x: ReportItem[]) => (this.pdfList = x),
      error: (err: Error) => console.error('Error getting report list: ' + err),
      complete: () => {
        this.finishInit(true, false);
        console.log('Report list initialized!');
        console.log(this.pdfList);
      },
    };
    listObservable.subscribe(listObserver);

    // get template list
    const templateObservable = this.reportService_.getTemplates();
    const templateObserver = {
      next: (x: PdfTemplate[]) => (this.templateList = x),
      error: (err: Error) => console.error('Error getting template list: ' + err),
      complete: () => {
        this.finishInit(false, true);
        console.log('Template list initialized!');
        console.log(this.templateList);
      },
    };
    templateObservable.subscribe(templateObserver);

    // Update the table once in a while
    this.updateInterval = setInterval(() => {
      this.updateTable();
    }, UPDATE_INTERVAL);
  }
  ngOnDestroy(): void {
    clearInterval(this.updateInterval);
    console.log('I have been destroyed! kd-report-list');
  }

  finishInit(reportListDone = false, templateListDone = false) {
    if (reportListDone) {
      this.reportListInitialized = true;
    }
    if (templateListDone) {
      this.templateListInitialized = true;
    }
    if (this.reportListInitialized && this.templateListInitialized) {
      // done initializing.
      console.log('kd-report-list initialized!');
      this.isInitialized = true;
    }
  }

  downloadPdf(row: ReportItem): void {
    console.log('Downloading pdf ' + row.name);
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
      complete: () => {
        console.log('saving pdf...');
        this.savePdf(pdfContents, row.name);
      },
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

  updateTable(): void {
    const listObservable = this.reportService_.getList();
    const listObserver = {
      next: (x: ReportItem[]) => (this.pdfList = x),
      error: (err: Error) => console.error('Error getting report list: ' + err),
      complete: () => {
        console.log('Table updated.');
        console.log(this.pdfList);
        this.table.renderRows();
      },
    };
    listObservable.subscribe(listObserver);
  }
}
