export default {
	props: ['package'],
	data() {
		return {
			versions: []
		};
	},
	async created() {
		const response = await fetch('manifests/' + this.package);
		const versions = await response.json();
		this.versions = versions;
	},
	template: `
		<div>
			<h1>{{package}} Versions</h1>
			<ul>
				<li v-for="version in versions">
					<router-link v-bind:to="'/' + package + '/' + version.Version">Version {{version.Version}}</router-link>
				</li>
			</ul>
		</div>`
}
