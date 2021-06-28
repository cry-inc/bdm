export default {
	data() {
		return {
			users: [],
			currentUser: {},
			loaded: false,
			newUserId: '',
			newUserPassword: ''
		};
	},
	async created() {
		await this.query();
	},
	methods: {
		async query() {
			const usersResponse = await fetch('users');
			this.users = usersResponse.ok ? await usersResponse.json() : [];
			const currentUserResponse = await fetch('login');
			this.currentUser = currentUserResponse.ok ? await currentUserResponse.json() : {};
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
		},
		async createUser() {
			if (this.newUserPassword.length < 8) {
				alert('Password must have at least 8 characters!');
				return;
			}
			const request = {
				Id: this.newUserId,
				Password: this.newUserPassword
			};
			const response = await fetch('/users', {
				method: 'POST',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				alert('Failed to create user!');
			} else {
				this.newUserId = '';
				this.newUserPassword = '';
			}
			await this.query();
		}
	},
	computed: {
		sortedUsers() {
			// Return sorted copy of original array
			return this.users.concat().sort(function(a, b) {
				if (a.Id > b.Id) {
					return 1;
				} else if (a.Id < b.Id) {
					return -1;
				} else {
					return 0;
				}
			});
		}
	},
	template: `
		<div v-if="loaded">
			<h1>Users</h1>
			<div class="alert alert-warning" role="alert" v-if="users.length === 0">
				No users found!
			</div>
			<table class="table table-sm table-striped" v-if="users.length > 0">
				<thead>
					<tr>
						<th>User</th>
						<th>Reader</th>
						<th>Writer</th>
						<th>Admin</th>
						<th>&nbsp;</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="user in sortedUsers">
						<td>
							<router-link v-bind:to="'/users/' + user.Id">{{user.Id}}</router-link>
						</td>
						<td><input v-bind:disabled="currentUser.Id === user.Id" class="form-check-input" type="checkbox" @click="changeRole(user, 'Reader')" v-model="user.Reader"></td>
						<td><input v-bind:disabled="currentUser.Id === user.Id" class="form-check-input" type="checkbox" @click="changeRole(user, 'Writer')" v-model="user.Writer"></td>
						<td><input v-bind:disabled="currentUser.Id === user.Id" class="form-check-input" type="checkbox" @click="changeRole(user, 'Admin')" v-model="user.Admin"></td>
						<td><button class="btn btn-sm btn-danger" v-bind:disabled="currentUser.Id === user.Id" @click="deleteUser(user)">Delete</button></td>
					</tr>
				</tbody>
			</table>
			<h2>Create New User</h2>
			<div class="mb-3">
				<label for="userId" class="form-label">User ID</label>
				<input type="text" v-model="newUserId" class="form-control" id="userId" placeholder="User ID">
			</div>
			<div class="mb-3">
				<label for="newPw" class="form-label">Initial Password (at least 8 characters)</label>
				<input type="password" v-model="newUserPassword" class="form-control" id="newPw" placeholder="Initial Password">
			</div>
			<button class="btn btn-primary" @click="createUser">Create User</button>
		</div>`
}
