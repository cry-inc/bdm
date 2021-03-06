export default {
	data() {
		return {
			breadcrumbs: []
		};
	},
	watch: {
		'$route'(route) {
			this.breadcrumbs = [];
			if (route.name === 'packages' || route.name === 'versions' || route.name === 'package' || route.name === 'compare') {
				this.breadcrumbs.push({
					Name: 'Packages',
					Route: '/'
				});
			}
			if (route.name === 'versions' || route.name === 'package' || route.name === 'compare') {
				this.breadcrumbs.push({
					Name: route.params.package,
					Route: '/' + route.params.package
				});
			}
			if (route.name === 'package' || route.name === 'compare') {
				this.breadcrumbs.push({
					Name: 'Version ' + route.params.version,
					Route: '/' + route.params.package + '/' + route.params.version
				});
			}
			if (route.name === 'compare') {
				this.breadcrumbs.push({
					Name: 'Compare with Version ' + route.params.versionOther,
					Route: '/' + route.params.package + '/' + route.params.version + '/compare/' + route.params.versionOther
				});
			}
			if (route.name === 'users' || route.name === 'user' || route.name === 'tokens') {
				this.breadcrumbs.push({
					Name: 'Users',
					Route: '/users'
				});
			}
			if (route.name === 'user' || route.name === 'tokens') {
				this.breadcrumbs.push({
					Name: route.params.userId,
					Route: '/users/' + route.params.userId
				});
			}
			if (route.name === 'tokens') {
				this.breadcrumbs.push({
					Name: 'Tokens',
					Route: '/users/' + route.params.userId + '/tokens'
				});
			}
			if (route.name === 'login') {
				this.breadcrumbs.push({
					Name: 'Login',
					Route: '/login'
				});
			}
		}
	},
	template: `
		<div class="breadcrumbs">
			<span v-for="(breadcrumb, index) in breadcrumbs">
				<span v-if="index !== 0"> / </span>
				<router-link  v-bind:to="breadcrumb.Route">{{breadcrumb.Name}}</router-link>
			</span>
		</div>`
}
