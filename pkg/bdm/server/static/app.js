import Packages from './components/packages.js'
import Versions from './components/versions.js'
import Package from './components/package.js'
import Compare from './components/compare.js'
import Breadcrumbs from './components/breadcrumbs.js'
import UserMenu from './components/user-menu.js'
import Helper from './helper.js'

const router = VueRouter.createRouter({
	history: VueRouter.createWebHashHistory(),
	routes: [
		{path: '/', name: 'packages', component: Packages},
		{path: '/:package', name: 'versions', component: Versions, props: true},
		{path: '/:package/:version', name: 'package', component: Package, props: true},
		{path: '/:package/:version/compare/:versionOther', name: 'compare', component: Compare, props: true},
	]
});

const app = Vue.createApp({});
app.use(router);

app.component('breadcrumbs', Breadcrumbs);
app.component('user-menu', UserMenu);

app.config.globalProperties.$filters = {
	size(bytes) {
		return Helper.getSizeString(bytes);
	},
	date(unixTime) {
		return new Date(unixTime * 1000).toLocaleString();
	}
};

app.mount('#bdm');
