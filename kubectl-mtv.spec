%global provider        github
%global provider_tld    com
%global project         yaacov
%global repo            kubectl-mtv
%global provider_prefix %{provider}.%{provider_tld}/%{project}/%{repo}
%global import_path     %{provider_prefix}

%undefine _missing_build_ids_terminate_build
%define debug_package %{nil}

Name:           %{repo}
Version:        0.1.5
Release:        1%{?dist}
Summary:        kubectl-mtv - Migration Toolkit for Virtualization CLI
License:        Apache
URL:            https://%{import_path}
Source0:        https://github.com/yaacov/kubectl-mtv/archive/v%{version}.tar.gz

BuildRequires:  git
BuildRequires:  golang >= 1.23.0

%description
The Migration Toolkit for Virtualization (MTV) simplifies the process of migrating virtual machines from traditional 
virtualization platforms (oVirt, VMware, OpenStack, and OVA files) to Kubernetes using KubeVirt. It handles the 
complexities of different virtualization platforms and provides a consistent way to define, plan, and execute migrations.

%prep
%setup -q -n kubectl-mtv-%{version}

%build
# set up temporary build gopath, and put our directory there
mkdir -p ./_build/src/github.com/yaacov
ln -s $(pwd) ./_build/src/github.com/yaacov/kubectl-mtv

VERSION=v%{version} make

%install
install -d %{buildroot}%{_bindir}
install -p -m 0755 ./kubectl-mtv %{buildroot}%{_bindir}/kubectl-mtv

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
%{_bindir}/kubectl-mtv

%changelog

* Mon Apr 7 2025 Initial package - 0.1.0-1
- First release of kubectl-mtv
- Migration Toolkit for Virtualization to migrate VMs to KubeVirt
