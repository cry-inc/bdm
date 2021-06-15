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
		<div class="usermenu">
			<span class="guest" v-if="!user">
				Guest
			</span>
			<router-link v-if="user" v-bind:to="'/users/' + user.Id">
				{{user.Id}}
			</router-link>
			<span v-if="user && user.Admin">
				| <router-link to="/users">Manage Users</router-link>
			</span>
			<span v-if="!user">
				| <router-link to="/login">Login</router-link>
			</span>
			<button v-if="user" @click="logout">
				Logout
			</button>
		</div>`
}
