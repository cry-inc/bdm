export default {
	props: ['userId'],
	data() {
		return {
			tokens: [],
			loaded: false,
			createTokenReader: true,
			createTokenWriter: false
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
		},
		async createToken() {
			const request = {
				Reader: this.createTokenReader,
				Writer: this.createTokenWriter
			};
			if (!request.Reader) {
				alert('Token must have reading permission, only writing is optional!');
				return;
			}
			const response = await fetch('/users/' + this.userId + '/tokens', {
				method: 'POST',
				headers: {'Content-Type': 'application/json'},
				body: JSON.stringify(request)
			});
			if (!response.ok) {
				alert('Failed to create token!');
			} else {
				this.createTokenReader = true;
				this.createTokenWriter = false;
			}
			await this.query();
		}
	},
	template: `
		<div v-if="loaded">
			<h1>Tokens</h1>
			<div class="alert alert-warning" role="alert" v-if="tokens.length === 0">
				No tokens found!
			</div>
			<table class="table table-sm table-striped" v-if="tokens.length > 0">
				<thead>
					<tr>
						<th>Token</th>
						<th>Reader</th>
						<th>Writer</th>
						<th>&nbsp;</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="token in tokens">
						<td>{{token.Id}}</td>
						<td><input class="form-check-input" disabled="disabled" type="checkbox" v-model="token.Reader"></td>
						<td><input class="form-check-input" disabled="disabled" type="checkbox" v-model="token.Writer"></td>
						<td><button class="btn btn-sm btn-danger" @click="deleteToken(token)">Delete</button></td>
					</tr>
				</tbody>
			</table>
			<h2 class="mt-4">Create New Token</h2>
			<div class="form-check">
				<input v-model="createTokenReader" class="form-check-input" type="checkbox" id="reader">
				<label class="form-check-label" for="reader">
					Reader
				</label>
			</div>
			<div class="form-check">
				<input v-model="createTokenWriter" class="form-check-input" type="checkbox" id="writer">
				<label class="form-check-label" for="writer">
					Writer
				</label>
			</div>
			<button class="mt-1 btn btn-primary" @click="createToken">Create Token</button>
		</div>`
}
