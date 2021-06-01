export default {
	data() {
		return {
			user: null,
			usersEnabled: false,
			showLoginForm: false,
			userId: '',
			password: ''
		};
	},
	async created() {
		return this.query();
	},
	methods: {
		async query() {
			const response = await fetch('login');
			this.usersEnabled = response.status !== 503;
			this.user = response.ok ? await response.json() : null;
		},
		async toggleLogin() {
			this.showLoginForm = !this.showLoginForm;
		},
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
				this.showLoginForm = false;
				await this.query();
			}
		},
		async logout() {
			await fetch('/login', {method: 'DELETE'});
			await this.query();
			this.$router.go();
		}
	},
	template: `
		<div class="usermenu">
			<span class="guest" v-if="!usersEnabled || !user">
				Guest
			</span>
			<router-link v-if="user" v-bind:to="'/users/' + user.Id">{{user.Id}}</router-link>
			<span v-if="user && user.Admin">
				| <router-link to="/users">Manage Users</router-link>
			</span>
			<button v-if="usersEnabled && !user" @click="toggleLogin">
				Login
			</button>
			<button v-if="user" @click="logout">
				Logout
			</button>
			<div class="loginform" v-if="showLoginForm">
				User ID: <input v-model="userId" placeholder="User ID"></input><br>
				Password: <input v-model="password" type="password" placeholder="Password"></input><br>
				<button @click="login">Login</button>
			</div>
		</div>`
}
