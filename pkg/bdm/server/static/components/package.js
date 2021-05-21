import * as Helper from '../helper.js';

export default {
	props: ['package', 'version'],
	data() {
		return {
			manifest: {},
			size: null,
			published: null
		};
	},
	async created() {
		const response = await fetch('manifests/' + this.package + '/' + this.version);
		this.manifest = await response.json();
		this.size = Helper.getPackageSize(this.manifest);
		Helper.addFileNames(this.manifest);
	},
	template: `
		<div>
			<h1>{{package}} Version {{version}}</h1>
			<table>
				<tr>
					<th>Name</th>
					<td>{{manifest.PackageName}}</td>
				</tr>
				<tr>
					<th>Version</th>
					<td>{{manifest.PackageVersion}}</td>
				</tr>
				<tr>
					<th>Files</th>
					<td>{{manifest.Files ? manifest.Files.length : 0}}</td>
				</tr>
				<tr>
					<th>Size</th>
					<td>{{$filters.size(size)}}</td>
				</tr>
				<tr>
					<th>Published</th>
					<td>{{$filters.date(manifest.Published)}}</td>
				</tr>
				<tr>
					<th>Hash</th>
					<td>{{manifest.Hash}}</td>
				</tr>
			</table>
			<p>
				<a v-bind:href="'zip/' + package + '/' + version">Download Package as ZIP</a><br>
				<a v-bind:href="'manifests/' + package + '/' + version">Package Manifest JSON</a>
			</p>
			<table>
				<tr>
					<th>File</th>
					<th class="size">Size</th>
					<th>Hash</th>
				</tr>
				<tr v-for="file in manifest.Files">
					<td><a v-bind:href="'files/' + package + '/' + version + '/' + file.Object.Hash + '/' + file.Name">{{file.Path}}</a></td>
					<td>{{file.Object.Size}}</td>
					<td>{{file.Object.Hash}}</td>
				</tr>
			</table>
		</div>`
}
