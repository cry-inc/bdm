import Helper from '../helper.js';

export default {
	props: ['package', 'version'],
	data() {
		return {
			loaded: false,
			manifest: null,
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
			<div class="alert alert-danger" role="alert" v-if="!manifest">
				The package {{package}} in version {{version}} does not exist!
			</div>
			<div v-if="manifest">
				<table class="table table-sm">
					<tbody>	
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
					</tbody>
				</table>
				<p>
					<a v-bind:href="'zip/' + package + '/' + version">Download Package as ZIP</a><br>
					<a target="_blank" rel="noopener" v-bind:href="'manifests/' + package + '/' + version">Package Manifest JSON</a><br>
					<router-link v-if="version > 1" v-bind:to="'/' + package + '/' + version + '/compare/' + (version - 1)">
						Compare with Previous Version
					</router-link>
				</p>
				<table class="table table-striped table-sm">
					<thead>	
						<tr>
							<th>File</th>
							<th class="size">Size</th>
							<th>Hash</th>
						</tr>
					</thead>
					<tbody>
						<tr v-for="file in manifest.Files">
							<td><a v-bind:href="'files/' + package + '/' + version + '/' + file.Object.Hash + '/' + file.Name">{{file.Path}}</a></td>
							<td>{{$filters.size(file.Object.Size)}}</td>
							<td>{{file.Object.Hash}}</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>`
}
