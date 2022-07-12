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

import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';

import {ReportItem} from './reporttypes';

@Injectable({
  providedIn: 'root',
})
export class ReportService {
  http_: HttpClient;

  constructor(readonly http: HttpClient) {
    this.http_ = http;
  }

  getList(): string[] {
    const pdfList: string[] = [];
    const listObservable = this.http_.request('GET', 'pdf');
    const listObserver = {
      next: (x: ReportItem[]) => {
        for (const item of x) {
          pdfList.push(item.name);
        }
        return pdfList;
      },
      error: (err: Error) => console.error('GETting report list error: ' + err),
      //complete: () => console.log('Observer got a complete notification'),
    };
    listObservable.subscribe(listObserver);

    return pdfList;
  }
}
