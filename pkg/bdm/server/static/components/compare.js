export default {
	props: ['package', 'version', 'versionOther'],
	data() {
		return {
			loaded: false,
			manifest: undefined,
			manifestOther: undefined,
			addedFiles: [],
			deletedFiles: [],
			modifiedFiles: []
		};
	},
	async created() {
		const response = await fetch('manifests/' + this.package + '/' + this.version);
		const responseOther = await fetch('manifests/' + this.package + '/' + this.versionOther);
		this.manifest = response.ok ? await response.json() : null;
		this.manifestOther = responseOther.ok ? await responseOther.json() : null;
		if (this.manifest && this.manifestOther) {
			this.manifest.Files.forEach(file => {
				const found = this.manifestOther.Files.filter(
					fileOther => fileOther.Path === file.Path).length > 0;
				if (!found) {
					this.addedFiles.push(file);
				}
			});
			this.manifestOther.Files.forEach(fileOther => {
				const found = this.manifest.Files.filter(
					file => fileOther.Path === file.Path).length > 0;
				if (!found) {
					this.deletedFiles.push(fileOther);
				}
			});
			this.manifestOther.Files.forEach(fileOther => {
				const file = this.manifest.Files.find(
					file => fileOther.Path === file.Path &&
					fileOther.Object.Hash !== file.Object.Hash);
				if (file) {
					this.modifiedFiles.push({new: file, old: fileOther});
				}
			});
		}
		this.loaded = true;
	},
	template: `
		<div v-if="loaded">
			<h1>Compare Package {{package}} Version {{version}} and {{versionOther}}</h1>
			<div class="alert alert-danger" role="alert" v-if="manifest === null">
				The package {{package}} in version {{version}} was not found!
			</div>
			<div class="alert alert-danger" role="alert" v-if="manifestOther === null">
				The package {{package}} in version {{versionOther}} was not found!
			</div>
			<table class="table table-sm table-striped" v-if="manifest && manifestOther">
				<thead>
					<tr>
						<th>File</th>
						<th>Status</th>
						<th>Old Size</th>
						<th>New Size</th>
						<th>Old Hash</th>
						<th>New Hash</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="file in addedFiles">
						<td>{{file.Path}}</td>
						<td>Added</td>
						<td>&dash;</td>
						<td v-bind:title="file.Object.Size + ' byte'">{{$filters.size(file.Object.Size)}}</td>
						<td>&dash;</td>
						<td>{{file.Object.Hash}}</td>
					</tr>
					<tr v-for="file in deletedFiles">
						<td>{{file.Path}}</td>
						<td>Deleted</td>
						<td v-bind:title="file.Object.Size + ' byte'">{{$filters.size(file.Object.Size)}}</td>
						<td>&dash;</td>
						<td>{{file.Object.Hash}}</td>
						<td>&dash;</td>
					</tr>
					<tr v-for="file in modifiedFiles">
						<td>{{file.old.Path}}</td>
						<td>Modified</td>
						<td v-bind:title="file.old.Object.Size + ' byte'">{{$filters.size(file.old.Object.Size)}}</td>
						<td v-bind:title="file.new.Object.Size + ' byte'">{{$filters.size(file.new.Object.Size)}}</td>
						<td>{{file.old.Object.Hash}}</td>
						<td>{{file.new.Object.Hash}}</td>
					</tr>
				</tbody>
			</table>
		</div>`
}
