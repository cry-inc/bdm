export default {
	props: ['userId'],
	data() {
		return {
			user: null,
			loaded: false
		};
	},
	async created() {
		const response = await fetch('users/' + this.userId);
		this.user = response.ok ? await response.json() : null;
		this.loaded = true;
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
				<tokens v-bind:userId="user.Id"></tokens>
			</div>
		</div>`
}
