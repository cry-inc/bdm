import Helper from '../helper.js';

export default {
	props: ['package', 'version'],
	data() {
		return {
			loaded: false,
			manifest: {},
			size: null,
			published: null
		};
	},
	async created() {
		const response = await fetch('manifests/' + this.package + '/' + this.version);
		if (response.ok) {
			this.manifest = await response.json();
			this.size = Helper.getPackageSize(this.manifest);
			Helper.addFileNames(this.manifest);
		}
		this.loaded = true;
	},
	template: `
		<div v-if="loaded">
			<h1>{{package}} Version {{version}}</h1>
			<div class="error" v-if="!manifest.Files">
				The {{package}} in version {{version}} does not exist!
			</div>
			<div>
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
					<a v-bind:href="'manifests/' + package + '/' + version">Package Manifest JSON</a><br>
					<router-link v-if="version > 1" v-bind:to="'/' + package + '/' + version + '/compare/' + (version - 1)">
						Compare with Previous Version
					</router-link>
				</p>
				<table>
					<tr>
						<th>File</th>
						<th class="size">Size</th>
						<th>Hash</th>
					</tr>
					<tr v-for="file in manifest.Files">
						<td><a v-bind:href="'files/' + package + '/' + version + '/' + file.Object.Hash + '/' + file.Name">{{file.Path}}</a></td>
						<td>{{$filters.size(file.Object.Size)}}</td>
						<td>{{file.Object.Hash}}</td>
					</tr>
				</table>
			</div>
		</div>`
}
