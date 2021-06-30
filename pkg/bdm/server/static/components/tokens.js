export default {
	props: ['userId'],
	data() {
		const d = new Date();
		d.setDate(d.getDate() + 1);
		const minDate = d.toISOString().split('T')[0];
		d.setFullYear(d.getFullYear() + 1);
		const maxDate = d.toISOString().split('T')[0];
		return {
			tokens: [],
			loaded: false,
			createTokenName: '',
			createTokenExpiration: maxDate,
			createTokenReader: true,
			createTokenWriter: false,
			minDate: minDate,
			maxDate: maxDate,
			secretToken: null
		};
	},
	async created() {
		await this.query();
	},
	methods: {
		async query() {
			const response = await fetch('users/' + this.userId + '/tokens');
			this.tokens = response.ok ? await response.json() : [];
			this.tokens.forEach(t => t.Expiration = new Date(t.Expiration).toISOString().split('T')[0]);
			this.loaded = true;
		},
		async deleteToken(token) {
			const confirmed = confirm('Really delete the token ' + token.Name + '?');
			if (!confirmed) {
				return;
			}
			const response = await fetch('/users/' + this.userId + '/tokens/' + token.Id, {method: 'DELETE'});
			if (!response.ok) {
				alert('Unable to delete token!');
			}
			await this.query();
		},
		confirmSecret() {
			this.secretToken = null;
		},
		async createToken() {
			if (this.createTokenName.length < 1 || this.createTokenName.length > 255) {
				alert('Invalid token name!');
				return;
			}
			const request = {
				Name: this.createTokenName,
				Expiration: new Date(this.createTokenExpiration),
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
				this.createTokenName = '';
				this.createTokenExpiration = this.maxDate;
				this.createTokenReader = true;
				this.createTokenWriter = false;
			}
			const newToken = await response.json();
			this.secretToken = newToken.Secret;
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
						<th>Expiration</th>
						<th>Reader</th>
						<th>Writer</th>
						<th>&nbsp;</th>
					</tr>
				</thead>
				<tbody>
					<tr v-for="token in tokens">
						<td>{{token.Name}}</td>
						<td>{{token.Expiration}}</td>
						<td><input class="form-check-input" disabled="disabled" type="checkbox" v-model="token.Reader"></td>
						<td><input class="form-check-input" disabled="disabled" type="checkbox" v-model="token.Writer"></td>
						<td><button class="btn btn-sm btn-danger" @click="deleteToken(token)">Delete</button></td>
					</tr>
				</tbody>
			</table>
			<div v-if="secretToken">
				<h2 class="mt-4">New Token</h2>
				<p>This is the new secret token value. Make sure to store it somewhere safe. You will not be able to look it up after confirmation!</p>
				<div class="mb-3">
					<input type="text" disabled v-model="secretToken" class="form-control" id="secretToken">
				</div>
				<button class="btn btn-primary" @click="confirmSecret">Confirm</button>
			</div>
			<h2 class="mt-4">Create New Token</h2>
			<div class="mb-3">
				<label for="tokenName" class="form-label">Token Name</label>
				<input type="text" v-model="createTokenName" class="form-control" id="tokenName" placeholder="Token Name">
			</div>
			<div class="mb-3">
				<label for="tokenExpiration" class="form-label">Token Expiration Date</label>
				<input type="date" v-bind:min="minDate" v-bind:max="maxDate" v-model="createTokenExpiration" class="form-control" id="tokenExpiration">
			</div>
			<div class="form-check">
				<input v-model="createTokenReader" class="form-check-input" type="checkbox" id="reader">
				<label class="form-check-label" for="reader">
					Read Permission
				</label>
			</div>
			<div class="form-check">
				<input v-model="createTokenWriter" class="form-check-input" type="checkbox" id="writer">
				<label class="form-check-label" for="writer">
					Write Permission
				</label>
			</div>
			<button class="mt-2 btn btn-primary" @click="createToken">Create Token</button>
		</div>`
}
