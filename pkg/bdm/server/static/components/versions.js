export default {
	props: ['package'],
	data() {
		return {
			versions: [],
			loaded: false
		};
	},
	async created() {
		const response = await fetch('manifests/' + this.package);
		this.versions = response.ok ? await response.json() : [];
		this.loaded = true;
	},
	template: `
		<div v-if="loaded">
			<h1>{{package}} Versions</h1>
			<div class="error" v-if="versions.length === 0">
				No versions for package {{package}} found!
			</div>
			<ul>
				<li v-for="version in versions">
					<router-link v-bind:to="'/' + package + '/' + version.Version">Version {{version.Version}}</router-link>
				</li>
			</ul>
		</div>`
}
