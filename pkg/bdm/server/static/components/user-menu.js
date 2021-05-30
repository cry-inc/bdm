export default {
	data() {
		return {
			user: null,
			usersEnabled: false,
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
		}
	},
	template: `
		<div class="usermenu">
			<div v-if="usersEnabled === false || !user">
				Guest
			</div>
			<div v-if="user">
				{{user.Id}}
			</div>
		</div>`
}
