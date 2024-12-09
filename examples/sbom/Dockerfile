# syntax=docker/dockerfile:1.5

# Copyright 2022 buildkit-syft-scanner authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM scratch

COPY <<EOF /var/share/sbom/foo.spdx.json
{
 "spdxVersion": "SPDX-2.3",
 "dataLicense": "CC0-1.0",
 "SPDXID": "SPDXRef-DOCUMENT",
 "packages": [
  {
   "name": "foo",
   "SPDXID": "SPDXRef-Package-foo-0d50d654eb648ebd",
   "versionInfo": "1.0",
   "supplier": "NOASSERTION",
   "downloadLocation": "NOASSERTION",
   "filesAnalyzed": true,
   "packageVerificationCode": {
    "packageVerificationCodeValue": "ee259e59ebc5bf49005492c1a393d32158491196"
   },
   "licenseConcluded": "NOASSERTION",
   "licenseDeclared": "GPL-2.0-only AND GPL-2.0-or-later",
   "copyrightText": "NOASSERTION",
   "externalRefs": [
    {
     "referenceCategory": "SECURITY",
     "referenceType": "cpe23Type",
     "referenceLocator": "cpe:2.3:a:deb:deb:1.0:*:*:*:*:*:*:*"
    },
    {
     "referenceCategory": "PACKAGE-MANAGER",
     "referenceType": "purl",
     "referenceLocator": "pkg:generic/foo@1.0"
    }
   ]
  }
 ]
}
EOF