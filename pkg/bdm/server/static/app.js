import Packages from './components/packages.js'
import Versions from './components/versions.js'
import Package from './components/package.js'
import PackageDiff from './components/packagediff.js'
import Helper from './helper.js'

const router = VueRouter.createRouter({
	history: VueRouter.createWebHashHistory(),
	routes: [
		{path: '/', component: Packages},
		{path: '/:package', component: Versions, props: true},
		{path: '/:package/:version', component: Package, props: true},
		{path: '/:package/:versionA/diff/:versionB', component: PackageDiff, props: true},
	]
});

const app = Vue.createApp({});
app.use(router);
app.config.globalProperties.$filters = {
	size(bytes) {
		return Helper.getSizeString(bytes);
	},
	date(unixTime) {
		return new Date(unixTime * 1000).toLocaleString();
	}
};
app.mount('#bdm');
