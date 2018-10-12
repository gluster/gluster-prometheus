# Copyright 2018 The Gluster Prometheus Authors.

# Licensed under GNU LESSER GENERAL PUBLIC LICENSE Version 3, 29 June 2007
# You may obtain a copy of the License at
#    https://opensource.org/licenses/lgpl-3.0.html

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#-- Build phase
FROM openshift/origin-release:golang-1.10 AS build

ENV GOPATH="/go/" \
    SRCDIR="/go/src/github.com/gluster/gluster-prometheus/"

RUN yum install -y \
    git

# Install dep
RUN mkdir -p /go/bin
RUN curl -L https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Vendor dependencies
COPY Gopkg.lock Gopkg.toml "${SRCDIR}"
WORKDIR "${SRCDIR}/gluster-exporter"
RUN /go/bin/dep ensure -v -vendor-only

# The version of the driver (git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2')
ARG version="(unknown)"

# Build executable
COPY . "${SRCDIR}"
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static" -X main.ExporterVersion=${version} -X main.defaultGlusterd1Workdir=/var/lib/glusterd -X main.defaultGlusterd2Workdir=/var/lib/glusterd2' -o /gluster-exporter .

# Ensure the binary is statically linked
RUN ldd /gluster-exporter | grep -q "not a dynamic executable"

RUN SBINDIR=/usr/sbin SYSCONFDIR=/etc ../scripts/gen-service.sh /

#-- Final container
FROM gluster/glusterd2-nightly

# Extra packages which are not available in Glusterd2 nightly image

RUN curl -o /etc/yum.repos.d/glusterfs-nightly.repo http://artifacts.ci.centos.org/gluster/nightly/master.repo

# Install dependencies
RUN yum --setopt=tsflags=nodocs -y install glusterfs

# Copy gluster-exporter from build phase
COPY --from=build /gluster-exporter.service /usr/lib/systemd/system/gluster-exporter.service
COPY --from=build /gluster-exporter /usr/sbin/gluster-exporter
COPY ./extras/conf/global.conf.sample /etc/gluster-exporter/global.conf
COPY ./extras/conf/collectors.conf.sample /etc/gluster-exporter/collectors.conf

# Make glusterd default
RUN systemctl enable gluster-exporter.service && \
        systemctl enable glusterd.service && \
        systemctl disable glusterd2.service

# The version of the driver (git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2')
ARG version="(unknown)"
# Container build time (date -u '+%Y-%m-%dT%H:%M:%S.%NZ')
ARG builddate="(unknown)"

LABEL build-date="${builddate}"
LABEL io.k8s.description="Gluster Prometheus exporter"
LABEL name="gluster/gluster-prometheus"
LABEL Summary="Gluster Prometheus exporter"
LABEL vcs-type="git"
LABEL vcs-url="https://github.com/gluster/gluster-prometheus"
LABEL vendor="gluster.org"
LABEL version="${version}"

CMD ["/usr/sbin/init"]
