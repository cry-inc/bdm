import Packages from './components/packages.js'
import Versions from './components/versions.js'

const router = VueRouter.createRouter({
	history: VueRouter.createWebHashHistory(),
	routes: [
		{path: '/', component: Packages},
		{path: '/:package', component: Versions, props: true},
	]
});

Vue.createApp({}).use(router).mount('#bdm');
