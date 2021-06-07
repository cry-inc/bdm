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
		User ID: <input v-model="userId" placeholder="User ID"></input><br>
		Password: <input v-model="password" type="password" placeholder="Password"></input><br>
		<button @click="login">Login</button>
	</div>`
}
