export default {
	props: ['userId'],
	data() {
		return {
			login: null,
			user: null,
			loaded: false,
			oldPassword: '',
			newPassword1: '',
			newPassword2: ''
		};
	},
	async created() {
		const loginResponse = await fetch('login');
		this.login = loginResponse.ok ? await loginResponse.json() : null;
		const userResponse = await fetch('users/' + this.userId);
		this.user = userResponse.ok ? await userResponse.json() : null;
		this.loaded = true;
	},
	methods: {
		async changePassword() {
			if (this.newPassword1.length < 8) {
				alert("The new password is not long enough!");
				return;
			}
			if (this.newPassword1 !== this.newPassword2) {
				alert("The new passwords do not match!");
				return;
			}
			const request = {
				OldPassword: this.oldPassword,
				NewPassword: this.newPassword1
			};
			const response = await fetch('/users/' + this.userId + '/password', {
				method: 'PATCH',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				alert('Failed to change password!');
			} else {
				this.oldPassword = '';
				this.newPassword1 = '';
				this.newPassword2 = '';
			}
		}
	},
	template: `
		<div>
			<h1>{{userId}}</h1>
			<div class="alert alert-danger" role="alert" v-if="loaded && !user">
				The user {{userId}} does not exist!
			</div>
			<div v-if="loaded && user">
				<table class="table table-sm">
					<tbody>	
						<tr>
							<th style="width: 20%">User ID</th>
							<td style="width: 80%">{{user.Id}}</td>
						</tr>
						<tr>
							<th>Reader</th>
							<td>{{user.Reader ? 'Yes' : 'No'}}</td>
						</tr>
						<tr>
							<th>Writer</th>
							<td>{{user.Writer ? 'Yes' : 'No'}}</td>
						</tr>
						<tr>
							<th>Admin</th>
							<td>{{user.Admin ? 'Yes' : 'No'}}</td>
						</tr>
					</tbody>
				</table>
				<p><router-link v-bind:to="'/users/' + user.Id + '/tokens'">Manage Tokens</router-link></p>
				
				<h2 class="mt-4">Change Password</h2>
				<div class="mb-3" v-if="!login || !login.Admin || login.Id === user.Id">
					<label for="oldPw" class="form-label">Old Password</label>
					<input type="password" v-model="oldPassword" class="form-control" id="oldPw" placeholder="Old Password">
				</div>
				<div class="mb-3">
					<label for="newPw1" class="form-label">New Password (at least 8 characters)</label>
					<input type="password" v-model="newPassword1" class="form-control" id="newPw1" placeholder="New Password">
				</div>
				<div class="mb-3">
					<label for="newPw2" class="form-label">Repeat New Password</label>
					<input type="password" v-model="newPassword2" class="form-control" id="newPw2" placeholder="New Password">
				</div>
				<button class="btn btn-primary" @click="changePassword">Change Password</button>
			</div>
		</div>`
}
