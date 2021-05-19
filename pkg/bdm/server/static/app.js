import Packages from './components/packages.js'

const router = VueRouter.createRouter({
	history: VueRouter.createWebHashHistory(),
	routes: [
		{ path: '/', component: Packages },
	]
});

Vue.createApp({}).use(router).mount('#bdm');
