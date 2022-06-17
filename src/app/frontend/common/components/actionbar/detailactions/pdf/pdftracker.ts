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

import {Container} from '@api/root.api';
import {ActionbarDetailExportpdfComponent} from './component';

export class ExportPdfComponent {
  static exportPdfComponent: ActionbarDetailExportpdfComponent;

  static curNamespace: string;
  static curResourceType: string;
  static curResourceName: string;
  static curContainers: Container[];

  static clearCurResource(): void {
    ExportPdfComponent.curNamespace = undefined;
    ExportPdfComponent.curResourceType = undefined;
    ExportPdfComponent.curResourceName = undefined;
    ExportPdfComponent.curContainers = undefined;
  }

  static setCurResource(namespace: string, type: string, name: string, containers: Container[]) {
    ExportPdfComponent.curNamespace = namespace;
    ExportPdfComponent.curResourceType = type;
    ExportPdfComponent.curResourceName = name;
    ExportPdfComponent.curContainers = containers;
  }
}
