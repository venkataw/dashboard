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
import {ReportItem} from './reporttypes';

const UPDATE_INTERVAL = 10000; // how often to request a new report list from the api server

@Component({
  selector: 'kd-report-list',
  templateUrl: './template.html',
  //styleUrls: ['style.scss'],
})
export class ReportComponent implements OnInit, OnDestroy {
  isInitialized = false;
  pdfList: ReportItem[] = [];
  colList = ['name'];

  updateInterval: NodeJS.Timer;

  @ViewChild(MatTable) table: MatTable<ReportItem[]>;

  constructor(private readonly reportService_: ReportService) {}

  ngOnInit(): void {
    const listObservable = this.reportService_.getList();
    const listObserver = {
      next: (x: ReportItem[]) => (this.pdfList = x),
      error: (err: Error) => console.error('Error getting report list: ' + err),
      complete: () => {
        console.log('Done getting list');
        this.isInitialized = true;
        console.log('I have been initialized! kd-report-list');
        console.log(this.pdfList);
      },
    };
    listObservable.subscribe(listObserver);

    this.updateInterval = setInterval(() => {
      this.updateTable();
    }, UPDATE_INTERVAL);
  }
  ngOnDestroy(): void {
    clearInterval(this.updateInterval);
    console.log('I have been destroyed! kd-report-list');
  }

  downloadPdf(row: any): void {
    console.log(row);
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
