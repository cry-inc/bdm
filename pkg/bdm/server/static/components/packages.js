export default {
	data() {
		return {
			packages: [],
			loaded: false
		};
	},
	async created() {
		const response = await fetch('manifests');
		const packages = await response.json();
		this.packages = packages;
		this.loaded = true;
	},
	template: `
		<div v-if="loaded">
			<h1>Packages</h1>
			<div class="error" v-if="packages.length === 0">
				No packages found!
			</div>
			<ul>
				<li v-for="package in packages">
					<router-link v-bind:to="'/' + package.Name">{{package.Name}}</router-link>
				</li>
			</ul>
		</div>`
}
