export default {
	props: ['package', 'versionA', 'versionB'],
	data() {
		return {
			addedFiles: [],
			deletedFiles: [],
			modifiedFiles: []
		};
	},
	async created() {
		const responseA = await fetch('manifests/' + this.package + '/' + this.versionA);
		const responseB = await fetch('manifests/' + this.package + '/' + this.versionB);
		this.manifestA = await responseA.json();
		this.manifestB = await responseB.json();

		this.manifestA.Files.forEach(fileA => {
			const found = this.manifestB.Files.filter(
				fileB => fileB.Path === fileA.Path).length > 0;
			if (!found) {
				this.addedFiles.push(fileA);
			}
		});
		this.manifestB.Files.forEach(fileB => {
			const found = this.manifestA.Files.filter(
				fileA => fileB.Path === fileA.Path).length > 0;
			if (!found) {
				this.deletedFiles.push(fileB);
			}
		});
		this.manifestB.Files.forEach(fileB => {
			const fileA = this.manifestA.Files.find(
				fileA => fileB.Path === fileA.Path &&
				fileB.Object.Hash !== fileA.Object.Hash);
			if (fileA) {
				this.modifiedFiles.push({new: fileA, old: fileB});
			}
		});
	},
	template: `
		<div>
			<h1>Compare Package {{package}} Version {{versionA}} and {{versionB}}</h1>
			<table>
				<tr>
					<th>File</th>
					<th>Status</th>
					<th>Old Size</th>
					<th>New Size</th>
					<th>Old Hash</th>
					<th>New Hash</th>
				</tr>
				<tr v-for="file in addedFiles">
					<td>{{file.Path}}</td>
					<td>Added</td>
					<td>&dash;</td>
					<td title="{{file.Object.Size}} byte">{{$filters.size(file.Object.Size)}}</td>
					<td>&dash;</td>
					<td>{{file.Object.Hash}}</td>
				</tr>
				<tr v-for="file in deletedFiles">
					<td>{{file.Path}}</td>
					<td>Deleted</td>
					<td title="{{file.Object.Size}} byte">{{$filters.size(file.Object.Size)}}</td>
					<td>&dash;</td>
					<td>{{file.Object.Hash}}</td>
					<td>&dash;</td>
				</tr>
				<tr v-for="file in modifiedFiles">
					<td>{{file.old.Path}}</td>
					<td>Modified</td>
					<td title="{{file.old.Object.Size}} byte">{{$filters.size(file.old.Object.Size)}}</td>
					<td title="{{file.new.Object.Size}} byte">{{$filters.size(file.new.Object.Size)}}</td>
					<td>{{file.old.Object.Hash}}</td>
					<td>{{file.new.Object.Hash}}</td>
				</tr>
			</table>
		</div>`
}
