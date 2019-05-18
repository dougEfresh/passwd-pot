%define ver 1.0
%define rel 1%{?dist}

Summary: Passwd pot 
Name: passwd-pot
Version: %{ver}
Release: %{rel}
URL: https://passwd-pot.io
License: BSD
Group: Applications/Internet
BuildRoot: %{_tmppath}/%{name}-%{version}-buildroot
Group: System Environment/Daemons

%description

%prep

%build

%install
rm -rf $RPM_BUILD_ROOT
install -d -m755 $RPM_BUILD_ROOT/%{_bindir}
install -d -m755 $RPM_BUILD_ROOT/%{_unitdir}
install -m755  /root/build/passwd-pot  $RPM_BUILD_ROOT/%{_bindir}/passwd-pot
install -m644 /root/build/contrib/awslinux/passwd-pot.service $RPM_BUILD_ROOT/%{_unitdir}/passwd-pot.service
install -m644 /root/build/contrib/awslinux/passwd-pot-proxy.service $RPM_BUILD_ROOT/%{_unitdir}/passwd-pot-proxy.service

%clean
rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
%attr(0755,root,root) %{_bindir}/passwd-pot
%attr(0644,root,root) %{_unitdir}/passwd-pot.service
%attr(0644,root,root) %{_unitdir}/passwd-pot-proxy.service

%post
%systemd_post passwd-pot.service
%systemd_post passwd-pot-proxy.service

%preun
%systemd_preun passwd-pot-proxy.service

%postun
%systemd_postun_with_restart passwd-pot-proxy.service


%changelog
