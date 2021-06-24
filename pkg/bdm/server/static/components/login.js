export default {
	data() {
		return {
			userId: '',
			password: ''
		};
	},
	methods: {
		async login() {
			const request = {
				UserId: this.userId,
				Password: this.password
			};
			const response = await fetch('/login', {
				method: 'POST',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				this.password = '';
				alert('Login failed!');
			} else {
				this.userId = '';
				this.password = '';
				// Navigato to home URL and reload
				await this.$router.push('/');
				await this.$router.go();
			}
		}
	},
	template: `
	<div>
		<h1>Login</h1>
		<div class="mb-3">
			<label for="userId" class="form-label">User ID address</label>
			<input v-model="userId" id="userId" placeholder="User ID" class="form-control" id="userId" aria-describedby="userIdHelp">
			<div id="userIdHelp" class="form-text">Your User ID is typically your E-Mail address.</div>
		</div>
		<div class="mb-3">
			<label for="password" class="form-label">Password</label>
			<input type="password" v-model="password" id="password" placeholder="Password" class="form-control">
		</div>
		<button @click="login" class="btn btn-primary">Login</button>
	</div>`
}
