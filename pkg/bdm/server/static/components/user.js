export default {
	props: ['userId'],
	data() {
		return {
			user: null,
			loaded: false,
			oldPassword: '',
			newPassword1: '',
			newPassword2: ''
		};
	},
	async created() {
		const response = await fetch('users/' + this.userId);
		this.user = response.ok ? await response.json() : null;
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
			<div class="error" v-if="loaded && !user">
				The user {{userId}} does not exist!
			</div>
			<div v-if="loaded && user">
				ID: {{user.Id}}<br>
				Reader: {{user.Reader ? 'yes' : 'no'}}<br>
				Writer: {{user.Writer ? 'yes' : 'no'}}<br>
				Admin: {{user.Admin ? 'yes' : 'no'}}<br>
				<p><router-link v-bind:to="'/users/' + user.Id + '/tokens'">Manage Tokens</router-link></p>
				<h2>Change Password</h2>
				Old Password: <input v-model="oldPassword" type="password" placeholder="Old Password"/> (not required for Admins when changing passwords for others)<br>
				New Password: <input v-model="newPassword1" type="password" placeholder="New Password"/> (at least 8 characters)<br>
				Repeat New Password: <input v-model="newPassword2" type="password" placeholder="New Password"/><br>
				<button @click="changePassword">Change Password</button>
			</div>
		</div>`
}
