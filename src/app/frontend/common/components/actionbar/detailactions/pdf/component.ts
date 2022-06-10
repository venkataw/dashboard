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

  constructor(private readonly verber_: VerberService) {}

  onClick(): void {
    console.log('PDF generating!');

    switch (this.typeMeta.kind) {
      case 'pod':
        this.generatePodPdf();
        break;
      default:
        console.log('K8s object type not supported. Aborting');
        break;
    }
  }

  private generatePodPdf() {
    const pdf = new jsPDF({
      orientation: 'landscape',
      format: 'legal',
      userUnit: 300,
    });
    pdf.html(document.getElementsByTagName('kd-pod-detail')[0] as HTMLElement, {
      callback: function (pdf) {
        pdf.save('Report-' + new Date().toISOString().replace(/T/, '_').replace(/:/g, '-') + '.pdf');
      },
    });
  }
}
