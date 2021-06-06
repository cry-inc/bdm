export default {
	props: ['userId'],
	data() {
		return {
			tokens: [],
			loaded: false
		};
	},
	async created() {
		await this.query();
	},
	methods: {
		async query() {
			const response = await fetch('users/' + this.userId + '/tokens');
			this.tokens = response.ok ? await response.json() : [];
			this.loaded = true;
		},
		async deleteToken(token) {
			const confirmed = confirm('Really delete the token ' + token.Id + '?');
			if (!confirmed) {
				return;
			}
			const response = await fetch('/users/' + this.userId + '/tokens/' + token.Id, {method: 'DELETE'});
			if (!response.ok) {
				alert('Unable to delete token!');
			}
			await this.query();
		}
	},
	template: `
		<div v-if="loaded">
			<h1>Tokens</h1>
			<div class="error" v-if="tokens.length === 0">
				No tokens found!
			</div>
			<table v-if="tokens.length > 0">
				<tr>
					<th>Token</th>
					<th>Reader</th>
					<th>Writer</th>
					<th>Admin</th>
					<th>&nbsp;</th>
				</tr>
				<tr v-for="token in tokens">
					<td>{{token.Id}}</td>
					<td><input disabled="disabled" type="checkbox" v-model="token.Reader"></td>
					<td><input disabled="disabled" type="checkbox" v-model="token.Writer"></td>
					<td><input disabled="disabled" type="checkbox" v-model="token.Admin"></td>
					<td><button @click="deleteToken(token)">Delete</button></td>
				</tr>
			</table>
		</div>`
}
