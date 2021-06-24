export default {
	data() {
		return {
			user: null
		};
	},
	async created() {
		const response = await fetch('login');
		this.user = response.ok ? await response.json() : null;
	},
	methods: {
		async logout() {
			if (confirm('Log out?')) {
				await fetch('/login', {method: 'DELETE'});
				// Go to packages and reload
				await this.$router.push('/');
				await this.$router.go();
			}
		}
	},
	template: `
		<div>
			<router-link v-if="user" v-bind:to="'/users/' + user.Id">
				My Profile
			</router-link>
			<span v-if="user && user.Admin">
				| <router-link to="/users">Manage Users</router-link>
			</span>
			<button class="ms-2 btn btn-sm btn-secondary" v-if="user" @click="logout">
				Logout
			</button>
			<router-link v-if="!user" class="btn btn-sm btn-secondary" to="/login">Login</router-link>
		</div>`
}
