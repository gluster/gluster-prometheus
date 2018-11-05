%if 0%{?fedora}
%global with_bundled 1
%else
%global with_bundled 1
%endif

%{!?with_debug: %global with_debug 1}

%if 0%{?with_debug}
%global _dwz_low_mem_die_limit 0
%else
%global debug_package   %{nil}
%endif

%{!?go_arches: %global go_arches x86_64 aarch64 ppc64le }

%global provider github
%global provider_tld com
%global project gluster
%global repo gluster-prometheus
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path %{provider_prefix}

%global gluster_prom_make %{__make} PREFIX=%{_prefix} EXEC_PREFIX=%{_exec_prefix} BINDIR=%{_bindir} SBINDIR=%{_sbindir} DATADIR=%{_datadir} LOCALSTATEDIR=%{_sharedstatedir} LOGDIR=%{_localstatedir}/log SYSCONFDIR=%{_sysconfdir} FASTBUILD=off

%global gluster_prom_ver 1
%global gluster_prom_rel 0

Name: %{repo}
Version: %{gluster_prom_ver}
Release: 0%{?dist}
Summary: The GlusterFS prometheus metrics collectors
License: GPLv2 or LGPLv3+
URL: https://%{provider_prefix}
%if 0%{?with_bundled}
Source0: https://%{provider_prefix}/releases/download/v%{version}/gluster-prometheus-exporter-v%{gluster_prom_ver}-%{gluster_prom_rel}-vendor.tar.xz
%else
Source0: https://%{provider_prefix}/releases/download/v%{version}/gluster-prometheus-exporter-v%{gluster_prom_ver}-%{gluster_prom_rel}.tar.xz
%endif

ExclusiveArch: %{go_arches}

BuildRequires: %{?go_compiler:compiler(go-compiler)}%{!?go_compiler:golang}
BuildRequires: systemd

%if ! 0%{?with_bundled}
BuildRequires: golang(github.com/BurntSushi/toml)
BuildRequires: golang(github.com/sirupsen/logrus)
BuildRequires: golang(github-prometheus-client_golang)
%endif

Requires: glusterfs-server >= 3.12.0
Requires: /usr/bin/strings
%{?systemd_requires}

%description
The project gluster-prometheus provides set of metrics collectors which
would be run on gluster storage nodes. These would generate metrics for
consumption by prometheus server.

%prep
%setup -q -n gluster-prometheus

%build
export GOPATH=$(pwd):%{gopath}
mkdir -p src/%(dirname %{import_path})
ln -s ../../../ src/%{import_path}

pushd src/%{import_path}
# Build gluster-prometheus
%{gluster_prom_make} gluster-exporter
popd

%install
# Install gluster-prometheus
%{gluster_prom_make} DESTDIR=%{buildroot} install

%post
%systemd_post gluster-exporter.service

%preun
%systemd_preun gluster-exporter.service

%files
%{_sbindir}/gluster-exporter
%{_unitdir}/gluster-exporter.service
%{_sysconfdir}/gluster-exporter/gluster-exporter.toml

%changelog
* Sat Nov 3 2018 Aravinda VK <avishwan@redhat.com> - 1.0.0-1
- Fixed version numbers and changed name from gluster-exporter to
  gluster-prometheus-exporter

* Thu Sep 27 2018 Shubhendu Ram Tripathi <shtripat@redhat.com> - 1.0.0-1
- Initial spec
