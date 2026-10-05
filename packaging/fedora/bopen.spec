# RPM spec for building bopen from source, used by COPR through Packit
# (.packit.yaml). Packit sets Version from the release tag; the build needs
# network access to download Go modules (enable_net in .packit.yaml).

Name:           bopen
Version:        0.3.0
Release:        1%{?dist}
Summary:        Link inspector that removes tracking and opens links in the browser you choose

License:        MIT
URL:            https://github.com/blackfyre/bopen
Source0:        %{url}/archive/v%{version}/%{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.26
BuildRequires:  gcc
BuildRequires:  pkgconfig(wayland-client)
BuildRequires:  pkgconfig(wayland-cursor)
BuildRequires:  pkgconfig(wayland-egl)
BuildRequires:  pkgconfig(egl)
BuildRequires:  pkgconfig(x11)
BuildRequires:  pkgconfig(x11-xcb)
BuildRequires:  pkgconfig(xkbcommon)
BuildRequires:  pkgconfig(xkbcommon-x11)
BuildRequires:  pkgconfig(xcursor)
BuildRequires:  pkgconfig(xfixes)
BuildRequires:  vulkan-headers

# Go binaries carry no debug sources in the expected layout.
%global debug_package %{nil}

%description
bopen registers as the default web browser. When a link is clicked it shows
the link with redirect wrappers and tracking parameters highlighted, each
with the reason it is suggested for removal, and opens the result in the
browser, profile or private window you choose.

%prep
%autosetup

%build
export CGO_ENABLED=1 GOFLAGS="-trimpath -mod=readonly"
go build -ldflags "-s -w -X main.version=%{version}" -o bopen ./cmd/bopen

%check
go test ./internal/...

%install
install -Dpm 0755 bopen %{buildroot}%{_bindir}/bopen

%files
%license LICENSE
%doc README.md
%{_bindir}/bopen

%changelog
* Mon Oct 05 2026 bopen maintainers <https://github.com/blackfyre/bopen> - 0.3.0-1
- Release notes: https://github.com/blackfyre/bopen/releases
