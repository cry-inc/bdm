export default {
	props: ['userId'],
	data() {
		return {
			user: null,
			loaded: false,
			newPassword: ''
		};
	},
	async created() {
		const response = await fetch('users/' + this.userId);
		this.user = response.ok ? await response.json() : null;
		this.loaded = true;
	},
	methods: {
		async changePassword() {
			const request = {
				Password: this.newPassword
			};
			const response = await fetch('/users/' + this.userId + '/password', {
				method: 'PATCH',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				alert('Failed to change password!');
			} else {
				this.newPassword = '';
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
				<h2>Change Password</h2>
				New Password: <input v-model="newPassword" type="password" placeholder="New Password"/><br>
				<button @click="changePassword">Change Password</button>
				<tokens v-bind:userId="user.Id"></tokens>
			</div>
		</div>`
}
