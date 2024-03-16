Name:           sharify-desktop
Version:        1.0.0
Release:        1%{?dist}
Summary:        Sharify Desktop for Linux
License:        MPL-2.0
URL:            https://xericl.dev/
Source0:        sharify-desktop-amd64
Source1:        %{name}.desktop
Source2:        %{name}-icon.png
Source3:        LICENSE

%description
Sharify Desktop for Linux

%install
install -D -m 755 %{SOURCE0} %{buildroot}%{_bindir}/%{name}
install -D -m 644 %{SOURCE1} %{buildroot}%{_datadir}/applications/%{name}.desktop
install -D -m 644 %{SOURCE2} %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/%{name}-icon.png
install -D -m 644 %{SOURCE3} %{buildroot}%{_datadir}/doc/%{name}/LICENSE

%post
update-desktop-database %{_datadir}/applications

%postun
update-desktop-database %{_datadir}/applications

%files
%{_bindir}/%{name}
%{_datadir}/applications/%{name}.desktop
%{_datadir}/icons/hicolor/256x256/apps/%{name}-icon.png
%doc %{_datadir}/doc/%{name}/LICENSE