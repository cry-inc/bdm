export default {
	data() {
		return {
			breadcrumbs: []
		};
	},
	watch: {
		'$route'(route) {
			this.breadcrumbs = [];
			if (route.name == 'packages' || route.name == 'versions' || route.name == 'package' || route.name == 'compare') {
				this.breadcrumbs.push({
					Name: 'Packages',
					Route: '/'
				});
			}
			if (route.name == 'versions' || route.name == 'package' || route.name == 'compare') {
				this.breadcrumbs.push({
					Name: route.params.package,
					Route: '/' + route.params.package
				});
			}
			if (route.name == 'package' || route.name == 'compare') {
				this.breadcrumbs.push({
					Name: 'Version ' + route.params.version,
					Route: '/' + route.params.package + '/' + route.params.version
				});
			}
			if (route.name == 'compare') {
				this.breadcrumbs.push({
					Name: 'Compare with Version ' + route.params.versionOther,
					Route: '/' + route.params.package + '/' + route.params.version + '/compare/' + route.params.versionOther
				});
			}
		}
	},
	template: `
		<div>
			BDM
			<span v-for="(breadcrumb, index) in breadcrumbs">
				<span v-if="index !== 0"> / </span>
				<router-link  v-bind:to="breadcrumb.Route">{{breadcrumb.Name}}</router-link>
			</span>
		</div>`
}
