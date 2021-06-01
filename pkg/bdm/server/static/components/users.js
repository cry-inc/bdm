export default {
	data() {
		return {
			users: [],
			loaded: false
		};
	},
	async created() {
		await this.query();
	},
	methods: {
		async query() {
			const response = await fetch('users');
			this.users = response.ok ? await response.json() : [];
			this.loaded = true;
		},
		async deleteUser(user) {
			const confirmed = confirm('Really delete user ' + user.Id + '?');
			if (!confirmed) {
				return;
			}
			const response = await fetch('/users/' + user.Id, {method: 'DELETE'});
			if (!response.ok) {
				alert('Unable to delete user!');
			}
			await this.query();
		},
		async changeRole(user, role) {
			const request = {
				Reader: user.Reader,
				Writer: user.Writer,
				Admin: user.Admin
			};
			request[role] = !request[role];
			const response = await fetch('/users/' + user.Id + '/roles', {
				method: 'PATCH',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				alert('Failed to change role!');
			}
			await this.query();
		}
	},
	template: `
		<div v-if="loaded">
			<h1>Users</h1>
			<div class="error" v-if="users.length === 0">
				No users found!
			</div>
			<table v-if="users.length > 0">
				<tr>
					<th>User</th>
					<th>Reader</th>
					<th>Writer</th>
					<th>Admin</th>
					<th>&nbsp;</th>
				</tr>
				<tr v-for="user in users">
					<td>
						<router-link v-bind:to="'/users/' + user.Id">{{user.Id}}</router-link>
					</td>
					<td><input type="checkbox" @click="changeRole(user, 'Reader')" v-model="user.Reader"></td>
					<td><input type="checkbox" @click="changeRole(user, 'Writer')" v-model="user.Writer"></td>
					<td><input type="checkbox" @click="changeRole(user, 'Admin')" v-model="user.Admin"></td>
					<td><button @click="deleteUser(user)">Delete</button></td>
				</tr>
			</table>
		</div>`
}
