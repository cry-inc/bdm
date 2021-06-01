export default {
	data() {
		return {
			users: [],
			loaded: false
		};
	},
	async created() {
		const response = await fetch('users');
		this.users = response.ok ? await response.json() : [];
		this.loaded = true;
	},
	template: `
		<div v-if="loaded">
			<h1>Users</h1>
			<div class="error" v-if="users.length === 0">
				No users found!
			</div>
			<ul>
				<li v-for="user in users">
					<router-link v-bind:to="'/users/' + user.Id">{{user.Id}}</router-link>
				</li>
			</ul>
		</div>`
}
