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

import {Component, OnDestroy, OnInit} from '@angular/core';

@Component({
  selector: 'kd-report-list',
  templateUrl: './template.html',
  //styleUrls: ['style.scss'],
})
export class ReportComponent implements OnInit, OnDestroy {
  isInitialized = false;

  ngOnInit(): void {
    setTimeout(() => {
      this.isInitialized = true;
      console.log('I have been initialized! kd-report-list');
    }, 1000);
  }
  ngOnDestroy(): void {
    console.log('I have been destroyed! kd-report-list');
  }
}
