export default {
	data() {
		return {
			packages: []
		};
	},
	async created() {
		const response = await fetch('manifests');
		const packages = await response.json();
		this.packages = packages;
	},
	template: `
		<div>
			<h1>Packages</h1>
			<ul>
				<li v-for="package in packages">
					<router-link v-bind:to="'/' + package.Name">{{package.Name}}</router-link>
				</li>
			</ul>
		</div>`
}
