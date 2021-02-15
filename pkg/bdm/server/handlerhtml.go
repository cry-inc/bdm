package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func createHTMLHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(html)
	}
}

var html = []byte(`
<!DOCTYPE html>
<html lang="en">
	<head>
		<title>BDM</title>
		<meta charset="utf-8"/>
		<style>
			body {font-family: sans-serif; margin: 10px}
			table {width: 100%}
			th, td {padding: 5px; border: 1px solid #eee}
			th {text-align: left; background-color: #fafafa}
			a {color: #007bff; text-decoration: none}
			a:hover {text-decoration: underline}
			#bc {background-color: #fafafa; padding: 10px; border-radius: 4px; margin-bottom: 10px}
			#error {background-color: #fee; padding: 10px; border-radius: 4px}
			.size {text-align: right; padding-left: 10px; padding-right: 10px}
			.tt {cursor: help}
		</style>
	</head>
	<body>
		<div id="bc"></div>

		<div id="packages">
			<h1>Packages</h1>
			<ul id="packagesList"></ul>
		</div>

		<div id="versions">
			<h1><span id="versionsTitle">X</span> Versions</h1>
			<ul id="versionsList"></ul>
		</div>

		<div id="package">
			<h1><span id="packageName">X</span> Version <span id="packageVersion">Y</span></h1>
			<table id="packageInfo">
				<tr><th>Files</th><td id="packageFiles"></td></tr>
				<tr><th>Size</th><td id="packageSize"></td></tr>
				<tr><th>Published</th><td id="packagePublished"></td></tr>
				<tr><th>Hash</th><td id="packageHash"></td></tr>
			</table>
			<p>
				<a id="packageDownload" href="">Download Package as ZIP</a><br>
				<a id="packageManifest" href="">Package Manifest JSON</a><br>
				<a id="packageCompare" href="">Compare with Previous Version</a><br>
			</p>
			<table>
				<tr>
					<th>File</th>
					<th class="size">Size</th>
					<th>Hash</th>
				</tr>
				<tbody id="packageTable"></tbody>
			</table>
		</div>

		<div id="compare">
			<h1>
				Compare Package <span id="comparePackage">X</span>
				Version <span id="compareA">A</span>
				and <span id="compareB">B</span>
			</h1>
			<table>
				<tr>
					<th>File</th>
					<th>Status</th>
					<th class="size">Old Size</th>
					<th class="size">New Size</th>
					<th>Old Hash</th>
					<th>New Hash</th>
				</tr>
				<tbody id="compareTable"></tbody>
			</table>
		</div>

		<div id="error"></div>

		<script>
			function hideAll() {
				['packages', 'versions', 'package', 'compare', 'error'].forEach(id => {
					document.getElementById(id).style.display = 'none';
				});
			}

			function sizeToString(bytes) {
				if (bytes < 1024) {
					return bytes + " byte";
				} else if (bytes < 1000 * 1000) {
					return Math.round(bytes / 1000) + " kB";
				} else if (bytes < 1000 * 1000 * 1000) {
					return Math.round(bytes / (1000 * 1000)) + " MB";
				} else {
					return Math.round(bytes / (1000 * 1000 * 1000)) + " GB";
				}
			}

			async function loadPackages() {
				const response = await fetch('manifests');
				const packages = await response.json();
				hideAll();
				renderPackages(packages);
			}

			function renderBreadcrumbs(crumbs) {
				let html = 'BDM';
				crumbs.forEach(crumb => {
					html += ' / ';
					if (crumb.length === 2) {
						html += '<a href="' + crumb[1] + '">' + crumb[0] + '</a>';
					} else {
						html += crumb[0];
					}
				});
				document.getElementById('bc').innerHTML = html;
			}

			function renderError(html) {
				document.getElementById('error').innerHTML = html;
				document.getElementById('error').style.display = 'block';
			}

			function renderPackages(packages) {
				renderBreadcrumbs([['Packages']]);
				if (packages.length > 0) {
					let html = '';
					packages.forEach(package => {
						html += '<li><a href="#/' + package.Name + '">' + package.Name + '</a></li>';
					});
					document.getElementById('packagesList').innerHTML = html;
					document.getElementById('packages').style.display = 'block';
				} else {
					renderError('This server does not have any packages.');
				}
			}

			async function loadVersions(package) {
				const response = await fetch('manifests/' + package);
				const versions = response.ok ? await response.json() : [];
				hideAll();
				renderVersions(package, versions);
			}

			function renderVersions(package, versions) {
				renderBreadcrumbs([['Packages', '#/'], [package]]);
				document.getElementById('versionsTitle').innerHTML = package;
				if (versions.length > 0) {
					let html = '';
					versions.forEach(version => {
						html += '<li><a href="#/' + package + '/' + version.Version + '">Version ' + version.Version + '</a></li>';
					});
					document.getElementById('versionsList').innerHTML = html;
					document.getElementById('versions').style.display = 'block';
				} else {
					renderError('Unable to find package.');
				}
			}

			async function loadPackage(package, version) {
				const response = await fetch('manifests/' + package + '/' + version);
				const manifest = response.ok ? await response.json() : null;
				hideAll();
				renderPackage(package, version, manifest);
			}

			function renderPackage(package, version, manifest) {
				renderBreadcrumbs([
					['Packages', '#/'],
					[package, '#/' + package],
					[version]
				]);
				if (manifest !== null) {
					document.getElementById('packageName').innerHTML = package;
					document.getElementById('packageVersion').innerHTML = version;
					let html = '';
					let size = 0;
					manifest.Files.forEach(file => {
						let fileName = file.Path;
						const lastSlash = fileName.lastIndexOf('/');
						if (lastSlash !== -1) {
							fileName = fileName.substr(lastSlash + 1);
						}
						const fileLink = '/files/' + package + '/' + version + '/' + file.Object.Hash + '/' + fileName;
						html += '<tr><td><a href="' + fileLink + '">' + file.Path + '</a></td><td class="size tt" title="' + file.Object.Size + ' byte">' + sizeToString(file.Object.Size) + '</td><td>' + file.Object.Hash + '</td></tr>';
						size += file.Object.Size;
					});
					document.getElementById('packageDownload').href = 'zip/' + package + '/' + version;
					document.getElementById('packageManifest').href = 'manifests/' + package + '/' + version;
					document.getElementById('packageTable').innerHTML = html;
					document.getElementById('packageFiles').innerHTML = manifest.Files.length.toString();
					document.getElementById('packageSize').innerHTML = sizeToString(size);
					document.getElementById('packagePublished').innerHTML = new Date(manifest.Published * 1000).toLocaleString();
					document.getElementById('packageCompare').href = '#/compare/' + package + '/' + version + '/' + (version - 1);
					document.getElementById('packageCompare').style.display = version > 1 ? 'initial' : 'none';
					document.getElementById('packageHash').innerHTML = manifest.Hash;
					document.getElementById('package').style.display = 'block';
				} else {
					renderError('Unable to find package version.');
				}
			}

			async function loadCompare(package, versionA, versionB) {
				const responseA = await fetch('manifests/' + package + '/' + versionA);
				const manifestA = responseA.ok ? await responseA.json() : null;
				const responseB = await fetch('manifests/' + package + '/' + versionB);
				const manifestB = responseB.ok ? await responseB.json() : null;
				hideAll();
				renderCompare(package, versionA, versionB, manifestA, manifestB);
			}

			function renderCompare(package, versionA, versionB, manifestA, manifestB) {
				renderBreadcrumbs([
					['Packages', '#/'],
					[package, '#/' + package],
					[versionA, '#/' + package + '/' + versionA],
					['Compare']
				]);
				if (manifestA !== null && manifestB !== null) {
					document.getElementById('comparePackage').innerHTML = package;
					document.getElementById('compareA').innerHTML = versionA;
					document.getElementById('compareB').innerHTML = versionB;
					const addedFiles = [];
					manifestA.Files.forEach(fileA => {
						const found = manifestB.Files.filter(fileB => fileB.Path === fileA.Path).length > 0;
						if (!found) {
							addedFiles.push(fileA);
						}
					});
					const deletedFiles = [];
					manifestB.Files.forEach(fileB => {
						const found = manifestA.Files.filter(fileA => fileB.Path === fileA.Path).length > 0;
						if (!found) {
							deletedFiles.push(fileB);
						}
					});
					const modifiedFiles = [];
					manifestB.Files.forEach(fileB => {
						const fileA = manifestA.Files.find(fileA => fileB.Path === fileA.Path && fileB.Object.Hash !== fileA.Object.Hash);
						if (fileA) {
							modifiedFiles.push({new: fileA, old: fileB});
						}
					});
					let html = '';
					addedFiles.forEach(file => {
						html += '<tr><td>' + file.Path + '</td><td>Added</td>';
						html += '<td class="size">&dash;</td>';
						html += '<td class="size tt" title="' + file.Object.Size + ' byte">' + sizeToString(file.Object.Size) + '</td>';
						html += '<td>&dash;</td>';
						html += '<td>' + file.Object.Hash + '</td></tr>';
					});
					deletedFiles.forEach(file => {
						html += '<tr><td>' + file.Path + '</td><td>Deleted</td>';
						html += '<td class="size tt" title="' + file.Object.Size + ' byte">' + sizeToString(file.Object.Size) + '</td>';
						html += '<td class="size">&dash;</td>';
						html += '<td>' + file.Object.Hash + '</td>';
						html += '<td>&dash;</td></tr>';
					});
					modifiedFiles.forEach(files => {
						html += '<tr><td>' + files.old.Path + '</td><td>Modified</td>';
						html += '<td class="size tt" title="' + files.old.Object.Size + ' byte">' + sizeToString(files.old.Object.Size) + '</td>';
						html += '<td class="size tt" title="' + files.new.Object.Size + ' byte">' + sizeToString(files.new.Object.Size) + '</td>';
						html += '<td>' + files.old.Object.Hash + '</td>';
						html += '<td>' + files.new.Object.Hash + '</td></tr>';
					});
					document.getElementById('compareTable').innerHTML = html;
					document.getElementById('compare').style.display = 'block';
				} else {
					renderError('Unable to find package versions for comparison.');
				}
			}

			function updateFromHash() {
				const hash = location.hash;

				const matchCompare = hash.match(/^#\/compare\/([a-z0-9_-]+)\/([0-9]+)\/([0-9]+)$/);
				if (matchCompare) {
					loadCompare(matchCompare[1], matchCompare[2], matchCompare[3]);
					return;
				}

				const matchPackageAndVersion = hash.match(/^#\/([a-z0-9_-]+)\/([0-9]+)$/);
				if (matchPackageAndVersion) {
					loadPackage(matchPackageAndVersion[1], matchPackageAndVersion[2]);
					return;
				}

				const matchPackage = hash.match(/^#\/([a-z0-9_-]+)$/);
				if (matchPackage) {
					loadVersions(matchPackage[1]);
					return;
				}

				if (hash === '#/') {
					hideAll();
					loadPackages();
					return;
				}

				location.hash = '#/';
			}

			window.onhashchange = updateFromHash;
			updateFromHash();
		</script>
	</body>
</html>
`)
