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
import {Observable} from 'rxjs';

@Injectable({
  providedIn: 'root',
})
export class ReportService {
  http_: HttpClient;

  constructor(readonly http: HttpClient) {
    this.http_ = http;
  }

  getList(): Observable<Object> {
    return this.http_.get('pdf');
  }

  getPdf(name: string): Observable<Object> {
    return this.http_.get('pdf/pdf/' + name);
  }

  getTemplates(): Observable<Object> {
    return this.http_.get('pdf/templates');
  }

  genPdf(templateName: string, namespace: string): Observable<Object> {
    return this.http_.get('pdf/gen/' + templateName + '/' + namespace);
  }

  getNamespaces(): Observable<Object> {
    return this.http_.get('api/v1/namespace');
  }

  getReportsZip(): Observable<Object> {
    return this.http_.get('pdf/zip');
  }

  requestDeleteReport(name: string) {
    return this.http_.get('pdf/delete/' + name);
  }
}
